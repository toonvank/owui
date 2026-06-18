package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/toonvank/owui/internal/repl"
)

// Run starts the full-screen interactive TUI for the given REPL session.
func Run(r *repl.REPL) error {
	initTerminal()

	m := New(r)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}