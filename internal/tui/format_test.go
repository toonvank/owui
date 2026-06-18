package tui

import "testing"

func TestPreservedMessagesPhrase(t *testing.T) {
	if got := preservedMessagesPhrase(1); got != "1 message kept" {
		t.Fatalf("got %q", got)
	}
	if got := preservedMessagesPhrase(3); got != "3 messages kept" {
		t.Fatalf("got %q", got)
	}
}