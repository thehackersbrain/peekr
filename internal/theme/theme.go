// Package theme provides compiled-in color themes for peekr.
// All themes are pure Go — no runtime files needed.
package theme

import "github.com/charmbracelet/lipgloss"

// Theme holds a complete color palette.
type Theme struct {
	Name    string
	BG      string
	FG      string
	Accent  string // primary highlight (blue-ish)
	Accent2 string // red / error
	Accent3 string // green / ok
	Accent4 string // cyan
	Accent5 string // purple / pink
	Muted   string // dimmed text
	// Palette is the ordered list for swatch display (6 colors)
	Palette [6]string
}

// Styles derived from a Theme, ready for lipgloss rendering.
type Styles struct {
	T *Theme

	Base      lipgloss.Style
	Header    lipgloss.Style
	Selected  lipgloss.Style
	Muted     lipgloss.Style
	Accent    lipgloss.Style
	Accent2   lipgloss.Style
	Accent3   lipgloss.Style
	Accent4   lipgloss.Style
	Accent5   lipgloss.Style
	Border    lipgloss.Style
	Title     lipgloss.Style
	Separator lipgloss.Style
}

// Build creates a Styles set from a Theme.
func (t *Theme) Build() Styles {
	bg := lipgloss.Color(t.BG)
	fg := lipgloss.Color(t.FG)
	ac := lipgloss.Color(t.Accent)
	ac2 := lipgloss.Color(t.Accent2)
	ac3 := lipgloss.Color(t.Accent3)
	ac4 := lipgloss.Color(t.Accent4)
	ac5 := lipgloss.Color(t.Accent5)
	mu := lipgloss.Color(t.Muted)

	base := lipgloss.NewStyle().Background(bg).Foreground(fg)

	return Styles{
		T:         t,
		Base:      base,
		Header:    lipgloss.NewStyle().Background(bg).Foreground(ac).Bold(true),
		Selected:  lipgloss.NewStyle().Background(lipgloss.Color(t.Accent)).Foreground(bg).Bold(true),
		Muted:     lipgloss.NewStyle().Background(bg).Foreground(mu),
		Accent:    lipgloss.NewStyle().Background(bg).Foreground(ac),
		Accent2:   lipgloss.NewStyle().Background(bg).Foreground(ac2),
		Accent3:   lipgloss.NewStyle().Background(bg).Foreground(ac3),
		Accent4:   lipgloss.NewStyle().Background(bg).Foreground(ac4),
		Accent5:   lipgloss.NewStyle().Background(bg).Foreground(ac5),
		Border:    lipgloss.NewStyle().Background(bg).Foreground(mu),
		Title:     lipgloss.NewStyle().Background(bg).Foreground(ac).Bold(true),
		Separator: lipgloss.NewStyle().Background(bg).Foreground(mu),
	}
}

// ---------------------------------------------------------------------------
// All built-in themes
// ---------------------------------------------------------------------------

var TokyoNight = &Theme{
	Name:    "tokyo-night",
	BG:      "#1a1b26",
	FG:      "#c0caf5",
	Accent:  "#7aa2f7",
	Accent2: "#f7768e",
	Accent3: "#9ece6a",
	Accent4: "#7dcfff",
	Accent5: "#bb9af7",
	Muted:   "#414868",
	Palette: [6]string{"#7aa2f7", "#f7768e", "#9ece6a", "#7dcfff", "#bb9af7", "#c0caf5"},
}

var Nord = &Theme{
	Name:    "nord",
	BG:      "#2e3440",
	FG:      "#eceff4",
	Accent:  "#88c0d0",
	Accent2: "#bf616a",
	Accent3: "#a3be8c",
	Accent4: "#81a1c1",
	Accent5: "#b48ead",
	Muted:   "#4c566a",
	Palette: [6]string{"#88c0d0", "#bf616a", "#a3be8c", "#81a1c1", "#b48ead", "#eceff4"},
}

var Dracula = &Theme{
	Name:    "dracula",
	BG:      "#282a36",
	FG:      "#f8f8f2",
	Accent:  "#8be9fd",
	Accent2: "#ff5555",
	Accent3: "#50fa7b",
	Accent4: "#8be9fd",
	Accent5: "#bd93f9",
	Muted:   "#6272a4",
	Palette: [6]string{"#8be9fd", "#50fa7b", "#ff5555", "#ffb86c", "#ff79c6", "#bd93f9"},
}

var Gruvbox = &Theme{
	Name:    "gruvbox",
	BG:      "#282828",
	FG:      "#ebdbb2",
	Accent:  "#fabd2f",
	Accent2: "#fb4934",
	Accent3: "#b8bb26",
	Accent4: "#83a598",
	Accent5: "#d3869b",
	Muted:   "#928374",
	Palette: [6]string{"#fabd2f", "#fb4934", "#b8bb26", "#83a598", "#d3869b", "#ebdbb2"},
}

var CatppuccinFrappe = &Theme{
	Name:    "catppuccin-frappe",
	BG:      "#303446",
	FG:      "#c6d0f5",
	Accent:  "#8caaee",
	Accent2: "#e78284",
	Accent3: "#a6d189",
	Accent4: "#ca9ee6",
	Accent5: "#f4b8e4",
	Muted:   "#626880",
	Palette: [6]string{"#8caaee", "#e78284", "#a6d189", "#ca9ee6", "#f4b8e4", "#c6d0f5"},
}

var EverforestDark = &Theme{
	Name:    "everforest-dark",
	BG:      "#2d353b",
	FG:      "#d3c6aa",
	Accent:  "#a7c080",
	Accent2: "#e67e80",
	Accent3: "#dbbc7f",
	Accent4: "#7fbbb3",
	Accent5: "#d699b6",
	Muted:   "#859289",
	Palette: [6]string{"#a7c080", "#e67e80", "#dbbc7f", "#7fbbb3", "#d699b6", "#d3c6aa"},
}

var EverforestLight = &Theme{
	Name:    "everforest-light",
	BG:      "#fdf6e3",
	FG:      "#5c6a72",
	Accent:  "#8da101",
	Accent2: "#f85552",
	Accent3: "#dfa000",
	Accent4: "#3a94c5",
	Accent5: "#df69ba",
	Muted:   "#939f91",
	Palette: [6]string{"#8da101", "#f85552", "#dfa000", "#3a94c5", "#df69ba", "#5c6a72"},
}

var Kanagawa = &Theme{
	Name:    "kanagawa",
	BG:      "#1f1f28",
	FG:      "#dcd7ba",
	Accent:  "#7e9cd8",
	Accent2: "#e82424",
	Accent3: "#98bb6c",
	Accent4: "#7fb4ca",
	Accent5: "#957fb8",
	Muted:   "#727169",
	Palette: [6]string{"#7e9cd8", "#e82424", "#98bb6c", "#7fb4ca", "#957fb8", "#dcd7ba"},
}

var Moonlight = &Theme{
	Name:    "moonlight",
	BG:      "#222436",
	FG:      "#c8d3f5",
	Accent:  "#82aaff",
	Accent2: "#ff757f",
	Accent3: "#c3e88d",
	Accent4: "#86e1fc",
	Accent5: "#c099ff",
	Muted:   "#444a73",
	Palette: [6]string{"#82aaff", "#ff757f", "#c3e88d", "#86e1fc", "#c099ff", "#c8d3f5"},
}

var MonokaiPro = &Theme{
	Name:    "monokai-pro",
	BG:      "#2d2a2e",
	FG:      "#fcfcfa",
	Accent:  "#ffd866",
	Accent2: "#ff6188",
	Accent3: "#a9dc76",
	Accent4: "#78dce8",
	Accent5: "#ab9df2",
	Muted:   "#727072",
	Palette: [6]string{"#ffd866", "#ff6188", "#a9dc76", "#78dce8", "#ab9df2", "#fcfcfa"},
}

var Nightfox = &Theme{
	Name:    "nightfox",
	BG:      "#192330",
	FG:      "#cdcecf",
	Accent:  "#719cd6",
	Accent2: "#c94f6d",
	Accent3: "#81b29a",
	Accent4: "#63cdcf",
	Accent5: "#9d79d6",
	Muted:   "#3b4261",
	Palette: [6]string{"#c94f6d", "#81b29a", "#dbc074", "#719cd6", "#9d79d6", "#63cdcf"},
}

var Oxocarbon = &Theme{
	Name:    "oxocarbon",
	BG:      "#161616",
	FG:      "#f2f4f8",
	Accent:  "#33b1ff",
	Accent2: "#ee5396",
	Accent3: "#08bdba",
	Accent4: "#82cfff",
	Accent5: "#be95ff",
	Muted:   "#393939",
	Palette: [6]string{"#ff7eb6", "#08bdba", "#33b1ff", "#ee5396", "#be95ff", "#82cfff"},
}

var ZenbonesDark = &Theme{
	Name:    "zenbones-dark",
	BG:      "#1c1917",
	FG:      "#b4bdc3",
	Accent:  "#6099c0",
	Accent2: "#de6e7c",
	Accent3: "#819b69",
	Accent4: "#66a5ad",
	Accent5: "#b279a7",
	Muted:   "#53565a",
	Palette: [6]string{"#6099c0", "#de6e7c", "#819b69", "#66a5ad", "#b279a7", "#b4bdc3"},
}

var ZenbonesLight = &Theme{
	Name:    "zenbones-light",
	BG:      "#f0edec",
	FG:      "#2c363c",
	Accent:  "#286486",
	Accent2: "#a8334c",
	Accent3: "#4f6c31",
	Accent4: "#3b8992",
	Accent5: "#88507d",
	Muted:   "#b0ada2",
	Palette: [6]string{"#286486", "#a8334c", "#4f6c31", "#3b8992", "#88507d", "#2c363c"},
}

var Cthulhain = &Theme{
	Name:    "cthulhain",
	BG:      "#162737",
	FG:      "#9a9a9a",
	Accent:  "#96bd64",
	Accent2: "#685c92",
	Accent3: "#5c9291",
	Accent4: "#5b6c86",
	Accent5: "#4a5b65",
	Muted:   "#3b4a55",
	Palette: [6]string{"#685c92", "#5c9291", "#4a5b65", "#5b6c86", "#96bd64", "#9a9a9a"},
}

// All is the ordered list of all built-in themes.
var All = []*Theme{
	TokyoNight,
	Nord,
	Dracula,
	Gruvbox,
	CatppuccinFrappe,
	EverforestDark,
	EverforestLight,
	Kanagawa,
	Moonlight,
	MonokaiPro,
	Nightfox,
	Oxocarbon,
	ZenbonesDark,
	ZenbonesLight,
	Cthulhain,
}

// Default is the default theme.
var Default = TokyoNight

// ByName returns a theme by name, falling back to Default.
func ByName(name string) *Theme {
	for _, t := range All {
		if t.Name == name {
			return t
		}
	}
	return Default
}
