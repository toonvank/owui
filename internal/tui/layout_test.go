package tui

import "testing"

func TestAlignRowPinsRight(t *testing.T) {
	row := alignRow("left", "[done]", 40)
	if len(row) < 40 {
		t.Fatalf("expected padded row, got %q", row)
	}
	if row[len(row)-6:] != "[done]" {
		t.Fatalf("right tag not pinned: %q", row)
	}
}

func TestFormatCount(t *testing.T) {
	if formatCount(412) != "412" {
		t.Fatalf("got %s", formatCount(412))
	}
	if formatCount(39800) != "39.8k" {
		t.Fatalf("got %s", formatCount(39800))
	}
}