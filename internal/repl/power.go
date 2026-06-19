package repl

import (
	"fmt"
	"strings"
	"time"

	"github.com/toonvank/owui/internal/session"
)

func (r *REPL) forkSession() SlashResult {
	old := shortID(r.session.LocalID)
	r.session.LocalID = session.NewID()
	r.session.ChatID = ""
	r.session.Title = ""
	r.session.LocalTitle = "fork of " + old
	r.persistSession()
	return SlashResult{
		Output: fmt.Sprintf("forked to session %s (%d messages, new server chat)", shortID(r.session.LocalID), len(r.session.Messages)),
	}
}

func (r *REPL) deleteServerChat() SlashResult {
	if r.session.ChatID == "" {
		return SlashResult{Err: fmt.Errorf("no server chat linked to this session")}
	}
	id := r.session.ChatID
	if err := r.client.DeleteChat(id); err != nil {
		return SlashResult{Err: err}
	}
	r.session.ChatID = ""
	r.session.Title = ""
	r.persistSession()
	r.refreshChats()
	return SlashResult{Output: fmt.Sprintf("deleted server chat %s", shortID(id))}
}

func (r *REPL) slashSearch(args []string) SlashResult {
	if len(args) == 0 {
		if r.lastSearchQuery == "" {
			return SlashResult{Output: "usage: /search <term>  (/search next to cycle matches)"}
		}
		args = []string{"next"}
	}
	if len(args) == 1 && args[0] == "next" {
		if r.lastSearchQuery == "" {
			return SlashResult{Err: fmt.Errorf("no active search — use /search <term> first")}
		}
		if len(r.searchHits) == 0 {
			return SlashResult{Output: "no matches for " + r.lastSearchQuery}
		}
		r.lastSearchIdx++
		if r.lastSearchIdx >= len(r.searchHits) {
			r.lastSearchIdx = 0
		}
		idx := r.searchHits[r.lastSearchIdx]
		return SlashResult{
			Output:      fmt.Sprintf("match %d/%d (message #%d)", r.lastSearchIdx+1, len(r.searchHits), idx+1),
			ScrollToMsg: idx,
		}
	}

	query := strings.ToLower(strings.Join(args, " "))
	r.lastSearchQuery = query
	r.searchHits = nil
	for i, msg := range r.session.Messages {
		if strings.Contains(strings.ToLower(msg.Content), query) {
			r.searchHits = append(r.searchHits, i)
		}
	}
	r.lastSearchIdx = 0
	if len(r.searchHits) == 0 {
		return SlashResult{Output: "no matches for " + query}
	}
	return SlashResult{
		Output:      fmt.Sprintf("found %d match(es) for %q — /search next to cycle", len(r.searchHits), query),
		ScrollToMsg: r.searchHits[0],
	}
}

// LastUserMessage returns the most recent user message content.
func (r *REPL) LastUserMessage() string {
	for i := len(r.session.Messages) - 1; i >= 0; i-- {
		if r.session.Messages[i].Role == "user" && strings.TrimSpace(r.session.Messages[i].Content) != "" {
			return r.session.Messages[i].Content
		}
	}
	return ""
}

// RecallInput returns a previous input line for arrow-key history navigation.
func (r *REPL) RecallInput(dir int) string {
	if len(r.inputHistory) == 0 {
		return ""
	}
	if dir < 0 {
		if r.historyIdx < len(r.inputHistory) {
			r.historyIdx++
		}
		if r.historyIdx >= len(r.inputHistory) {
			r.historyIdx = len(r.inputHistory)
		}
		idx := len(r.inputHistory) - r.historyIdx
		if idx < 0 {
			return ""
		}
		return r.inputHistory[idx]
	}
	// dir > 0 = down
	if r.historyIdx > 0 {
		r.historyIdx--
	}
	if r.historyIdx == 0 {
		return ""
	}
	idx := len(r.inputHistory) - r.historyIdx
	if idx < 0 || idx >= len(r.inputHistory) {
		return ""
	}
	return r.inputHistory[idx]
}

// PushInputHistory saves a sent prompt for ↑ recall.
func (r *REPL) PushInputHistory(line string) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "/") {
		return
	}
	if len(r.inputHistory) > 0 && r.inputHistory[len(r.inputHistory)-1] == line {
		return
	}
	r.inputHistory = append(r.inputHistory, line)
	if len(r.inputHistory) > 100 {
		r.inputHistory = r.inputHistory[len(r.inputHistory)-100:]
	}
	r.historyIdx = 0
}

// LastTurnLatency returns the duration of the most recent chat turn.
func (r *REPL) LastTurnLatency() time.Duration {
	return r.lastTurnDuration
}

// VimKeysEnabled reports whether vim-style navigation is on.
func (r *REPL) VimKeysEnabled() bool {
	return r.cfg.VimKeys
}

// HistoryBrowsing reports whether the user is navigating input history.
func (r *REPL) HistoryBrowsing() bool {
	return r.historyIdx > 0
}