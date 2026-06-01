package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/thehackersbrain/peekr/internal/theme"
)

type themePickerModel struct {
	width  int
	height int
	styles theme.Styles
	cursor int
	prev   screen
}

func newThemePickerModel(s theme.Styles) themePickerModel {
	// Start cursor at current theme.
	cursor := 0
	for i, t := range theme.All {
		if t.Name == s.T.Name {
			cursor = i
			break
		}
	}
	return themePickerModel{styles: s, cursor: cursor}
}

// update returns (model, cmd, chosenTheme, goBack)
func (m themePickerModel) update(msg tea.Msg) (themePickerModel, tea.Cmd, *theme.Theme, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(theme.All)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m, nil, theme.All[m.cursor], false
		case "esc", "backspace", "left", "t", "q":
			return m, nil, nil, true
		}
	}
	return m, nil, nil, false
}

func (m themePickerModel) view(w, h int) string {
	s := m.styles
	blank := s.Base.Width(w).Render("")

	listWidth := 28
	previewWidth := w - listWidth - 3
	if previewWidth < 20 {
		previewWidth = 20
	}

	// Build list lines.
	var listLines []string
	for i, t := range theme.All {
		label := fmt.Sprintf("  %-24s", t.Name)
		if i == m.cursor {
			listLines = append(listLines, s.Selected.Width(listWidth).Render(label))
		} else if t.Name == s.T.Name {
			listLines = append(listLines, s.Accent3.Width(listWidth).Render("• "+label[2:]))
		} else {
			listLines = append(listLines, s.Base.Width(listWidth).Render(label))
		}
	}

	hovered := theme.All[m.cursor]
	preview := renderThemePreview(hovered, previewWidth, h-4)
	previewLines := strings.Split(preview, "\n")

	// Build combined rows (list | divider | preview).
	numRows := max(len(listLines), len(previewLines))
	var rows []string
	for i := 0; i < numRows; i++ {
		listCell := ""
		if i < len(listLines) {
			listCell = listLines[i]
		} else {
			listCell = s.Base.Width(listWidth).Render("")
		}
		divCell := s.Muted.Render(" │")
		previewCell := ""
		if i < len(previewLines) {
			previewCell = s.Base.Width(previewWidth).Render(previewLines[i])
		} else {
			previewCell = s.Base.Width(previewWidth).Render("")
		}
		rows = append(rows, listCell+divCell+previewCell)
	}

	var lines []string
	lines = append(lines, s.Title.Width(w).Render(" Theme Picker "))
	lines = append(lines, s.Muted.Width(w).Render("  ↑/↓ or j/k navigate  •  enter/space select  •  esc/backspace cancel"))
	lines = append(lines, blank)
	lines = append(lines, rows...)

	for len(lines) < h {
		lines = append(lines, blank)
	}
	return strings.Join(lines[:h], "\n")
}

// renderThemePreview draws a self-contained swatch + fake-UI preview of a theme.
func renderThemePreview(t *theme.Theme, w, h int) string {
	bg := lipgloss.Color(t.BG)
	fg := lipgloss.Color(t.FG)
	mu := lipgloss.Color(t.Muted)
	base := lipgloss.NewStyle().Background(bg).Foreground(fg)
	muted := lipgloss.NewStyle().Background(bg).Foreground(mu)

	var lines []string

	// Theme name
	nameStyle := lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color(t.Accent)).Bold(true)
	lines = append(lines, nameStyle.Render(fmt.Sprintf("  %s", t.Name)))
	lines = append(lines, base.Render(""))

	// Palette swatches — two rows
	swatchRow1 := "  "
	swatchRow2 := "  "
	for i, hex := range t.Palette {
		swatch := lipgloss.NewStyle().
			Background(lipgloss.Color(hex)).
			Foreground(lipgloss.Color(hex)).
			Render("  ")
		label := lipgloss.NewStyle().
			Background(bg).
			Foreground(lipgloss.Color(hex)).
			Render(fmt.Sprintf(" %-7s", hex))
		if i < 3 {
			swatchRow1 += swatch + label + "  "
		} else {
			swatchRow2 += swatch + label + "  "
		}
	}
	lines = append(lines, base.Render(swatchRow1))
	lines = append(lines, base.Render(swatchRow2))
	lines = append(lines, base.Render(""))

	// Fake table preview
	headerStyle := lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color(t.Accent)).Bold(true)
	selStyle := lipgloss.NewStyle().Background(lipgloss.Color(t.Accent)).Foreground(bg)

	sep := muted.Render("  " + strings.Repeat("─", min(w-4, 42)))
	lines = append(lines, headerStyle.Render("  INTERFACE   IP              TX/s        RX/s"))
	lines = append(lines, sep)

	fakeRows := []struct{ iface, ip, tx, rx string }{
		{"eth0", "192.168.1.10", "1.2 MB/s", "3.4 MB/s"},
		{"wlan0", "10.0.0.5", "512 KB/s", "1.1 MB/s"},
		{"lo", "127.0.0.1", "0 B/s", "0 B/s"},
	}
	for i, r := range fakeRows {
		row := fmt.Sprintf("  %-12s %-15s %-11s %-11s", r.iface, r.ip, r.tx, r.rx)
		if i == 0 {
			lines = append(lines, selStyle.Render(row))
		} else {
			lines = append(lines, base.Render(row))
		}
	}
	lines = append(lines, base.Render(""))

	// Color role labels
	roles := []struct {
		label string
		color string
	}{
		{"accent ", t.Accent},
		{"accent2", t.Accent2},
		{"accent3", t.Accent3},
		{"accent4", t.Accent4},
		{"accent5", t.Accent5},
		{"muted  ", t.Muted},
	}
	for _, r := range roles {
		dot := lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color(r.color)).Render("●")
		label := lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color(r.color)).
			Render(fmt.Sprintf(" %s  %s", r.label, r.color))
		lines = append(lines, base.Render("  ")+dot+label)
	}

	return strings.Join(lines, "\n")
}
