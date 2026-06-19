package repl

import (
	"sort"
	"strings"
	"sync"

	"github.com/toonvank/owui/internal/api"
)

type functionCache struct {
	mu      sync.RWMutex
	entries []api.OWUIFunction
	loaded  bool
	err     error
}

func (c *functionCache) set(fns []api.OWUIFunction, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.err = err
		c.loaded = true
		return
	}
	c.entries = append([]api.OWUIFunction(nil), fns...)
	c.loaded = true
	c.err = nil
}

func (c *functionCache) list() ([]api.OWUIFunction, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.err != nil {
		return nil, c.err
	}
	if !c.loaded {
		return nil, nil
	}
	out := make([]api.OWUIFunction, len(c.entries))
	copy(out, c.entries)
	return out, nil
}

func (c *functionCache) ready() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded && c.err == nil
}

// FunctionPick is one entry in the interactive filter/tool picker.
type FunctionPick struct {
	ID      string
	Name    string
	Scope   string
	Enabled bool
	Global  bool
}

func (r *REPL) preloadFunctions() {
	go func() {
		fns, err := r.client.ListFunctions()
		r.functions.set(fns, err)
	}()
}

func (r *REPL) ensureFunctions() {
	if r.functions.ready() {
		return
	}
	fns, err := r.client.ListFunctions()
	r.functions.set(fns, err)
}

// FunctionsReady reports whether the function list has been fetched.
func (r *REPL) FunctionsReady() bool {
	return r.functions.ready()
}

func functionMatchScore(query string, fn api.OWUIFunction) int {
	if query == "" {
		return 1
	}
	q := strings.ToLower(query)
	id := strings.ToLower(fn.ID)
	name := strings.ToLower(fn.Name)
	if strings.HasPrefix(name, q) {
		return 700 + len(q)
	}
	if strings.Contains(name, q) {
		return 500 + len(q)
	}
	if strings.HasPrefix(id, q) {
		return 400 + len(q)
	}
	if strings.Contains(id, q) {
		return 300 + len(q)
	}
	return 0
}

func (r *REPL) searchFunctions(fnType, query string, limit int) []FunctionPick {
	if limit <= 0 {
		limit = 50
	}
	r.ensureFunctions()
	entries, err := r.functions.list()
	if err != nil {
		return nil
	}

	var enabled func(string) bool
	switch fnType {
	case "filter":
		enabled = r.isFilterEnabled
	case "tool":
		enabled = r.isToolEnabled
	default:
		return nil
	}

	type scored struct {
		fn    api.OWUIFunction
		score int
	}
	scoredList := make([]scored, 0)
	for _, fn := range entries {
		if fn.Type != fnType {
			continue
		}
		score := functionMatchScore(query, fn)
		if score <= 0 {
			continue
		}
		scoredList = append(scoredList, scored{fn: fn, score: score})
	}
	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].score != scoredList[j].score {
			return scoredList[i].score > scoredList[j].score
		}
		return scoredList[i].fn.Name < scoredList[j].fn.Name
	})
	if len(scoredList) > limit {
		scoredList = scoredList[:limit]
	}

	out := make([]FunctionPick, 0, len(scoredList))
	for _, s := range scoredList {
		scope := "model"
		if s.fn.IsGlobal {
			scope = "global"
		}
		out = append(out, FunctionPick{
			ID:      s.fn.ID,
			Name:    s.fn.Name,
			Scope:   scope,
			Enabled: enabled(s.fn.ID),
			Global:  s.fn.IsGlobal,
		})
	}
	return out
}

// SearchFilters returns filter functions for the interactive picker.
func (r *REPL) SearchFilters(query string, limit int) []FunctionPick {
	return r.searchFunctions("filter", query, limit)
}

// SearchTools returns tool functions for the interactive picker.
func (r *REPL) SearchTools(query string, limit int) []FunctionPick {
	return r.searchFunctions("tool", query, limit)
}