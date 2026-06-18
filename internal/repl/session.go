package repl

import (
	"fmt"
	"strings"

	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/session"
)

func (r *REPL) initLocalSession() {
	store, err := session.NewStore()
	if err != nil {
		r.session.LocalID = session.NewID()
		return
	}
	latest, err := store.Latest()
	if err != nil {
		r.session.LocalID = session.NewID()
		return
	}
	r.session.LocalID = latest.ID
	r.session.LocalTitle = latest.Title
	r.session.Messages = latest.Messages
	r.session.ChatID = latest.ChatID
	if latest.Model != "" {
		r.session.Model = latest.Model
	}
}

func (r *REPL) persistSession() {
	if r.session.LocalID == "" {
		r.session.LocalID = session.NewID()
	}
	store, err := session.NewStore()
	if err != nil {
		return
	}
	_ = store.Save(session.Saved{
		ID:       r.session.LocalID,
		Title:    r.session.LocalTitle,
		Model:    r.session.Model,
		Messages: r.session.Messages,
		ChatID:   r.session.ChatID,
	})
}

func (r *REPL) newLocalSession() {
	r.session.LocalID = session.NewID()
	r.session.LocalTitle = ""
	r.session.Messages = nil
	r.session.ChatID = ""
	r.session.Title = ""
	if r.cfg.SystemPrompt != "" {
		r.session.Messages = append(r.session.Messages, api.Message{Role: "system", Content: r.cfg.SystemPrompt})
	}
	r.persistSession()
}

func (r *REPL) formatLocalSessions() string {
	store, err := session.NewStore()
	if err != nil {
		return "error: " + err.Error()
	}
	all, err := store.List()
	if err != nil {
		return "error: " + err.Error()
	}
	if len(all) == 0 {
		return "no local sessions"
	}
	limit := 15
	var b strings.Builder
	for i, s := range all {
		if i >= limit {
			fmt.Fprintf(&b, "showing %d sessions — /session load <id>", limit)
			break
		}
		title := s.Title
		if title == "" {
			title = "untitled"
		}
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		marker := " "
		if s.ID == r.session.LocalID {
			marker = "*"
		}
		short := s.ID
		if len(short) > 12 {
			short = short[:12]
		}
		fmt.Fprintf(&b, "%s %s  %s  (%d msgs)\n", marker, short, title, len(s.Messages))
	}
	return strings.TrimRight(b.String(), "\n")
}

func (r *REPL) loadLocalSession(id string) error {
	store, err := session.NewStore()
	if err != nil {
		return err
	}
	all, err := store.List()
	if err != nil {
		return err
	}
	for _, s := range all {
		if s.ID == id || strings.HasPrefix(s.ID, id) {
			r.session.LocalID = s.ID
			r.session.LocalTitle = s.Title
			r.session.Messages = s.Messages
			r.session.ChatID = s.ChatID
			if s.Model != "" {
				r.session.Model = s.Model
			}
			return nil
		}
	}
	return fmt.Errorf("no local session matching %q", id)
}

func shortID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

// LocalSessionLabel returns a short label for the TUI header.
func (r *REPL) LocalSessionLabel() string {
	title := r.session.Title
	if title == "" {
		title = r.session.LocalTitle
	}
	if title != "" {
		t := title
		if len(t) > 24 {
			return t[:21] + "..."
		}
		return t
	}
	if r.session.LocalID != "" && len(r.session.LocalID) >= 8 {
		return r.session.LocalID[:8]
	}
	return ""
}