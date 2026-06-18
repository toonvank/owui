package repl

import (
	"strings"
	"testing"

	"github.com/toonvank/owui/internal/config"
)

func TestShortcutBarContainsEssentials(t *testing.T) {
	r := &REPL{cfg: configWithStream(false)}
	line := r.shortcutBarLine()
	for _, want := range []string{":help", ":send", "Tab", "/model", "Ctrl+C", ":collapse"} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected %q in shortcut bar, got %q", want, line)
		}
	}
}

func TestShortcutBarLoadedChat(t *testing.T) {
	r := &REPL{
		session: Session{ChatID: "55b5fc39-abcd", Title: "My test chat"},
		cfg:     configWithStream(true),
	}
	line := r.shortcutBarLine()
	if !strings.Contains(line, "My test chat") {
		t.Fatalf("expected loaded chat title in bar, got %q", line)
	}
}

func TestShortcutsPanelSections(t *testing.T) {
	panel := ShortcutsPanel()
	for _, section := range []string{"Keyboard shortcuts", "Slash commands", "Rendering"} {
		if !strings.Contains(panel, section) {
			t.Fatalf("expected section %q in panel", section)
		}
	}
}

func configWithStream(stream bool) config.Config {
	return config.Config{Stream: stream}
}