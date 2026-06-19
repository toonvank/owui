package tui

import "testing"

func TestParseProfileCommand(t *testing.T) {
	active, filter := parseProfileCommand("/profile work")
	if !active || filter != "work" {
		t.Fatalf("got (%v, %q)", active, filter)
	}
	active, _ = parseProfileCommand("/profile")
	if !active {
		t.Fatal("expected /profile active")
	}
}