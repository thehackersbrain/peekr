package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/thehackersbrain/peekr/internal/collector"
	"github.com/thehackersbrain/peekr/internal/theme"
)

// Fixed netstats column widths.
const (
	fixedTotalSent = 12
	fixedTotalRecv = 12
	fixedTxRate    = 12
	fixedRxRate    = 12
	fixedPktSent   = 11
	fixedPktRecv   = 11
	// 7 separating spaces between 8 columns
	netFixedOverhead = fixedTotalSent + fixedTotalRecv + fixedTxRate + fixedRxRate + fixedPktSent + fixedPktRecv + 7
)

// netLayout holds dynamically computed column widths for network stats.
type netLayout struct {
	iface, ip int
	rowW      int
}

func makeNetLayout(termW int) netLayout {
	avail := termW - netFixedOverhead
	if avail < 28 {
		avail = 28
	}
	// Interface names are short; give them 1/4 of space, capped at 20.
	iface := avail / 4
	if iface < 12 {
		iface = 12
	}
	if iface > 20 {
		iface = 20
	}
	ip := avail - iface
	rw := iface + 1 + ip + 1 + fixedTotalSent + 1 + fixedTotalRecv + 1 + fixedTxRate + 1 + fixedRxRate + 1 + fixedPktSent + 1 + fixedPktRecv
	return netLayout{iface, ip, rw}
}

type netStatsModel struct {
	width  int
	height int
	styles theme.Styles

	snap    *collector.IfaceSnapshot
	allRows []collector.IfaceRow
	rows    []collector.IfaceRow

	cursor       int
	offset       int
	hideInactive bool

	popup []string
}

func newNetStatsModel(s theme.Styles, snap *collector.IfaceSnapshot) netStatsModel {
	return netStatsModel{styles: s, snap: snap}
}

func (m *netStatsModel) applyFilter() {
	if !m.hideInactive {
		m.rows = m.allRows
		return
	}
	var out []collector.IfaceRow
	for _, r := range m.allRows {
		if r.IsUp && r.Name != "lo" {
			out = append(out, r)
		}
	}
	m.rows = out
}

func (m netStatsModel) update(msg tea.Msg) (netStatsModel, tea.Cmd, bool) {
	if m.popup != nil {
		if _, ok := msg.(tea.KeyMsg); ok {
			m.popup = nil
		}
		return m, nil, false
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace", "left", "esc":
			return m, nil, true
		case "h":
			m.hideInactive = !m.hideInactive
			m.applyFilter()
			m.cursor = 0
			m.offset = 0
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else if m.offset > 0 {
				m.offset--
			}
		case "down", "j":
			maxV := m.visibleRows()
			if m.cursor < maxV-1 && m.cursor < len(m.rows)-1 {
				m.cursor++
			} else if m.offset+maxV < len(m.rows) {
				m.offset++
			}
		case "enter":
			idx := m.offset + m.cursor
			if idx < len(m.rows) {
				m.popup = buildIfacePopup(m.rows[idx])
			}
		case "?":
			m.popup = []string{
				"  ↑/↓ or j/k  scroll",
				"  h            toggle hide inactive",
				"  enter        interface details",
				"  t            theme picker",
				"  backspace    back to menu",
				"  q            quit",
			}
		}
	}
	return m, nil, false
}

func buildIfacePopup(r collector.IfaceRow) []string {
	status := "DOWN"
	if r.IsUp {
		status = "UP"
	}
	return []string{
		fmt.Sprintf("  Interface: %s", r.Name),
		fmt.Sprintf("  IP:        %s", r.IP),
		fmt.Sprintf("  Status:    %s", status),
		"",
		fmt.Sprintf("  Total Sent: %s", r.TotalSent),
		fmt.Sprintf("  Total Recv: %s", r.TotalRecv),
		fmt.Sprintf("  TX Rate:    %s", r.TxRate),
		fmt.Sprintf("  RX Rate:    %s", r.RxRate),
		"",
		fmt.Sprintf("  Pkts Sent:  %d", r.PktSent),
		fmt.Sprintf("  Pkts Recv:  %d", r.PktRecv),
		"",
		"  press any key to close",
	}
}

// clampScroll ensures offset/cursor stay in bounds after the rows slice shrinks.
func (m *netStatsModel) clampScroll() {
	n := len(m.rows)
	if n == 0 {
		m.offset = 0
		m.cursor = 0
		return
	}
	if m.offset >= n {
		m.offset = n - 1
	}
	maxV := m.visibleRows()
	if maxV < 1 {
		maxV = 1
	}
	if m.cursor >= maxV || m.offset+m.cursor >= n {
		m.cursor = min(maxV-1, n-m.offset-1)
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m netStatsModel) visibleRows() int {
	v := m.height - 5
	if v < 1 {
		return 1
	}
	return v
}

func (m netStatsModel) bg(w int) lipgloss.Style {
	return m.styles.Base.Width(w)
}

func (m netStatsModel) view(w, h int) string {
	if m.popup != nil {
		return renderPopup(m.styles, "Interface Details", m.popup, w, h)
	}

	s := m.styles
	blank := m.bg(w).Render("")

	hideHint := "h=hide inactive"
	if m.hideInactive {
		hideHint = "h=show inactive"
	}

	lay := makeNetLayout(w)

	var lines []string
	lines = append(lines, s.Title.Width(w).Render(" Network Statistics "))
	lines = append(lines, s.Muted.Width(w).Render(fmt.Sprintf(
		"  ↑/↓ scroll  %s  enter=details  ?=help  t=theme  backspace=back  q=quit", hideHint,
	)))
	lines = append(lines, s.Header.Width(w).Render(
		truncPad("INTERFACE", lay.iface)+" "+
			truncPad("IP", lay.ip)+" "+
			truncPad("TOTAL SENT", fixedTotalSent)+" "+
			truncPad("TOTAL RECV", fixedTotalRecv)+" "+
			truncPad("TX/s", fixedTxRate)+" "+
			truncPad("RX/s", fixedRxRate)+" "+
			truncPad("PKT SENT", fixedPktSent)+" "+
			truncPad("PKT RECV", fixedPktRecv),
	))
	lines = append(lines, s.Separator.Width(w).Render(strings.Repeat("─", w)))

	visRows := h - len(lines)
	if visRows < 1 {
		visRows = 1
	}

	if len(m.rows) == 0 {
		lines = append(lines, s.Muted.Width(w).Render("  (no interfaces)"))
	} else {
		rowBG := lipgloss.Color(s.T.BG)
		cell := func(fg string) lipgloss.Style {
			return lipgloss.NewStyle().Background(rowBG).Foreground(lipgloss.Color(fg))
		}
		cName := cell(s.T.Accent)
		cIP := cell(s.T.Accent4)
		cTSent := cell(s.T.Accent3)
		cTRecv := cell(s.T.Accent2)
		cTxHi := cell(s.T.Accent5)
		cRxHi := cell(s.T.Accent5)
		cMuted := cell(s.T.Muted)
		cSep := cell(s.T.FG)
		cPad := lipgloss.NewStyle().Background(rowBG)

		end := min(m.offset+visRows, len(m.rows))
		for i, row := range m.rows[m.offset:end] {
			absIdx := m.offset + i
			pktSent := fmt.Sprintf("%d", row.PktSent)
			pktRecv := fmt.Sprintf("%d", row.PktRecv)

			if absIdx == m.offset+m.cursor {
				lines = append(lines, s.Selected.Width(w).Render(
					truncPad(row.Name, lay.iface)+" "+
						truncPad(row.IP, lay.ip)+" "+
						truncPad(row.TotalSent, fixedTotalSent)+" "+
						truncPad(row.TotalRecv, fixedTotalRecv)+" "+
						truncPad(row.TxRate, fixedTxRate)+" "+
						truncPad(row.RxRate, fixedRxRate)+" "+
						truncPad(pktSent, fixedPktSent)+" "+
						truncPad(pktRecv, fixedPktRecv),
				))
			} else {
				txStyle := cMuted
				if row.TxRate != "0 B/s" && row.TxRate != "0B/s" {
					txStyle = cTxHi
				}
				rxStyle := cMuted
				if row.RxRate != "0 B/s" && row.RxRate != "0B/s" {
					rxStyle = cRxHi
				}

				line := cName.Render(truncPad(row.Name, lay.iface)) + cSep.Render(" ") +
					cIP.Render(truncPad(row.IP, lay.ip)) + cSep.Render(" ") +
					cTSent.Render(truncPad(row.TotalSent, fixedTotalSent)) + cSep.Render(" ") +
					cTRecv.Render(truncPad(row.TotalRecv, fixedTotalRecv)) + cSep.Render(" ") +
					txStyle.Render(truncPad(row.TxRate, fixedTxRate)) + cSep.Render(" ") +
					rxStyle.Render(truncPad(row.RxRate, fixedRxRate)) + cSep.Render(" ") +
					cMuted.Render(truncPad(pktSent, fixedPktSent)) + cSep.Render(" ") +
					cMuted.Render(truncPad(pktRecv, fixedPktRecv))

				if pad := w - lay.rowW; pad > 0 {
					line += cPad.Render(strings.Repeat(" ", pad))
				}
				lines = append(lines, line)
			}
		}
	}

	for len(lines) < h {
		lines = append(lines, blank)
	}
	return strings.Join(lines[:h], "\n")
}
