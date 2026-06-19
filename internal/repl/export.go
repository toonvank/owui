package repl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/toonvank/owui/internal/api"
)

func (r *REPL) exportSession(format, path string) (string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "md"
	}

	var data []byte
	var ext string
	switch format {
	case "md", "markdown":
		data = []byte(r.exportMarkdown())
		ext = ".md"
	case "json":
		payload := struct {
			ID             string        `json:"id"`
			Title          string        `json:"title"`
			Model          string        `json:"model"`
			ChatID         string        `json:"chat_id,omitempty"`
			CollectionID   string        `json:"collection_id,omitempty"`
			CollectionName string        `json:"collection_name,omitempty"`
			AttachedFiles  any           `json:"attached_files,omitempty"`
			Messages       []api.Message `json:"messages"`
			Exported       time.Time     `json:"exported_at"`
			ServerURL      string        `json:"server_url"`
		}{
			ID:             r.session.LocalID,
			Title:          r.session.LocalTitle,
			Model:          r.session.Model,
			ChatID:         r.session.ChatID,
			CollectionID:   r.session.CollectionID,
			CollectionName: r.session.CollectionName,
			AttachedFiles:  r.session.AttachedFiles,
			Messages:       r.session.Messages,
			Exported:       time.Now().UTC(),
			ServerURL:      r.cfg.BaseURL,
		}
		var err error
		data, err = json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return "", err
		}
		ext = ".json"
	default:
		return "", fmt.Errorf("unknown format %q (use md or json)", format)
	}

	if path == "" {
		path = fmt.Sprintf("owui-export-%s%s", shortID(r.session.LocalID), ext)
	}
	path = filepath.Clean(path)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	abs, _ := filepath.Abs(path)
	return abs, nil
}

func (r *REPL) exportMarkdown() string {
	var b strings.Builder
	title := r.session.LocalTitle
	if title == "" {
		title = r.session.Title
	}
	if title == "" {
		title = "owui export"
	}
	fmt.Fprintf(&b, "# %s\n\n", title)
	fmt.Fprintf(&b, "_model: %s · exported from %s_\n\n", r.session.Model, r.cfg.BaseURL)
	for _, msg := range r.session.Messages {
		role := msg.Role
		if role == "" {
			role = "unknown"
		}
		fmt.Fprintf(&b, "## %s\n\n%s\n\n", strings.ToUpper(role), msg.Content)
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func (r *REPL) copyLastAssistant() (string, error) {
	content := r.lastAssistantContent()
	if content == "" {
		return "", fmt.Errorf("no assistant reply to copy")
	}
	if err := writeClipboard(content); err != nil {
		return "", err
	}
	n := len(content)
	if n > 80 {
		n = 80
	}
	return fmt.Sprintf("copied %d chars to clipboard", len(content)), nil
}

func (r *REPL) lastAssistantContent() string {
	for i := len(r.session.Messages) - 1; i >= 0; i-- {
		if r.session.Messages[i].Role == "assistant" && strings.TrimSpace(r.session.Messages[i].Content) != "" {
			return r.session.Messages[i].Content
		}
	}
	return ""
}

func (r *REPL) regenPrompt() (string, error) {
	msgs := r.session.Messages
	if len(msgs) == 0 {
		return "", fmt.Errorf("no messages to regenerate")
	}

	// Drop trailing assistant turn if present.
	if msgs[len(msgs)-1].Role == "assistant" {
		msgs = msgs[:len(msgs)-1]
	}
	if len(msgs) == 0 || msgs[len(msgs)-1].Role != "user" {
		return "", fmt.Errorf("no user message to regenerate from")
	}

	prompt := msgs[len(msgs)-1].Content
	r.session.Messages = msgs[:len(msgs)-1]
	r.persistSession()
	return prompt, nil
}

func (r *REPL) setSystemPrompt(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		for _, msg := range r.session.Messages {
			if msg.Role == "system" {
				return msg.Content
			}
		}
		return "(no system prompt)"
	}

	found := false
	for i, msg := range r.session.Messages {
		if msg.Role == "system" {
			r.session.Messages[i].Content = text
			found = true
			break
		}
	}
	if !found {
		r.session.Messages = append([]api.Message{{Role: "system", Content: text}}, r.session.Messages...)
	}
	r.persistSession()
	return "system prompt updated"
}

func (r *REPL) setSessionTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		if r.session.LocalTitle != "" {
			return r.session.LocalTitle
		}
		if r.session.Title != "" {
			return r.session.Title
		}
		return "(untitled)"
	}
	r.session.LocalTitle = title
	r.persistSession()
	return "title set to " + title
}