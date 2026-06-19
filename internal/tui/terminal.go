package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/toonvank/owui/internal/config"
)

// initTerminal configures lipgloss for the TUI display environment.
func initTerminal(cfg config.Config) {
	ApplyDisplayOptions(cfg)
}

// ApplyDisplayOptions sets color profile and background from config and NO_COLOR.
func ApplyDisplayOptions(cfg config.Config) {
	if os.Getenv("NO_COLOR") != "" || cfg.Theme == "none" {
		lipgloss.SetColorProfile(termenv.Ascii)
		return
	}
	lipgloss.SetColorProfile(termenv.TrueColor)
	switch cfg.Theme {
	case "light":
		lipgloss.SetHasDarkBackground(false)
	default:
		lipgloss.SetHasDarkBackground(true)
	}
}