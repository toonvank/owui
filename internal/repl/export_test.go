package repl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/config"
)

func TestRegenPrompt(t *testing.T) {
	r := &REPL{
		cfg: config.Default(),
		session: Session{
			LocalID: "test-session",
			Model:   "llama",
			Messages: []api.Message{
				{Role: "user", Content: "hello"},
				{Role: "assistant", Content: "hi there"},
			},
		},
	}

	prompt, err := r.regenPrompt()
	if err != nil {
		t.Fatal(err)
	}
	if prompt != "hello" {
		t.Fatalf("expected hello, got %q", prompt)
	}
	if len(r.session.Messages) != 0 {
		t.Fatalf("expected messages cleared, got %d", len(r.session.Messages))
	}
}

func TestExportSessionMarkdown(t *testing.T) {
	dir := t.TempDir()
	r := &REPL{
		cfg: config.Config{BaseURL: "http://localhost:3000"},
		session: Session{
			LocalID:    "abc123",
			LocalTitle: "Test chat",
			Model:      "llama",
			Messages: []api.Message{
				{Role: "user", Content: "ping"},
				{Role: "assistant", Content: "pong"},
			},
		},
	}

	path := filepath.Join(dir, "out.md")
	written, err := r.exportSession("md", path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(written)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"# Test chat", "## USER", "ping", "## ASSISTANT", "pong"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in export, got:\n%s", want, text)
		}
	}
}