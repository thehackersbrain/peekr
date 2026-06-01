// Package collector gathers network connection and interface statistics.
package collector

import (
	"fmt"
	"net"
	"os/user"
	"strconv"
	"strings"
	"time"

	gnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// ---------------------------------------------------------------------------
// Connections
// ---------------------------------------------------------------------------

// ConnRow is one row in the connections table.
type ConnRow struct {
	LocalAddr  string
	RemoteAddr string
	Status     string
	PID        int32
	Process    string
	User       string
	SentBytes  uint64    // cumulative process write bytes (/proc/PID/io)
	RecvBytes  uint64    // cumulative process read bytes
	FirstSeen  time.Time // when this connection was first observed
}

// ConnCache tracks when each connection was first seen across refreshes.
type ConnCache struct {
	seenAt map[string]time.Time
}

// NewConnCache creates an empty connection cache.
func NewConnCache() *ConnCache {
	return &ConnCache{seenAt: make(map[string]time.Time)}
}

// stamp sets FirstSeen on each row and evicts stale entries from the cache.
func (c *ConnCache) stamp(rows []ConnRow) {
	now := time.Now()
	active := make(map[string]struct{}, len(rows))
	for i := range rows {
		key := rows[i].LocalAddr + "|" + rows[i].RemoteAddr + "|" + fmt.Sprint(rows[i].PID)
		active[key] = struct{}{}
		if t, ok := c.seenAt[key]; ok {
			rows[i].FirstSeen = t
		} else {
			c.seenAt[key] = now
			rows[i].FirstSeen = now
		}
	}
	for key := range c.seenAt {
		if _, ok := active[key]; !ok {
			delete(c.seenAt, key)
		}
	}
}

// Connections fetches established and listening connections.
func Connections(cache *ConnCache) (established, listening []ConnRow, err error) {
	conns, err := gnet.Connections("all")
	if err != nil {
		return nil, nil, err
	}

	for _, c := range conns {
		status := c.Status
		switch status {
		case "ESTABLISHED", "LISTEN":
			// keep TCP statuses as-is
		case "", "NONE":
			status = "UDP" // UDP sockets have no TCP state
		default:
			continue // skip TIME_WAIT, CLOSE_WAIT, SYN_*, etc.
		}
		if c.Laddr.Port == 0 {
			continue // skip fully anonymous entries
		}

		laddr := formatAddr(c.Laddr.IP, c.Laddr.Port)
		raddr := formatAddr(c.Raddr.IP, c.Raddr.Port)

		row := ConnRow{
			LocalAddr:  laddr,
			RemoteAddr: raddr,
			Status:     status,
			PID:        c.Pid,
		}

		if c.Pid > 0 {
			row.Process, row.User, row.SentBytes, row.RecvBytes = procInfo(c.Pid)
		} else {
			row.Process = "-"
			row.User = "-"
		}

		// UDP with a remote addr counts as established; all others go to listening.
		if status == "ESTABLISHED" || (status == "UDP" && c.Raddr.Port > 0) {
			established = append(established, row)
		} else {
			listening = append(listening, row)
		}
	}
	if cache != nil {
		cache.stamp(established)
		cache.stamp(listening)
	}
	return established, listening, nil
}

func formatAddr(ip string, port uint32) string {
	if ip == "" && port == 0 {
		return "-"
	}
	if ip == "" {
		ip = "*"
	}
	// compress full IPv6
	if strings.Contains(ip, ":") && !strings.HasPrefix(ip, "[") {
		parsed := net.ParseIP(ip)
		if parsed != nil {
			ip = parsed.String()
		}
		return fmt.Sprintf("[%s]:%d", ip, port)
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

func procInfo(pid int32) (name, username string, sent, recv uint64) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return "-", "-", 0, 0
	}
	name, _ = p.Name()
	if name == "" {
		name = "-"
	}
	if io, err := p.IOCounters(); err == nil {
		sent = io.WriteBytes
		recv = io.ReadBytes
	}
	uids, err := p.Uids()
	if err == nil && len(uids) > 0 {
		u, err := user.LookupId(strconv.Itoa(int(uids[0])))
		if err == nil {
			username = u.Username
		}
	}
	if username == "" {
		username = "-"
	}
	return
}

// ---------------------------------------------------------------------------
// Interface stats
// ---------------------------------------------------------------------------

// IfaceRow is one row in the network stats table.
type IfaceRow struct {
	Name      string
	IP        string
	TotalSent string
	TotalRecv string
	TxRate    string
	RxRate    string
	PktSent   uint64
	PktRecv   uint64
	IsUp      bool
}

// IfaceSnapshot holds raw counters for delta computation.
type IfaceSnapshot struct {
	counters map[string]gnet.IOCountersStat
	at       time.Time
}

// NewIfaceSnapshot takes an initial snapshot (no rate data yet).
func NewIfaceSnapshot() (*IfaceSnapshot, error) {
	counters, err := gnet.IOCounters(true)
	if err != nil {
		return nil, err
	}
	m := make(map[string]gnet.IOCountersStat, len(counters))
	for _, c := range counters {
		m[c.Name] = c
	}
	return &IfaceSnapshot{counters: m, at: time.Now()}, nil
}

// IfaceStats computes per-interface stats given a previous snapshot.
// Returns rows, a new snapshot to use next time, and any error.
func IfaceStats(prev *IfaceSnapshot) ([]IfaceRow, *IfaceSnapshot, error) {
	counters, err := gnet.IOCounters(true)
	if err != nil {
		return nil, prev, err
	}

	ifaces, err := gnet.Interfaces()
	if err != nil {
		return nil, prev, err
	}

	addrMap := make(map[string]string)
	upMap := make(map[string]bool)
	for _, iface := range ifaces {
		for _, a := range iface.Addrs {
			ip, _, err := net.ParseCIDR(a.Addr)
			if err == nil && ip.To4() != nil {
				addrMap[iface.Name] = ip.String()
				break
			}
		}
		up := false
		for _, f := range iface.Flags {
			if f == "up" {
				up = true
				break
			}
		}
		upMap[iface.Name] = up
	}

	now := time.Now()
	dt := now.Sub(prev.at).Seconds()
	if dt <= 0 {
		dt = 1
	}

	newSnap := &IfaceSnapshot{counters: make(map[string]gnet.IOCountersStat, len(counters)), at: now}
	var rows []IfaceRow

	for _, c := range counters {
		newSnap.counters[c.Name] = c

		txRate, rxRate := uint64(0), uint64(0)
		if old, ok := prev.counters[c.Name]; ok {
			sentDiff := saturatingSub(c.BytesSent, old.BytesSent)
			recvDiff := saturatingSub(c.BytesRecv, old.BytesRecv)
			txRate = uint64(float64(sentDiff) / dt)
			rxRate = uint64(float64(recvDiff) / dt)
		}

		ip := addrMap[c.Name]
		if ip == "" {
			ip = "-"
		}

		rows = append(rows, IfaceRow{
			Name:      c.Name,
			IP:        ip,
			TotalSent: FormatBytes(c.BytesSent),
			TotalRecv: FormatBytes(c.BytesRecv),
			TxRate:    FormatBytes(txRate) + "/s",
			RxRate:    FormatBytes(rxRate) + "/s",
			PktSent:   c.PacketsSent,
			PktRecv:   c.PacketsRecv,
			IsUp:      upMap[c.Name],
		})
	}

	return rows, newSnap, nil
}

func saturatingSub(a, b uint64) uint64 {
	if a < b {
		return 0
	}
	return a - b
}

// FormatBytes formats a byte count to a human-readable string.
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

