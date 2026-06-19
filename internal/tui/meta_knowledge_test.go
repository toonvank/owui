package tui

import "testing"

func TestParseKnowledgeCommand(t *testing.T) {
	cases := []struct {
		line   string
		active bool
		filter string
	}{
		{"/knowledge", true, ""},
		{"/knowledge docs", true, "docs"},
		{"/kb", true, ""},
		{"/kb work", true, "work"},
		{"/file", false, ""},
	}
	for _, tc := range cases {
		active, filter := parseKnowledgeCommand(tc.line)
		if active != tc.active || filter != tc.filter {
			t.Fatalf("parseKnowledgeCommand(%q) = (%v, %q), want (%v, %q)",
				tc.line, active, filter, tc.active, tc.filter)
		}
	}
}