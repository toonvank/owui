package tui

import "testing"

func TestParseFiltersCommand(t *testing.T) {
	active, filter := parseFiltersCommand("/filters web")
	if !active || filter != "web" {
		t.Fatalf("got (%v, %q)", active, filter)
	}
	active, _ = parseFiltersCommand("/functions")
	if !active {
		t.Fatal("expected /functions active")
	}
}

func TestParseToolsCommand(t *testing.T) {
	active, filter := parseToolsCommand("/tools search")
	if !active || filter != "search" {
		t.Fatalf("got (%v, %q)", active, filter)
	}
}