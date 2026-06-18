package repl

import (
	"strings"
	"testing"
)

func TestSlashCompleterPrefix(t *testing.T) {
	r := &REPL{}
	got := r.slashSuggests("mod")
	if len(got) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(got))
	}
	for _, s := range got {
		if !strings.HasPrefix(s.Text, "/model") {
			t.Fatalf("unexpected %s", s.Text)
		}
	}
}

func TestModelMatchMarkCustom(t *testing.T) {
	r := &REPL{session: Session{Model: "smart:small"}}
	r.models.setFromEntries([]modelEntry{
		{ID: "mark-with-a-k", Name: "Mark with a K", Custom: true},
		{ID: "smart:small", Name: "smart:small", Custom: false},
		{ID: "DeepSeek-Coder:latest", Name: "DeepSeek-Coder:latest", Custom: true},
	})

	matches := r.matchModels("mark", 10)
	if len(matches) == 0 || matches[0].entry.ID != "mark-with-a-k" {
		t.Fatalf("expected mark-with-a-k first, got %v", matches)
	}
}

func TestModelMatchDeepSeek(t *testing.T) {
	r := &REPL{}
	r.models.setFromEntries([]modelEntry{
		{ID: "DeepSeek-Coder:latest", Custom: true},
		{ID: "Deepseek-Coder:latest", Custom: true},
	})

	matches := r.matchModels("DeepSeek", 10)
	if len(matches) < 2 {
		t.Fatalf("expected multiple DeepSeek models, got %d", len(matches))
	}
}

func TestModelSuggestsAtPrefix(t *testing.T) {
	r := &REPL{session: Session{Model: "smart:small"}}
	r.models.setFromEntries([]modelEntry{
		{ID: "mark-with-a-k", Name: "Mark with a K", Custom: true},
	})

	got := r.modelSuggests("mark", true)
	if len(got) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(got))
	}
	if got[0].Text != "@mark-with-a-k" {
		t.Fatalf("unexpected text %s", got[0].Text)
	}
	if !strings.Contains(got[0].Description, "custom") {
		t.Fatalf("expected custom label, got %s", got[0].Description)
	}
}