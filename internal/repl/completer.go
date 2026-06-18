package repl

import (
	"sort"
	"strings"
	"sync"

	"github.com/toonvank/owui/internal/api"
)

type slashCommand struct {
	name string
	desc string
}

var slashCommands = []slashCommand{
	{name: "help", desc: "Commands & shortcuts (?)"},
	{name: "quit", desc: "Exit (Ctrl+C)"},
	{name: "clear", desc: "Clear conversation"},
	{name: "model", desc: "Switch model · Tab"},
	{name: "models", desc: "List models · Tab"},
	{name: "chats", desc: "Browse & resume chats"},
	{name: "server", desc: "Show server URL and setup hints"},
	{name: "setup", desc: "How to run owui setup"},
	{name: "resume", desc: "Resume chat by id"},
	{name: "load", desc: "Alias for /resume"},
	{name: "filters", desc: "Filters & features"},
	{name: "functions", desc: "Alias for /filters"},
	{name: "stream", desc: "Toggle streaming"},
	{name: "session", desc: "Local session save/list/load"},
}

type modelEntry struct {
	ID     string
	Name   string
	Custom bool
}

type modelCache struct {
	mu      sync.RWMutex
	entries []modelEntry
	loaded  bool
	err     error
}

func (c *modelCache) setFromEntries(entries []modelEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append([]modelEntry(nil), entries...)
	c.loaded = true
	c.err = nil
}

func (c *modelCache) set(models []api.Model) {
	entries := make([]modelEntry, 0, len(models))
	for _, m := range models {
		if m.ID == "" {
			continue
		}
		entries = append(entries, modelEntry{
			ID:     m.ID,
			Name:   m.DisplayName(),
			Custom: m.IsCustom(),
		})
	}
	c.setFromEntries(entries)
}

func (c *modelCache) setErr(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = err
	c.loaded = true
}

func (c *modelCache) entriesList() ([]modelEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.err != nil {
		return nil, c.err
	}
	if !c.loaded {
		return nil, nil
	}
	out := make([]modelEntry, len(c.entries))
	copy(out, c.entries)
	return out, nil
}

func (c *modelCache) ready() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded && c.err == nil
}

func (r *REPL) preloadModels() {
	go func() {
		models, err := r.client.ListModels()
		if err != nil {
			r.models.setErr(err)
			return
		}
		r.models.set(models)
	}()
}

func (r *REPL) ensureModels() {
	if r.models.ready() {
		return
	}
	models, err := r.client.ListModels()
	if err != nil {
		r.models.setErr(err)
		return
	}
	r.models.set(models)
}

func (r *REPL) completeSlash(line string) []Suggestion {
	trimmed := strings.TrimPrefix(line, "/")
	space := strings.Index(trimmed, " ")
	if space >= 0 {
		cmd := trimmed[:space]
		arg := trimmed[space+1:]
		switch cmd {
		case "model", "models":
			return r.modelSuggests(arg, false)
		case "resume", "load", "chats":
			return r.chatSuggests(arg)
		}
		return nil
	}
	return r.slashSuggests(trimmed)
}

func (r *REPL) slashSuggests(prefix string) []Suggestion {
	prefix = strings.ToLower(prefix)
	out := make([]Suggestion, 0, len(slashCommands))
	for _, cmd := range slashCommands {
		if prefix == "" || strings.HasPrefix(cmd.name, prefix) {
			out = append(out, Suggestion{
				Text:        "/" + cmd.name + " ",
				Description: cmd.desc,
			})
		}
	}
	return out
}

type scoredModel struct {
	entry modelEntry
	score int
}

func modelMatchScore(query string, e modelEntry) int {
	if query == "" {
		return 1
	}
	q := strings.ToLower(query)
	id := strings.ToLower(e.ID)
	name := strings.ToLower(e.Name)

	if id == q {
		return 1000
	}
	if name == q {
		return 950
	}
	if strings.HasPrefix(id, q) {
		return 800 + len(q)
	}
	if strings.HasPrefix(name, q) {
		return 750 + len(q)
	}
	for _, word := range strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == ':'
	}) {
		if strings.HasPrefix(word, q) {
			return 600 + len(q)
		}
	}
	if strings.Contains(id, q) {
		return 300 + len(q)
	}
	if strings.Contains(name, q) {
		return 250 + len(q)
	}
	return 0
}

func (r *REPL) matchModels(query string, limit int) []scoredModel {
	r.ensureModels()
	entries, err := r.models.entriesList()
	if err != nil || len(entries) == 0 {
		return nil
	}

	scored := make([]scoredModel, 0, limit)
	for _, e := range entries {
		score := modelMatchScore(query, e)
		if score <= 0 {
			continue
		}
		if e.Custom {
			score += 50
		}
		scored = append(scored, scoredModel{entry: e, score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].entry.ID < scored[j].entry.ID
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}
	return scored
}

func (e modelEntry) description(current string) string {
	kind := "model"
	if e.Custom {
		kind = "custom"
	}
	if e.Name != "" {
		kind += " · " + e.Name
	}
	if e.ID == current {
		kind += " (current)"
	}
	return kind
}

func (r *REPL) modelSuggests(prefix string, useAtPrefix bool) []Suggestion {
	matches := r.matchModels(prefix, 20)
	if len(matches) == 0 {
		if !r.models.ready() {
			return []Suggestion{{
				Text:        prefix,
				Description: "loading models...",
			}}
		}
		return nil
	}

	out := make([]Suggestion, 0, len(matches))
	for _, m := range matches {
		text := m.entry.ID
		if useAtPrefix {
			text = "@" + m.entry.ID
		}
		out = append(out, Suggestion{
			Text:        strings.TrimSuffix(text, " "),
			Description: m.entry.description(r.session.Model),
		})
	}
	return out
}

func (r *REPL) chatSuggests(prefix string) []Suggestion {
	chats, err := r.client.ListChats(1)
	if err != nil {
		return nil
	}
	prefixLower := strings.ToLower(prefix)
	out := make([]Suggestion, 0, 15)
	for _, ch := range chats {
		if prefix != "" {
			if !strings.HasPrefix(strings.ToLower(ch.ID), prefixLower) &&
				!strings.Contains(strings.ToLower(ch.Title), prefixLower) {
				continue
			}
		}
		title := ch.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		out = append(out, Suggestion{
			Text:        ch.ID[:8],
			Description: title,
		})
		if len(out) >= 15 {
			break
		}
	}
	return out
}

func (r *REPL) bestModelMatch(query string) string {
	matches := r.matchModels(query, 1)
	if len(matches) == 0 {
		return ""
	}
	return matches[0].entry.ID
}