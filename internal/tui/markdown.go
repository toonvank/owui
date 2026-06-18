package tui

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

var (
	glamourMu       sync.Mutex
	glamourRenderer *glamour.TermRenderer
	glamourWidth    int
)

// PrewarmMarkdown initializes the glamour renderer in the background.
func PrewarmMarkdown(width int) {
	go func() {
		_, _ = glamourRendererFor(width)
	}()
}

func renderMarkdown(text string, width int) string {
	if width < 24 {
		width = 24
	}
	if !needsRichMarkdown(text) {
		return wrapContentDisplay(text, width)
	}
	r, err := glamourRendererFor(width)
	if err != nil {
		return wrapContentDisplay(text, width)
	}
	out, err := r.Render(text)
	if err != nil {
		return wrapContentDisplay(text, width)
	}
	return strings.TrimSpace(out)
}

func glamourRendererFor(width int) (*glamour.TermRenderer, error) {
	glamourMu.Lock()
	defer glamourMu.Unlock()

	if glamourRenderer != nil && glamourWidth == width {
		return glamourRenderer, nil
	}
	// Use explicit dark style — WithAutoStyle() queries the terminal (OSC 11)
	// which leaks color responses into the input buffer.
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}
	glamourRenderer = r
	glamourWidth = width
	return r, nil
}