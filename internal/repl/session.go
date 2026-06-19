package repl

import (
	"fmt"
	"strings"

	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/session"
)

func (r *REPL) initLocalSession() {
	store, err := session.NewStore(r.cfg.ProfileName)
	if err != nil {
		r.session.LocalID = session.NewID()
		return
	}
	latest, err := store.Latest()
	if err != nil {
		r.session.LocalID = session.NewID()
		return
	}
	r.applySavedSession(latest)
}

func (r *REPL) applySavedSession(s session.Saved) {
	r.session.LocalID = s.ID
	r.session.LocalTitle = s.Title
	r.session.Messages = s.Messages
	r.session.ChatID = s.ChatID
	if s.Model != "" {
		r.session.Model = s.Model
	}
	r.session.CollectionID = s.CollectionID
	r.session.CollectionName = s.CollectionName
	r.session.AttachedFiles = append([]session.AttachedFile(nil), s.AttachedFiles...)
	r.syncFileIDs()
	r.session.ActiveFilterIDs = append([]string(nil), s.ActiveFilterIDs...)
	r.session.FiltersCustomized = s.FiltersCustomized
	r.session.ActiveToolIDs = append([]string(nil), s.ActiveToolIDs...)
	r.session.ToolsCustomized = s.ToolsCustomized
}

func (r *REPL) persistSession() {
	if r.session.LocalID == "" {
		r.session.LocalID = session.NewID()
	}
	store, err := session.NewStore(r.cfg.ProfileName)
	if err != nil {
		return
	}
	_ = store.Save(session.Saved{
		ID:             r.session.LocalID,
		Title:          r.session.LocalTitle,
		Model:          r.session.Model,
		Messages:       r.session.Messages,
		ChatID:         r.session.ChatID,
		CollectionID:   r.session.CollectionID,
		CollectionName: r.session.CollectionName,
		AttachedFiles:     append([]session.AttachedFile(nil), r.session.AttachedFiles...),
		ActiveFilterIDs:   append([]string(nil), r.session.ActiveFilterIDs...),
		FiltersCustomized: r.session.FiltersCustomized,
		ActiveToolIDs:     append([]string(nil), r.session.ActiveToolIDs...),
		ToolsCustomized:   r.session.ToolsCustomized,
	})
}

func (r *REPL) newLocalSession() {
	r.session.LocalID = session.NewID()
	r.session.LocalTitle = ""
	r.session.Messages = nil
	r.session.ChatID = ""
	r.session.Title = ""
	r.session.CollectionID = ""
	r.session.CollectionName = ""
	r.session.AttachedFiles = nil
	r.session.FileIDs = nil
	r.session.ActiveFilterIDs = nil
	r.session.FiltersCustomized = false
	r.session.ActiveToolIDs = nil
	r.session.ToolsCustomized = false
	if r.cfg.SystemPrompt != "" {
		r.session.Messages = append(r.session.Messages, api.Message{Role: "system", Content: r.cfg.SystemPrompt})
	}
	r.persistSession()
}

func (r *REPL) formatLocalSessions() string {
	store, err := session.NewStore(r.cfg.ProfileName)
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
	store, err := session.NewStore(r.cfg.ProfileName)
	if err != nil {
		return err
	}
	all, err := store.List()
	if err != nil {
		return err
	}
	for _, s := range all {
		if s.ID == id || strings.HasPrefix(s.ID, id) {
			r.applySavedSession(s)
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