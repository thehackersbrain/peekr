// Package ui provides all Bubble Tea TUI models for peekr.
package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/thehackersbrain/peekr/internal/collector"
	"github.com/thehackersbrain/peekr/internal/theme"
)

// ---------------------------------------------------------------------------
// Screen identifiers
// ---------------------------------------------------------------------------

type screen int

const (
	screenMenu screen = iota
	screenEstablished
	screenListening
	screenBoth
	screenNetStats
	screenThemePicker
)

// ---------------------------------------------------------------------------
// Shared tick messages
// ---------------------------------------------------------------------------

type (
	tickMsg        time.Time
	connRefreshMsg struct {
		established []collector.ConnRow
		listening   []collector.ConnRow
		err         error
	}
)

type netRefreshMsg struct {
	rows []collector.IfaceRow
	snap *collector.IfaceSnapshot
	err  error
}

func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// ---------------------------------------------------------------------------
// Root model
// ---------------------------------------------------------------------------

// Model is the root Bubble Tea model.
type Model struct {
	width    int
	height   int
	current  screen
	prev     screen
	interval time.Duration

	styles    theme.Styles
	connCache *collector.ConnCache

	menu        menuModel
	conns       connsModel
	netStats    netStatsModel
	themePicker themePickerModel
}

// New creates a root Model with the given theme and refresh interval.
func New(t *theme.Theme, interval time.Duration) Model {
	styles := t.Build()
	snap, _ := collector.NewIfaceSnapshot()

	m := Model{
		current:   screenMenu,
		interval:  interval,
		styles:    styles,
		connCache: collector.NewConnCache(),
	}
	m.menu = newMenuModel(styles)
	m.conns = newConnsModel(styles)
	m.netStats = newNetStatsModel(styles, snap)
	m.themePicker = newThemePickerModel(styles)
	return m
}

func (m Model) Init() tea.Cmd {
	return tickCmd(m.interval)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menu.width = msg.Width
		m.menu.height = msg.Height
		m.conns.width = msg.Width
		m.conns.height = msg.Height
		m.netStats.width = msg.Width
		m.netStats.height = msg.Height
		m.themePicker.width = msg.Width
		m.themePicker.height = msg.Height
		return m, nil

	case tickMsg:
		// Kick off background refreshes for the active data screens.
		var cmds []tea.Cmd
		if m.current == screenEstablished || m.current == screenListening ||
			m.current == screenBoth || m.current == screenMenu {
			cmds = append(cmds, fetchConns(m.connCache))
		}
		if m.current == screenNetStats || m.current == screenMenu {
			snap := m.netStats.snap
			cmds = append(cmds, fetchNet(snap))
		}
		cmds = append(cmds, tickCmd(m.interval))
		return m, tea.Batch(cmds...)

	case connRefreshMsg:
		m.conns.established = msg.established
		m.conns.listening = msg.listening
		m.conns.applyFilter()
		m.conns.clampScroll()
		m.menu.estCount = len(msg.established)
		m.menu.lisCount = len(msg.listening)
		return m, nil

	case netRefreshMsg:
		m.netStats.snap = msg.snap
		m.netStats.allRows = msg.rows
		m.netStats.applyFilter()
		m.netStats.clampScroll()
		return m, nil

	case tea.KeyMsg:
		// ctrl+c always quits.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Some sub-models have modes that must consume all keys before any
		// global shortcut fires (search input, popup visible, theme picker).
		// Let those sub-models handle the key first; only intercept globally
		// when no sub-model is in such a consuming state.
		consuming := false
		switch m.current {
		case screenEstablished, screenListening, screenBoth:
			consuming = m.conns.searching || m.conns.popup != nil
		case screenNetStats:
			consuming = m.netStats.popup != nil
		case screenThemePicker:
			consuming = true // theme picker owns all its own navigation
		}

		if !consuming {
			switch msg.String() {
			case "q":
				if m.current == screenMenu {
					return m, tea.Quit
				}
				m.current = screenMenu
				return m, nil
			case "esc", "backspace", "left":
				if m.current != screenMenu {
					m.current = screenMenu
					return m, nil
				}
			case "t":
				if m.current != screenThemePicker {
					m.prev = m.current
					m.themePicker.prev = m.current
					m.current = screenThemePicker
					return m, nil
				}
			}
		}
	}

	// Delegate to active sub-model.
	switch m.current {
	case screenMenu:
		newMenu, cmd, next := m.menu.update(msg)
		m.menu = newMenu
		if next != screenMenu {
			m.current = next
			m.conns.searching = false
			m.conns.searchBuf = ""
			m.conns.popup = nil
			m.conns.cursor = 0
			m.conns.offset = 0
			m.conns.mode = next
			m.conns.searchTerm = ""
			m.conns.applyFilter()
			m.netStats.popup = nil
			m.netStats.cursor = 0
			m.netStats.offset = 0
		}
		return m, cmd

	case screenEstablished, screenListening, screenBoth:
		newConns, cmd, goBack := m.conns.update(msg)
		m.conns = newConns
		if goBack {
			m.current = screenMenu
		}
		return m, cmd

	case screenNetStats:
		newNS, cmd, goBack := m.netStats.update(msg)
		m.netStats = newNS
		if goBack {
			m.current = screenMenu
		}
		return m, cmd

	case screenThemePicker:
		newTP, cmd, chosen, goBack := m.themePicker.update(msg)
		m.themePicker = newTP
		if goBack {
			m.current = m.prev
			return m, cmd
		}
		if chosen != nil {
			// Apply new theme everywhere.
			s := chosen.Build()
			m.styles = s
			m.menu.styles = s
			m.conns.styles = s
			m.netStats.styles = s
			m.themePicker.styles = s
			m.current = m.prev
		}
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	switch m.current {
	case screenMenu:
		return m.menu.view(m.width, m.height)
	case screenEstablished, screenListening, screenBoth:
		return m.conns.view(m.width, m.height)
	case screenNetStats:
		return m.netStats.view(m.width, m.height)
	case screenThemePicker:
		return m.themePicker.view(m.width, m.height)
	}
	return ""
}

// ---------------------------------------------------------------------------
// Background fetch commands
// ---------------------------------------------------------------------------

func fetchConns(cache *collector.ConnCache) tea.Cmd {
	return func() tea.Msg {
		est, lis, err := collector.Connections(cache)
		return connRefreshMsg{established: est, listening: lis, err: err}
	}
}

func fetchNet(snap *collector.IfaceSnapshot) tea.Cmd {
	return func() tea.Msg {
		if snap == nil {
			snap, _ = collector.NewIfaceSnapshot()
		}
		rows, newSnap, err := collector.IfaceStats(snap)
		return netRefreshMsg{rows: rows, snap: newSnap, err: err}
	}
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// renderPopup renders a centered detail popup over a filled background.
func renderPopup(s theme.Styles, title string, lines []string, w, h int) string {
	maxLen := len(title)
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}
	innerW := maxLen + 2
	if innerW > w-6 {
		innerW = w - 6
	}
	if innerW < 10 {
		innerW = 10
	}

	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(s.T.Accent4)).
		Padding(0, 1).
		Width(innerW)

	titleLine := s.Title.Render(" " + title + " ")
	divider := s.Separator.Render(strings.Repeat("─", innerW))
	var body []string
	body = append(body, titleLine, divider)
	for _, l := range lines {
		body = append(body, s.Base.Render(l))
	}
	popup := popupStyle.Render(strings.Join(body, "\n"))

	return lipgloss.Place(
		w, h,
		lipgloss.Center, lipgloss.Center,
		popup,
		lipgloss.WithWhitespaceBackground(lipgloss.Color(s.T.BG)),
	)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
