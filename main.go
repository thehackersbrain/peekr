package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/thehackersbrain/peekr/internal/theme"
	"github.com/thehackersbrain/peekr/internal/ui"
)

const version = "1.0.0"

func main() {
	var (
		themeName string
		interval  int
		showVer   bool
	)

	flag.StringVar(&themeName, "t", "tokyo-night", "theme name (e.g. nord, dracula, gruvbox)")
	flag.StringVar(&themeName, "theme", "tokyo-night", "theme name")
	flag.IntVar(&interval, "d", 3, "refresh interval in seconds")
	flag.BoolVar(&showVer, "v", false, "show version")
	flag.Parse()

	if showVer {
		fmt.Printf("peekr %s\n", version)
		os.Exit(0)
	}

	t := theme.ByName(themeName)
	m := ui.New(t, time.Duration(interval)*time.Second)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
