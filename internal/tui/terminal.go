package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// initTerminal disables OSC color queries that leak into stdin as garbage text
// (e.g. "11;rgb:1a1a1a") and break the input field.
func initTerminal() {
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
}