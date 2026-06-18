package tui

import "strconv"

type collapseHit struct {
	blockKey string
	line     int
}

type collapseState struct {
	collapsed map[string]bool
	hits      []collapseHit
}

func newCollapseState() *collapseState {
	return &collapseState{collapsed: make(map[string]bool)}
}

func collapseKey(msgIdx int, blockID string) string {
	return strconv.Itoa(msgIdx) + ":" + blockID
}

func (c *collapseState) isCollapsed(msgIdx int, block Block) bool {
	key := collapseKey(msgIdx, block.ID)
	if v, ok := c.collapsed[key]; ok {
		return v
	}
	return block.Collapsed
}

func (c *collapseState) toggle(msgIdx int, blockID string, defaultCollapsed bool) {
	key := collapseKey(msgIdx, blockID)
	cur, ok := c.collapsed[key]
	if !ok {
		cur = defaultCollapsed
	}
	c.collapsed[key] = !cur
}

func (c *collapseState) toggleByKey(key string) {
	cur := c.collapsed[key]
	c.collapsed[key] = !cur
}

func (c *collapseState) toggleAllInMessage(msgIdx int, blocks []Block) {
	anyExpanded := false
	for _, b := range blocks {
		if isCollapsible(b) && !c.isCollapsed(msgIdx, b) {
			anyExpanded = true
			break
		}
	}
	for _, b := range blocks {
		if !isCollapsible(b) {
			continue
		}
		c.collapsed[collapseKey(msgIdx, b.ID)] = anyExpanded
	}
}

func (c *collapseState) resetHits() {
	c.hits = c.hits[:0]
}

func (c *collapseState) addHit(msgIdx int, blockID string, line int) {
	c.hits = append(c.hits, collapseHit{
		blockKey: collapseKey(msgIdx, blockID),
		line:     line,
	})
}

func (c *collapseState) hitAt(contentLine int) (string, bool) {
	for _, h := range c.hits {
		if h.line == contentLine {
			return h.blockKey, true
		}
	}
	return "", false
}