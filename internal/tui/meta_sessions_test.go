package tui

import "testing"

func TestParseSessionsCommand(t *testing.T) {
	cases := []struct {
		line   string
		active bool
		filter string
	}{
		{"/sessions", true, ""},
		{"/sessions work", true, "work"},
		{"/session load", true, ""},
		{"/session load abc", true, "abc"},
		{"/model", false, ""},
	}
	for _, tc := range cases {
		active, filter := parseSessionsCommand(tc.line)
		if active != tc.active || filter != tc.filter {
			t.Fatalf("parseSessionsCommand(%q) = (%v, %q), want (%v, %q)",
				tc.line, active, filter, tc.active, tc.filter)
		}
	}
}