package repl

import (
	"sort"
	"strings"
	"sync"

	"github.com/toonvank/owui/internal/api"
)

type knowledgeCache struct {
	mu      sync.RWMutex
	entries []api.KnowledgeItem
	loaded  bool
	err     error
}

func (c *knowledgeCache) set(items []api.KnowledgeItem, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.err = err
		c.loaded = true
		return
	}
	c.entries = append([]api.KnowledgeItem(nil), items...)
	c.loaded = true
	c.err = nil
}

func (c *knowledgeCache) list() ([]api.KnowledgeItem, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.err != nil {
		return nil, c.err
	}
	if !c.loaded {
		return nil, nil
	}
	out := make([]api.KnowledgeItem, len(c.entries))
	copy(out, c.entries)
	return out, nil
}

func (c *knowledgeCache) ready() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded && c.err == nil
}

// KnowledgePick is one entry in the interactive /knowledge picker.
type KnowledgePick struct {
	ID      string
	Name    string
	Desc    string
	Current bool
}

func (r *REPL) preloadKnowledge() {
	go func() {
		items, err := r.client.ListKnowledge()
		r.knowledge.set(items, err)
	}()
}

func (r *REPL) ensureKnowledge() {
	if r.knowledge.ready() {
		return
	}
	items, err := r.client.ListKnowledge()
	r.knowledge.set(items, err)
}

// KnowledgeReady reports whether the knowledge list has been fetched.
func (r *REPL) KnowledgeReady() bool {
	return r.knowledge.ready()
}

func knowledgeMatchScore(query string, item api.KnowledgeItem) int {
	if query == "" {
		return 1
	}
	q := strings.ToLower(query)
	id := strings.ToLower(item.ID)
	name := strings.ToLower(item.Name)
	desc := strings.ToLower(item.Description)

	if strings.HasPrefix(id, q) {
		return 800 + len(q)
	}
	if strings.HasPrefix(name, q) {
		return 700 + len(q)
	}
	if strings.Contains(name, q) {
		return 500 + len(q)
	}
	if strings.Contains(desc, q) {
		return 400 + len(q)
	}
	if strings.Contains(id, q) {
		return 300 + len(q)
	}
	return 0
}

// SearchKnowledge returns collections matching a fuzzy query.
func (r *REPL) SearchKnowledge(query string, limit int) []KnowledgePick {
	if limit <= 0 {
		limit = 50
	}
	r.ensureKnowledge()
	entries, err := r.knowledge.list()
	if err != nil || len(entries) == 0 {
		return nil
	}

	type scored struct {
		item  api.KnowledgeItem
		score int
	}
	scoredList := make([]scored, 0, len(entries))
	for _, item := range entries {
		score := knowledgeMatchScore(query, item)
		if score <= 0 {
			continue
		}
		scoredList = append(scoredList, scored{item: item, score: score})
	}
	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].score != scoredList[j].score {
			return scoredList[i].score > scoredList[j].score
		}
		return scoredList[i].item.Name < scoredList[j].item.Name
	})
	if len(scoredList) > limit {
		scoredList = scoredList[:limit]
	}

	out := make([]KnowledgePick, 0, len(scoredList))
	for _, s := range scoredList {
		desc := s.item.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		out = append(out, KnowledgePick{
			ID:      s.item.ID,
			Name:    s.item.Name,
			Desc:    desc,
			Current: s.item.ID == r.session.CollectionID,
		})
	}
	return out
}

func (r *REPL) bestKnowledgeMatch(query string) api.KnowledgeItem {
	picks := r.SearchKnowledge(query, 1)
	if len(picks) == 0 {
		return api.KnowledgeItem{}
	}
	return api.KnowledgeItem{ID: picks[0].ID, Name: picks[0].Name}
}

// SetKnowledgeCollection switches the active knowledge collection.
func (r *REPL) SetKnowledgeCollection(id, name string) {
	r.setKnowledgeCollection(id, name)
}