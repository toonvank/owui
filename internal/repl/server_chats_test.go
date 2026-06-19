package repl

import (
	"testing"

	"github.com/toonvank/owui/internal/api"
)

func TestNeedsAutoTitle(t *testing.T) {
	for _, title := range []string{"", "New Chat", "UNTITLED", "chat"} {
		if !needsAutoTitle(title) {
			t.Fatalf("expected generic title %q", title)
		}
	}
	if needsAutoTitle("My project notes") {
		t.Fatal("expected custom title to skip auto-title")
	}
}

func TestCountUserMessages(t *testing.T) {
	msgs := []api.Message{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "hello"},
		{Role: "user", Content: "   "},
		{Role: "user", Content: "again"},
	}
	if countUserMessages(msgs) != 2 {
		t.Fatalf("got %d", countUserMessages(msgs))
	}
}

func TestSortChatsPinnedFirst(t *testing.T) {
	chats := []api.ChatSummary{
		{ID: "a", Title: "recent", UpdatedAt: 300},
		{ID: "b", Title: "pinned old", Pinned: true, UpdatedAt: 100},
		{ID: "c", Title: "middle", UpdatedAt: 200},
	}
	sortChatsByPinnedRecency(chats)
	if chats[0].ID != "b" {
		t.Fatalf("pinned chat should be first: %+v", chats)
	}
}