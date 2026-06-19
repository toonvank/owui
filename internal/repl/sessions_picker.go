package repl

import (
	"fmt"
	"strings"
	"time"

	"github.com/toonvank/owui/internal/session"
)

// SessionPick is one entry in the interactive /sessions picker.
type SessionPick struct {
	ID       string
	Title    string
	ShortID  string
	MsgCount int
	Age      string
	Current  bool
}

func formatSessionAge(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}

func sessionMatchScore(query string, s session.Saved) int {
	if query == "" {
		return 1
	}
	q := strings.ToLower(query)
	id := strings.ToLower(s.ID)
	title := strings.ToLower(s.Title)

	if strings.HasPrefix(id, q) {
		return 800 + len(q)
	}
	if strings.Contains(title, q) {
		return 500 + len(q)
	}
	if strings.Contains(id, q) {
		return 300 + len(q)
	}
	return 0
}

// SearchLocalSessions returns local sessions matching a fuzzy query.
func (r *REPL) SearchLocalSessions(query string, limit int) []SessionPick {
	if limit <= 0 {
		limit = 50
	}
	store, err := session.NewStore(r.cfg.ProfileName)
	if err != nil {
		return nil
	}
	all, err := store.List()
	if err != nil || len(all) == 0 {
		return nil
	}

	type scored struct {
		s     session.Saved
		score int
	}
	scoredList := make([]scored, 0, len(all))
	for _, s := range all {
		score := sessionMatchScore(query, s)
		if score <= 0 {
			continue
		}
		scoredList = append(scoredList, scored{s: s, score: score})
	}
	if len(scoredList) == 0 && query == "" {
		for _, s := range all {
			scoredList = append(scoredList, scored{s: s, score: 1})
		}
	}
	if len(scoredList) > limit {
		scoredList = scoredList[:limit]
	}

	out := make([]SessionPick, 0, len(scoredList))
	for _, sc := range scoredList {
		title := sc.s.Title
		if title == "" {
			title = "untitled"
		}
		short := sc.s.ID
		if len(short) > 12 {
			short = short[:12]
		}
		out = append(out, SessionPick{
			ID:       sc.s.ID,
			Title:    title,
			ShortID:  short,
			MsgCount: len(sc.s.Messages),
			Age:      formatSessionAge(sc.s.UpdatedAt),
			Current:  sc.s.ID == r.session.LocalID,
		})
	}
	return out
}

// LoadLocalSessionByID loads a local session into the current REPL session.
func (r *REPL) LoadLocalSessionByID(id string) SlashResult {
	if err := r.loadLocalSession(id); err != nil {
		return SlashResult{Err: err}
	}
	return SlashResult{
		Output:         fmt.Sprintf("loaded local session %s (%d messages)", shortID(r.session.LocalID), len(r.session.Messages)),
		ReloadMessages: true,
	}
}