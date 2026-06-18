package repl

import (
	"sort"
	"strings"
	"sync"

	"github.com/toonvank/owui/internal/api"
)

type chatCache struct {
	mu      sync.RWMutex
	entries []api.ChatSummary
	loaded  bool
	err     error
}

func chatRecency(ch api.ChatSummary) int64 {
	if ch.UpdatedAt > 0 {
		return ch.UpdatedAt
	}
	return ch.CreatedAt
}

func sortChatsByRecency(chats []api.ChatSummary) {
	sort.Slice(chats, func(i, j int) bool {
		return chatRecency(chats[i]) > chatRecency(chats[j])
	})
}

func (c *chatCache) set(chats []api.ChatSummary, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.err = err
		c.loaded = true
		return
	}
	entries := append([]api.ChatSummary(nil), chats...)
	sortChatsByRecency(entries)
	c.entries = entries
	c.loaded = true
	c.err = nil
}

func (c *chatCache) list() ([]api.ChatSummary, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.err != nil {
		return nil, c.err
	}
	if !c.loaded {
		return nil, nil
	}
	out := make([]api.ChatSummary, len(c.entries))
	copy(out, c.entries)
	return out, nil
}

func (c *chatCache) ready() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded && c.err == nil
}

// ChatPick is one entry in the interactive /chats picker.
type ChatPick struct {
	ID      string
	Title   string
	ShortID string
	Current bool
}

func (r *REPL) preloadChats() {
	go func() {
		chats, err := r.client.ListChats(1)
		r.chats.set(chats, err)
	}()
}

func (r *REPL) ensureChats() {
	if r.chats.ready() {
		return
	}
	chats, err := r.client.ListChats(1)
	r.chats.set(chats, err)
}

// ChatsReady reports whether the chat list has been fetched.
func (r *REPL) ChatsReady() bool {
	return r.chats.ready()
}

func chatMatchScore(query string, ch api.ChatSummary) int {
	if query == "" {
		return 1
	}
	q := strings.ToLower(query)
	id := strings.ToLower(ch.ID)
	title := strings.ToLower(ch.Title)

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

// SearchChats returns chats matching a fuzzy query for the interactive picker.
func (r *REPL) SearchChats(query string, limit int) []ChatPick {
	if limit <= 0 {
		limit = 50
	}
	r.ensureChats()
	entries, err := r.chats.list()
	if err != nil || len(entries) == 0 {
		return nil
	}

	type scored struct {
		ch    api.ChatSummary
		score int
	}
	scoredList := make([]scored, 0, len(entries))
	for _, ch := range entries {
		score := chatMatchScore(query, ch)
		if score <= 0 {
			continue
		}
		scoredList = append(scoredList, scored{ch: ch, score: score})
	}
	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].score != scoredList[j].score {
			return scoredList[i].score > scoredList[j].score
		}
		return chatRecency(scoredList[i].ch) > chatRecency(scoredList[j].ch)
	})
	if len(scoredList) > limit {
		scoredList = scoredList[:limit]
	}

	out := make([]ChatPick, 0, len(scoredList))
	for _, s := range scoredList {
		title := s.ch.Title
		if title == "" {
			title = s.ch.ID[:8]
		}
		short := s.ch.ID
		if len(short) > 8 {
			short = short[:8]
		}
		out = append(out, ChatPick{
			ID:      s.ch.ID,
			Title:   title,
			ShortID: short,
			Current: s.ch.ID == r.session.ChatID,
		})
	}
	return out
}

// ResumeChatByID loads a server chat into the current session.
func (r *REPL) ResumeChatByID(id string) SlashResult {
	return r.slashLoad(id)
}