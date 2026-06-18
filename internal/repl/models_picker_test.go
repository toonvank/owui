package repl

import (
	"testing"

	"github.com/toonvank/owui/internal/api"
)

func TestSearchModelsFuzzy(t *testing.T) {
	r := &REPL{session: Session{Model: "smart:small"}}
	r.models.setFromEntries([]modelEntry{
		{ID: "mark-with-a-k", Name: "Mark with a K", Custom: true},
		{ID: "smart:small", Name: "smart:small"},
		{ID: "claude-sonnet-4", Name: "Claude Sonnet 4"},
	})

	matches := r.SearchModels("sonnet", 10)
	if len(matches) != 1 || matches[0].ID != "claude-sonnet-4" {
		t.Fatalf("expected sonnet match, got %+v", matches)
	}

	all := r.SearchModels("", 10)
	if len(all) < 3 {
		t.Fatalf("expected all models, got %d", len(all))
	}
}

func TestSetModelIDPreservesSession(t *testing.T) {
	r := &REPL{
		session: Session{
			Model:    "a",
			Messages: []api.Message{{Role: "user", Content: "hi"}},
			LocalID:  "sess-1",
		},
	}
	r.SetModelID("b")
	if r.session.Model != "b" {
		t.Fatal("model not updated")
	}
	if len(r.session.Messages) != 1 {
		t.Fatal("messages should be preserved")
	}
}