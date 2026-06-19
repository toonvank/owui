package repl

import (
	"testing"

	"github.com/toonvank/owui/internal/config"
)

func TestToggleFilter(t *testing.T) {
	r := &REPL{
		cfg: config.Default(),
		session: Session{
			Model:             "llama",
			FiltersCustomized: true,
			ActiveFilterIDs:   []string{"f1"},
		},
	}
	if !r.isFilterEnabled("f1") {
		t.Fatal("expected f1 enabled")
	}
	r.ToggleFilter("f1")
	if r.isFilterEnabled("f1") {
		t.Fatal("expected f1 disabled after toggle")
	}
	r.ToggleFilter("f2")
	if !r.isFilterEnabled("f2") {
		t.Fatal("expected f2 enabled after toggle")
	}
}

func TestChatOptionsExplicitFilters(t *testing.T) {
	r := &REPL{
		session: Session{
			FiltersCustomized: true,
			ActiveFilterIDs:   []string{"a", "b"},
		},
	}
	opts := r.chatOptions()
	if opts == nil || !opts.ExplicitFilters {
		t.Fatal("expected explicit filters")
	}
	if len(opts.FilterIDs) != 2 {
		t.Fatalf("filter ids = %v", opts.FilterIDs)
	}
}

func TestToggleID(t *testing.T) {
	next, on := toggleID([]string{"a"}, "b")
	if !on || len(next) != 2 {
		t.Fatalf("add failed: %v %v", next, on)
	}
	next, on = toggleID(next, "a")
	if on || len(next) != 1 || next[0] != "b" {
		t.Fatalf("remove failed: %v %v", next, on)
	}
}