package repl

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/session"
)

var genericChatTitles = map[string]bool{
	"":           true,
	"new chat":   true,
	"untitled":   true,
	"chat":       true,
	"new":        true,
}

func needsAutoTitle(title string) bool {
	return genericChatTitles[strings.ToLower(strings.TrimSpace(title))]
}

func countUserMessages(messages []api.Message) int {
	n := 0
	for _, m := range messages {
		if m.Role == "user" && strings.TrimSpace(m.Content) != "" {
			n++
		}
	}
	return n
}

func newServerChatID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (r *REPL) ensureServerChatID() {
	if r.session.ChatID == "" {
		r.session.ChatID = newServerChatID()
	}
}

func (r *REPL) maybeAutoTitleServerChat() {
	id := r.session.ChatID
	if id == "" {
		return
	}
	if !needsAutoTitle(r.session.Title) {
		return
	}
	if countUserMessages(r.session.Messages) != 1 {
		return
	}

	title := session.DeriveTitle(r.session.Messages)
	if needsAutoTitle(title) {
		return
	}

	r.session.Title = title
	r.session.LocalTitle = title
	r.persistSession()

	go func() {
		if err := r.client.UpdateChatTitle(id, title); err != nil {
			return
		}
		r.chats.updateTitle(id, title)
	}()
}

func (r *REPL) ToggleServerChatPin(id string) (bool, error) {
	pinned, err := r.client.ToggleChatPinned(id)
	if err != nil {
		return false, err
	}
	r.chats.setPinned(id, pinned)
	return pinned, nil
}

func (r *REPL) toggleCurrentChatPin() SlashResult {
	if r.session.ChatID == "" {
		return SlashResult{Err: fmt.Errorf("no linked server chat — resume one with /chats or start chatting")}
	}
	pinned, err := r.ToggleServerChatPin(r.session.ChatID)
	if err != nil {
		return SlashResult{Err: err}
	}
	state := "unpinned"
	if pinned {
		state = "pinned"
	}
	return SlashResult{Output: fmt.Sprintf("chat %s", state)}
}

func (r *REPL) TogglePickerChatPin(id string) (bool, error) {
	if id == "" {
		return false, fmt.Errorf("no chat selected")
	}
	return r.ToggleServerChatPin(id)
}