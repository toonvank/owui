package tui

import "testing"

func TestParseModelCommand(t *testing.T) {
	cases := []struct {
		line   string
		active bool
		filter string
	}{
		{"/model", true, ""},
		{"/model sonnet", true, "sonnet"},
		{"/model  mark-with-a-k", true, "mark-with-a-k"},
		{"/models", false, ""},
		{"/models llama", false, ""},
		{"/mode", false, ""},
		{"hello", false, ""},
	}
	for _, tc := range cases {
		active, filter := parseModelCommand(tc.line)
		if active != tc.active || filter != tc.filter {
			t.Fatalf("parseModelCommand(%q) = (%v, %q), want (%v, %q)",
				tc.line, active, filter, tc.active, tc.filter)
		}
	}
}

func TestParseChatsCommand(t *testing.T) {
	cases := []struct {
		line   string
		active bool
		filter string
	}{
		{"/chats", true, ""},
		{"/chats latency", true, "latency"},
		{"/resume", true, ""},
		{"/resume abc", true, "abc"},
		{"/load", true, ""},
		{"/load 55b5fc39", true, "55b5fc39"},
		{"/chat", false, ""},
	}
	for _, tc := range cases {
		active, filter := parseChatsCommand(tc.line)
		if active != tc.active || filter != tc.filter {
			t.Fatalf("parseChatsCommand(%q) = (%v, %q), want (%v, %q)",
				tc.line, active, filter, tc.active, tc.filter)
		}
	}
}