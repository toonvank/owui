package tui

import (
	"strings"
	"testing"
)

func TestRenderChatLogEmpty(t *testing.T) {
	out := renderChatLog(nil, 80, newCollapseState(), newRenderCache(), "test-model")
	if !strings.Contains(out, "Send a message") {
		t.Fatalf("expected empty state hint, got %q", out)
	}
}

func TestHighlightCodeBlocksViaPlain(t *testing.T) {
	in := "Here is code:\n```go\nfmt.Println(\"hi\")\n```\ndone"
	out := renderPlainWithFences(in, 60)
	if !strings.Contains(out, "fmt.Println") {
		t.Fatalf("expected code content preserved, got %q", out)
	}
}

func TestWrapLineDisplay(t *testing.T) {
	lines := wrapLineDisplay("one two three four five six seven", 10)
	if len(lines) < 2 {
		t.Fatalf("expected wrap into multiple lines, got %v", lines)
	}
}

func TestRenderDiffColors(t *testing.T) {
	out := renderDiff("+added\n-removed\n@@ chunk @@", 40)
	if !strings.Contains(out, "added") || !strings.Contains(out, "removed") {
		t.Fatalf("diff render missing content: %q", out)
	}
}