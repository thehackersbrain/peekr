package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/thehackersbrain/peekr/internal/theme"
)

type menuItem struct {
	label  string
	target screen
}

var menuItems = []menuItem{
	{"Established Connections", screenEstablished},
	{"Listening Connections", screenListening},
	{"Both (Split View)", screenBoth},
	{"Network Statistics", screenNetStats},
}

type menuModel struct {
	cursor int
	width  int
	height int
	styles theme.Styles

	// live summary counters shown on menu
	estCount int
	lisCount int
}

func newMenuModel(s theme.Styles) menuModel {
	return menuModel{styles: s}
}

func (m menuModel) update(msg tea.Msg) (menuModel, tea.Cmd, screen) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m, nil, menuItems[m.cursor].target
		}
	}
	return m, nil, screenMenu
}

func (m menuModel) view(w, h int) string {
	s := m.styles
	blank := s.Base.Width(w).Render("")

	// Stats bar
	estStr := s.Accent3.Render(fmt.Sprintf("%d", m.estCount))
	lisStr := s.Accent4.Render(fmt.Sprintf("%d", m.lisCount))
	statsLine := s.Muted.Render("  established: ") + estStr +
		s.Muted.Render("   listening: ") + lisStr

	// Menu items
	var itemLines []string
	for i, item := range menuItems {
		if i == m.cursor {
			arrow := s.Accent.Render("▶ ")
			label := s.Selected.Width(36).Render(item.label)
			itemLines = append(itemLines, s.Base.Width(w).Render(arrow+label))
		} else {
			itemLines = append(itemLines, s.Base.Width(w).Render("  "+item.label))
		}
	}

	// Build content block, vertically centered
	contentLines := []string{
		s.Title.Width(w).Render(" peekr "),
		s.Muted.Width(w).Render(" network monitor  •  t=themes  •  q=quit"),
		blank,
		s.Base.Width(w).Render(statsLine),
		blank,
	}
	contentLines = append(contentLines, itemLines...)
	contentLines = append(
		contentLines,
		blank,
		s.Muted.Width(w).Render("  ↑/↓ or j/k navigate  •  enter select  •  t themes  •  q quit"),
	)

	// Pad top to vertically center the content block
	topPad := (h - len(contentLines)) / 2
	if topPad < 0 {
		topPad = 0
	}

	var lines []string
	for i := 0; i < topPad; i++ {
		lines = append(lines, blank)
	}
	lines = append(lines, contentLines...)
	for len(lines) < h {
		lines = append(lines, blank)
	}
	return strings.Join(lines[:h], "\n")
}
