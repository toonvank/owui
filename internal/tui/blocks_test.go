package tui

import (
	"strings"
	"testing"
	"time"
)

func TestParseThinkingTag(t *testing.T) {
	in := "Hello <think>secret thought</think> world"
	blocks := ParseBlocks(in)
	if len(blocks) < 2 {
		t.Fatalf("expected multiple blocks, got %d", len(blocks))
	}
	found := false
	for _, b := range blocks {
		if b.Kind == BlockThinking && strings.Contains(b.Content, "secret") {
			found = true
			if !b.Collapsed {
				t.Fatal("thinking should default collapsed")
			}
		}
	}
	if !found {
		t.Fatal("thinking block not found")
	}
}

func TestParseDiffFence(t *testing.T) {
	in := "```diff\n+added\n-removed\n```"
	blocks := ParseBlocks(in)
	if len(blocks) != 1 || blocks[0].Kind != BlockDiff {
		t.Fatalf("expected diff block, got %+v", blocks)
	}
}

func TestParseLongCodeCollapsed(t *testing.T) {
	var lines []string
	for i := 0; i < 25; i++ {
		lines = append(lines, "line")
	}
	in := "```go\n" + strings.Join(lines, "\n") + "\n```"
	blocks := ParseBlocks(in)
	if len(blocks) != 1 || !blocks[0].Collapsed {
		t.Fatalf("long code should default collapsed")
	}
}

func TestBoldMarkdownNotToolHeader(t *testing.T) {
	if isToolHeader("**tools are great**") {
		t.Fatal("bold markdown should not be a tool header")
	}
	if !isToolHeader("**Tool:** list_files") {
		t.Fatal("explicit tool header should match")
	}
}

func TestParseLargeMarkdownFast(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 500; i++ {
		b.WriteString("**item ")
		b.WriteString(strings.Repeat("x", 20))
		b.WriteString("**: some description line\n")
	}
	start := time.Now()
	blocks := ParseBlocks(b.String())
	elapsed := time.Since(start)
	if elapsed > 500*time.Millisecond {
		t.Fatalf("parse took too long: %v (%d blocks)", elapsed, len(blocks))
	}
}