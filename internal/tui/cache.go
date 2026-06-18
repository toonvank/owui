package tui

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type renderCache struct {
	entries map[string]string
}

func newRenderCache() *renderCache {
	return &renderCache{entries: make(map[string]string)}
}

func (c *renderCache) key(msgIdx int, role, model, content string, streaming bool, width int, collapse *collapseState) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d|%s|%s|%t|%d|%d|", msgIdx, role, model, streaming, width, len(content))
	h.Write([]byte(content))
	if collapse != nil {
		for k, v := range collapse.collapsed {
			fmt.Fprintf(h, "%s=%t;", k, v)
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func (c *renderCache) get(key string) (string, bool) {
	v, ok := c.entries[key]
	return v, ok
}

func (c *renderCache) set(key, rendered string) {
	if len(c.entries) > 200 {
		c.entries = make(map[string]string)
	}
	c.entries[key] = rendered
}

func (c *renderCache) invalidateFrom(idx int) {
	prefix := fmt.Sprintf("%d", idx)
	for k := range c.entries {
		_ = prefix
		delete(c.entries, k)
	}
}

func (c *renderCache) clear() {
	c.entries = make(map[string]string)
}