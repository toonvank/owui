package repl

import (
	"testing"

	"github.com/toonvank/owui/internal/api"
)

func TestSortChatsByRecency(t *testing.T) {
	chats := []api.ChatSummary{
		{ID: "a", Title: "older", UpdatedAt: 100},
		{ID: "b", Title: "newest", UpdatedAt: 300},
		{ID: "c", Title: "middle", UpdatedAt: 200},
	}
	sortChatsByRecency(chats)
	if chats[0].ID != "b" || chats[1].ID != "c" || chats[2].ID != "a" {
		t.Fatalf("unexpected order: %+v", chats)
	}
}

func TestChatMatchScore(t *testing.T) {
	ch := api.ChatSummary{ID: "55b5fc39-abcd-1234", Title: "p99 latency regression"}
	if chatMatchScore("55b5", ch) <= 0 {
		t.Fatal("expected id prefix match")
	}
	if chatMatchScore("latency", ch) <= 0 {
		t.Fatal("expected title match")
	}
	if chatMatchScore("nomatch", ch) != 0 {
		t.Fatal("expected no match")
	}
}