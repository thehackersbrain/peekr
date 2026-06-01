package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/thehackersbrain/peekr/internal/collector"
	"github.com/thehackersbrain/peekr/internal/theme"
)

// Fixed column widths that don't grow with terminal width.
const (
	fixedStatus  = 11
	fixedPID     = 7
	fixedProcess = 18
	fixedUser    = 12
	fixedSent    = 10
	fixedRecv    = 10
	fixedAge     = 8
	// 8 separating spaces between 9 columns
	fixedOverhead = fixedStatus + fixedPID + fixedProcess + fixedUser + fixedSent + fixedRecv + fixedAge + 8
)

// connLayout holds dynamically computed column widths for a given terminal width.
type connLayout struct {
	local, remote int // grow to fill available space
	rowW          int // total printable row width
}

func makeConnLayout(termW int) connLayout {
	avail := termW - fixedOverhead
	if avail < 40 {
		avail = 40 // minimum 20 chars each
	}
	l := avail / 2
	r := avail - l
	return connLayout{
		local:  l,
		remote: r,
		rowW:   l + 1 + r + 1 + fixedStatus + 1 + fixedPID + 1 + fixedProcess + 1 + fixedUser + 1 + fixedSent + 1 + fixedRecv + 1 + fixedAge,
	}
}

type connsModel struct {
	width  int
	height int
	styles theme.Styles
	mode   screen

	established []collector.ConnRow
	listening   []collector.ConnRow
	displayed   []collector.ConnRow

	cursor     int
	offset     int
	searchTerm string
	searching  bool
	searchBuf  string

	activePane int

	popup []string
}

func newConnsModel(s theme.Styles) connsModel {
	return connsModel{styles: s, mode: screenEstablished}
}

func (m *connsModel) applyFilter() {
	var src []collector.ConnRow
	switch m.mode {
	case screenEstablished:
		src = m.established
	case screenListening:
		src = m.listening
	case screenBoth:
		if m.activePane == 0 {
			src = m.established
		} else {
			src = m.listening
		}
	}

	needle := m.searchTerm
	if m.searching {
		needle = m.searchBuf // live filter while typing
	}
	if needle == "" {
		m.displayed = src
		return
	}
	needle = strings.ToLower(needle)
	var out []collector.ConnRow
	for _, r := range src {
		if strings.Contains(strings.ToLower(r.LocalAddr), needle) ||
			strings.Contains(strings.ToLower(r.RemoteAddr), needle) ||
			strings.Contains(strings.ToLower(r.Process), needle) {
			out = append(out, r)
		}
	}
	m.displayed = out
}

func (m connsModel) update(msg tea.Msg) (connsModel, tea.Cmd, bool) {
	if m.popup != nil {
		if _, ok := msg.(tea.KeyMsg); ok {
			m.popup = nil
		}
		return m, nil, false
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "enter":
				m.searchTerm = m.searchBuf
				m.searching = false
				m.cursor = 0
				m.offset = 0
				m.applyFilter()
			case "esc":
				m.searching = false
				m.searchBuf = m.searchTerm
				m.applyFilter() // restore committed filter
			case "backspace":
				if len(m.searchBuf) > 0 {
					r := []rune(m.searchBuf)
					m.searchBuf = string(r[:len(r)-1])
					m.applyFilter()
				}
			default:
				if len(msg.String()) == 1 {
					m.searchBuf += msg.String()
					m.applyFilter()
				}
			}
			return m, nil, false
		}

		switch msg.String() {
		case "backspace", "left", "esc":
			return m, nil, true
		case "s":
			m.searching = true
			m.searchBuf = m.searchTerm
		case "tab":
			if m.mode == screenBoth {
				m.activePane = 1 - m.activePane
				m.cursor = 0
				m.offset = 0
				m.applyFilter()
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else if m.offset > 0 {
				m.offset--
			}
		case "down", "j":
			maxV := m.visibleRows()
			if m.cursor < maxV-1 && m.cursor < len(m.displayed)-1 {
				m.cursor++
			} else if m.offset+maxV < len(m.displayed) {
				m.offset++
			}
		case "g":
			m.cursor = 0
			m.offset = 0
		case "G":
			m.offset = max(0, len(m.displayed)-m.visibleRows())
			m.cursor = clamp(len(m.displayed)-m.offset-1, 0, m.visibleRows()-1)
		case "enter":
			idx := m.offset + m.cursor
			if idx < len(m.displayed) {
				m.popup = buildConnPopup(m.displayed[idx])
			}
		case "?":
			m.popup = []string{
				"  ↑/↓ or j/k  scroll",
				"  tab          switch pane (both view)",
				"  s            filter connections",
				"  g / G        jump top / bottom",
				"  enter        connection details",
				"  t            theme picker",
				"  backspace    back to menu",
				"  q            quit",
			}
		}
	}
	return m, nil, false
}

func buildConnPopup(r collector.ConnRow) []string {
	return []string{
		fmt.Sprintf("  Local:    %s", r.LocalAddr),
		fmt.Sprintf("  Remote:   %s", r.RemoteAddr),
		fmt.Sprintf("  Status:   %s", r.Status),
		"",
		fmt.Sprintf("  PID:      %d", r.PID),
		fmt.Sprintf("  Process:  %s", r.Process),
		fmt.Sprintf("  User:     %s", r.User),
		"",
		fmt.Sprintf("  Sent:     %s", collector.FormatBytes(r.SentBytes)),
		fmt.Sprintf("  Recv:     %s", collector.FormatBytes(r.RecvBytes)),
		"",
		fmt.Sprintf("  First seen: %s ago (%s)", formatAge(r.FirstSeen), r.FirstSeen.Format("15:04:05")),
		"",
		"  press any key to close",
	}
}

// clampScroll ensures offset/cursor stay in bounds after the displayed slice shrinks.
func (m *connsModel) clampScroll() {
	n := len(m.displayed)
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

func (m connsModel) visibleRows() int {
	if m.mode == screenBoth {
		h := (m.height - 7) / 2
		if h < 1 {
			return 1
		}
		return h
	}
	h := m.height - 5
	if h < 1 {
		return 1
	}
	return h
}

func (m connsModel) view(w, h int) string {
	if m.popup != nil {
		return renderPopup(m.styles, "Connection Details", m.popup, w, h)
	}
	switch m.mode {
	case screenBoth:
		return m.viewBoth(w, h)
	default:
		return m.viewSingle(w, h)
	}
}

func (m connsModel) viewSingle(w, h int) string {
	s := m.styles
	lay := makeConnLayout(w)
	blank := s.Base.Width(w).Render("")

	modeLabel := "Established"
	if m.mode == screenListening {
		modeLabel = "Listening"
	}

	var lines []string
	lines = append(lines, s.Title.Width(w).Render(fmt.Sprintf(" %s Connections ", modeLabel)))
	if m.searching {
		lines = append(lines, s.Accent.Width(w).Render("  SEARCH  —  type to filter  •  Enter=confirm  •  Esc=cancel"))
	} else {
		lines = append(lines, s.Muted.Width(w).Render("  ↑/↓ scroll  s=search  enter=details  ?=help  t=theme  backspace=back  q=quit"))
	}

	rowBG := lipgloss.Color(s.T.BG)
	bgOnly := lipgloss.NewStyle().Background(rowBG)
	if m.searching {
		prompt := "  Search: "
		cur := "█"
		padLen := w - len([]rune(prompt)) - len([]rune(m.searchBuf)) - len([]rune(cur))
		if padLen < 0 {
			padLen = 0
		}
		line := lipgloss.NewStyle().Background(rowBG).Foreground(lipgloss.Color(s.T.Accent)).Render(prompt) +
			lipgloss.NewStyle().Background(rowBG).Foreground(lipgloss.Color(s.T.FG)).Render(m.searchBuf) +
			lipgloss.NewStyle().Background(rowBG).Foreground(lipgloss.Color(s.T.Accent2)).Render(cur) +
			bgOnly.Render(strings.Repeat(" ", padLen))
		lines = append(lines, line)
	} else if m.searchTerm != "" {
		text := fmt.Sprintf("  Filter: %q — %d shown", m.searchTerm, len(m.displayed))
		padLen := w - len([]rune(text))
		if padLen < 0 {
			padLen = 0
		}
		line := lipgloss.NewStyle().Background(rowBG).Foreground(lipgloss.Color(s.T.Muted)).Render(text) +
			bgOnly.Render(strings.Repeat(" ", padLen))
		lines = append(lines, line)
	}

	lines = append(lines, m.renderHeader(w, lay))
	lines = append(lines, m.renderSep(w))

	visRows := h - len(lines)
	if visRows < 1 {
		visRows = 1
	}
	lines = append(lines, m.renderRows(w, visRows, lay)...)

	for len(lines) < h {
		lines = append(lines, blank)
	}
	return strings.Join(lines[:h], "\n")
}

func (m connsModel) viewBoth(w, h int) string {
	s := m.styles
	lay := makeConnLayout(w)
	blank := s.Base.Width(w).Render("")

	rowBGBoth := lipgloss.Color(s.T.BG)
	bgOnlyBoth := lipgloss.NewStyle().Background(rowBGBoth)

	var lines []string
	lines = append(lines, s.Title.Width(w).Render(" Connections (Both) "))
	if m.searching {
		lines = append(lines, s.Accent.Width(w).Render("  SEARCH  —  type to filter  •  Enter=confirm  •  Esc=cancel"))
	} else {
		lines = append(lines, s.Muted.Width(w).Render("  tab=switch pane  ↑/↓ scroll  s=search  enter=details  ?=help  t=theme  backspace=back  q=quit"))
	}
	if m.searching {
		prompt := "  Search: "
		cur := "█"
		padLen := w - len([]rune(prompt)) - len([]rune(m.searchBuf)) - len([]rune(cur))
		if padLen < 0 {
			padLen = 0
		}
		searchLine := lipgloss.NewStyle().Background(rowBGBoth).Foreground(lipgloss.Color(s.T.Accent)).Render(prompt) +
			lipgloss.NewStyle().Background(rowBGBoth).Foreground(lipgloss.Color(s.T.FG)).Render(m.searchBuf) +
			lipgloss.NewStyle().Background(rowBGBoth).Foreground(lipgloss.Color(s.T.Accent2)).Render(cur) +
			bgOnlyBoth.Render(strings.Repeat(" ", padLen))
		lines = append(lines, searchLine)
	}
	lines = append(lines, m.renderHeader(w, lay))
	lines = append(lines, m.renderSep(w))

	// Extra line when search bar is visible reduces space for each pane.
	searchExtra := 0
	if m.searching {
		searchExtra = 1
	}
	paneRows := (h - 7 - searchExtra) / 2
	if paneRows < 1 {
		paneRows = 1
	}

	origMode := m.mode
	origDisplayed := m.displayed
	origCursor := m.cursor
	origOffset := m.offset

	// ── Established pane ──────────────────────────────────────────
	m.mode = screenEstablished
	m.displayed = m.established
	if m.searchTerm != "" {
		m.applyFilter()
	}
	if m.activePane != 0 {
		m.cursor = -1 // -1 so no row is highlighted in the inactive pane
		m.offset = 0
	} else {
		m.cursor = origCursor
		m.offset = origOffset
	}
	if m.activePane == 0 {
		lines = append(lines, s.Selected.Width(w).Render(" ▸ Established "))
	} else {
		lines = append(lines, s.Accent3.Width(w).Render(" ▸ Established "))
	}
	estRows := m.renderRows(w, paneRows, lay)
	lines = append(lines, estRows...)
	for len(lines) < 5+paneRows {
		lines = append(lines, blank)
	}

	// ── Separator + Listening pane ────────────────────────────────
	lines = append(lines, m.renderSep(w))

	m.mode = screenListening
	m.displayed = m.listening
	if m.searchTerm != "" {
		m.applyFilter()
	}
	if m.activePane != 1 {
		m.cursor = -1 // -1 so no row is highlighted in the inactive pane
		m.offset = 0
	} else {
		m.cursor = origCursor
		m.offset = origOffset
	}
	if m.activePane == 1 {
		lines = append(lines, s.Selected.Width(w).Render(" ▸ Listening "))
	} else {
		lines = append(lines, s.Accent4.Width(w).Render(" ▸ Listening "))
	}
	lisRows := m.renderRows(w, paneRows, lay)
	lines = append(lines, lisRows...)

	m.mode = origMode
	m.displayed = origDisplayed
	m.cursor = origCursor
	m.offset = origOffset

	for len(lines) < h {
		lines = append(lines, blank)
	}
	return strings.Join(lines[:h], "\n")
}

func (m connsModel) renderHeader(w int, lay connLayout) string {
	return m.styles.Header.Width(w).Render(
		truncPad("LOCAL ADDRESS", lay.local) + " " +
			truncPad("REMOTE ADDRESS", lay.remote) + " " +
			truncPad("STATUS", fixedStatus) + " " +
			truncPad("PID", fixedPID) + " " +
			truncPad("PROCESS", fixedProcess) + " " +
			truncPad("USER", fixedUser) + " " +
			truncPad("SENT", fixedSent) + " " +
			truncPad("RECV", fixedRecv) + " " +
			truncPad("TIME", fixedAge),
	)
}

func (m connsModel) renderSep(w int) string {
	return m.styles.Separator.Width(w).Render(strings.Repeat("─", w))
}

func (m connsModel) renderRows(w, maxRows int, lay connLayout) []string {
	s := m.styles

	if len(m.displayed) == 0 {
		return []string{s.Muted.Width(w).Render("  (no connections)")}
	}

	rowBG := lipgloss.Color(s.T.BG)
	cell := func(fg string) lipgloss.Style {
		return lipgloss.NewStyle().Background(rowBG).Foreground(lipgloss.Color(fg))
	}
	cLocal := cell(s.T.Accent)
	cRemote := cell(s.T.Accent2)
	cEst := cell(s.T.Accent3)
	cLis := cell(s.T.Accent4)
	cOth := cell(s.T.Accent2)
	cPID := cell(s.T.Accent4)
	cProc := cell(s.T.Accent5)
	cUser := cell(s.T.Muted)
	cSent := cell(s.T.Accent2)
	cRecv := cell(s.T.Accent3)
	cSep := cell(s.T.FG)
	cPad := lipgloss.NewStyle().Background(rowBG)

	var lines []string
	end := min(m.offset+maxRows, len(m.displayed))

	for i, row := range m.displayed[m.offset:end] {
		absIdx := m.offset + i
		pidStr := fmt.Sprintf("%d", row.PID)
		sentStr := collector.FormatBytes(row.SentBytes)
		recvStr := collector.FormatBytes(row.RecvBytes)
		ageStr := row.FirstSeen.Format("15:04:05")

		if absIdx == m.offset+m.cursor {
			plain := truncPad(row.LocalAddr, lay.local) + " " +
				truncPad(row.RemoteAddr, lay.remote) + " " +
				truncPad(row.Status, fixedStatus) + " " +
				truncPad(pidStr, fixedPID) + " " +
				truncPad(row.Process, fixedProcess) + " " +
				truncPad(row.User, fixedUser) + " " +
				truncPad(sentStr, fixedSent) + " " +
				truncPad(recvStr, fixedRecv) + " " +
				truncPad(ageStr, fixedAge)
			lines = append(lines, s.Selected.Width(w).Render(plain))
			continue
		}

		var cStatus lipgloss.Style
		switch row.Status {
		case "ESTABLISHED":
			cStatus = cEst
		case "LISTEN":
			cStatus = cLis
		default:
			cStatus = cOth
		}

		cAge := cell(s.T.Muted)
		line := cLocal.Render(truncPad(row.LocalAddr, lay.local)) + cSep.Render(" ") +
			cRemote.Render(truncPad(row.RemoteAddr, lay.remote)) + cSep.Render(" ") +
			cStatus.Render(truncPad(row.Status, fixedStatus)) + cSep.Render(" ") +
			cPID.Render(truncPad(pidStr, fixedPID)) + cSep.Render(" ") +
			cProc.Render(truncPad(row.Process, fixedProcess)) + cSep.Render(" ") +
			cUser.Render(truncPad(row.User, fixedUser)) + cSep.Render(" ") +
			cSent.Render(truncPad(sentStr, fixedSent)) + cSep.Render(" ") +
			cRecv.Render(truncPad(recvStr, fixedRecv)) + cSep.Render(" ") +
			cAge.Render(truncPad(ageStr, fixedAge))

		if pad := w - lay.rowW; pad > 0 {
			line += cPad.Render(strings.Repeat(" ", pad))
		}
		lines = append(lines, line)
	}
	return lines
}

func formatAge(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", m, s)
	case d < 24*time.Hour:
		h := int(d.Hours())
		mn := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh%dm", h, mn)
	default:
		days := int(d.Hours()) / 24
		h := int(d.Hours()) % 24
		return fmt.Sprintf("%dd%dh", days, h)
	}
}

func truncPad(s string, width int) string {
	runes := []rune(s)
	if len(runes) > width {
		if width > 1 {
			return string(runes[:width-1]) + "…"
		}
		return string(runes[:width])
	}
	return s + strings.Repeat(" ", width-len(runes))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
