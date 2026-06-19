package repl

import (
	"testing"

	"github.com/toonvank/owui/internal/api"
)

func TestSlashSearch(t *testing.T) {
	r := &REPL{
		session: Session{
			Messages: []api.Message{
				{Role: "user", Content: "hello world"},
				{Role: "assistant", Content: "hi there"},
				{Role: "user", Content: "goodbye world"},
			},
		},
	}
	res := r.slashSearch([]string{"world"})
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if len(r.searchHits) != 2 {
		t.Fatalf("expected 2 hits, got %v", r.searchHits)
	}
	if res.ScrollToMsg != 0 {
		t.Fatalf("scroll to %d", res.ScrollToMsg)
	}
}

func TestForkSession(t *testing.T) {
	r := &REPL{
		session: Session{
			LocalID:  "old123",
			ChatID:   "chat-abc",
			Messages: []api.Message{{Role: "user", Content: "hi"}},
		},
	}
	res := r.forkSession()
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if r.session.ChatID != "" {
		t.Fatal("expected chat id cleared")
	}
	if r.session.LocalID == "old123" {
		t.Fatal("expected new local id")
	}
}

func TestInputHistory(t *testing.T) {
	r := &REPL{}
	r.PushInputHistory("first")
	r.PushInputHistory("second")
	if r.RecallInput(-1) != "second" {
		t.Fatal("expected second")
	}
	if r.RecallInput(-1) != "first" {
		t.Fatal("expected first")
	}
}