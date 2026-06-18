package tui

import (
	"strings"
)

type renderContext struct {
	width    int
	collapse *collapseState
	msgIdx   int
	lineNo   int
}

func (rc *renderContext) nextLine(s string) string {
	rc.lineNo++
	return s
}

func renderMessageBlocks(role, model, content string, width int, msgIdx int, collapse *collapseState, streaming bool, cache *renderCache) string {
	cacheKey := cache.key(msgIdx, role, model, content, streaming, width, collapse)
	if cached, ok := cache.get(cacheKey); ok {
		return cached
	}

	rendered := renderMessageBlocksUncached(role, model, content, width, msgIdx, collapse, streaming)
	cache.set(cacheKey, rendered)
	return rendered
}

func contentWidth(width int) int {
	w := width - 2
	if w < 20 {
		return 20
	}
	return w
}

func renderMessageBlocksUncached(role, model, content string, width int, msgIdx int, collapse *collapseState, streaming bool) string {
	innerW := contentWidth(width)

	switch role {
	case "info":
		return infoLineStyle.Render(pipeStyle.Render("│ ") + content)
	case "system":
		return dimStyle.Render(pipeStyle.Render("│ ") + content)
	case "error":
		return errorStyle.Render(pipeStyle.Render("│ ") + content)
	case "user":
		return renderUserMessage(content, width, innerW)
	case "assistant":
		return renderAssistantMessage(model, content, width, innerW, msgIdx, collapse, streaming)
	default:
		return dimStyle.Render(content)
	}
}

func renderUserMessage(content string, width, innerW int) string {
	header := alignRow(
		chevronStyle.Render("❯ ")+activeStyle.Render("you"),
		dimStyle.Render("sent"),
		width,
	)
	body := wrapContentDisplay(content, innerW-2)
	card := userCardStyle.Width(width).Render(body)
	return header + "\n" + card
}

func renderAssistantMessage(model, content string, width, innerW int, msgIdx int, collapse *collapseState, streaming bool) string {
	meta := mutedStyle.Render(truncateDisplay(model, 40))
	if meta == "" {
		meta = mutedStyle.Render("assistant")
	}

	var status string
	switch {
	case streaming && content == "":
		status = mutedStyle.Render("[…]")
	case streaming:
		status = mutedStyle.Render("[streaming]")
	default:
		status = doneStyle.Render("[done]")
	}
	header := alignRow(meta, status, width)

	if streaming && content == "" {
		return header + "\n" + dimStyle.Render("│ …")
	}

	var body string
	if streaming {
		body = renderPlainWithFences(content, innerW)
	} else {
		rc := &renderContext{width: innerW, collapse: collapse, msgIdx: msgIdx}
		blocks := ParseBlocks(content)
		if len(blocks) == 0 {
			body = renderMarkdown(content, innerW)
		} else {
			var parts []string
			for _, block := range blocks {
				parts = append(parts, rc.renderBlock(block))
			}
			body = strings.Join(parts, "\n")
		}
	}

	if body == "" {
		return header
	}
	return header + "\n" + assistBodyStyle.Render(body)
}

func (rc *renderContext) renderBlock(block Block) string {
	switch block.Kind {
	case BlockDiff:
		return rc.nextLine(pipeStyle.Render("│ ") + renderDiff(block.Content, rc.width))
	case BlockCode:
		return rc.renderCollapsible(block, func() string {
			return codeBlockStyle.Width(rc.width).Render(wrapContentDisplay(block.Content, rc.width-2))
		})
	case BlockThinking, BlockTool:
		return rc.renderCollapsible(block, func() string {
			return mutedStyle.Render(wrapContentDisplay(block.Content, rc.width))
		})
	default:
		if needsRichMarkdown(block.Content) {
			return rc.nextLine(renderMarkdown(block.Content, rc.width))
		}
		return rc.nextLine(wrapContentDisplay(block.Content, rc.width))
	}
}

func needsRichMarkdown(text string) bool {
	if len(text) > 8000 {
		return false
	}
	for _, marker := range []string{"**", "__", "# ", "## ", "- ", "1. ", "[", "`", "|"} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return strings.Contains(text, "\n\n")
}

func (rc *renderContext) renderCollapsible(block Block, renderBody func() string) string {
	if !isCollapsible(block) {
		return rc.nextLine(pipeStyle.Render("│ ") + renderBody())
	}

	collapsed := rc.collapse.isCollapsed(rc.msgIdx, block)
	icon := "▾"
	if collapsed {
		icon = "▸"
	}
	title := collapseHeaderStyle.Render(icon+" ") + collapseTitleStyle.Render(blockSummary(block))
	rc.collapse.addHit(rc.msgIdx, block.ID, rc.lineNo)
	line := rc.nextLine(pipeStyle.Render("│ ") + title)
	if collapsed {
		return line
	}
	body := renderBody()
	return line + "\n" + prefixLines(body, pipeStyle.Render("│ "))
}

func renderPlainWithFences(content string, width int) string {
	parts := strings.Split(content, "```")
	if len(parts) < 3 {
		return wrapContentDisplay(content, width)
	}
	var b strings.Builder
	for i, part := range parts {
		if i%2 == 0 {
			b.WriteString(wrapContentDisplay(part, width))
			continue
		}
		fence := strings.TrimPrefix(part, "\n")
		lang := ""
		body := fence
		if nl := strings.Index(fence, "\n"); nl >= 0 {
			lang = strings.TrimSpace(fence[:nl])
			body = fence[nl+1:]
		}
		body = strings.TrimSuffix(strings.TrimSpace(body), "\n")
		if strings.EqualFold(lang, "diff") || looksLikeDiff(body) {
			b.WriteString(pipeStyle.Render("│ "))
			b.WriteString(renderDiff(body, width))
		} else {
			title := lang
			if title == "" {
				title = "code"
			}
			b.WriteString(pipeStyle.Render("│ "))
			b.WriteString(codeBlockStyle.Width(width).Render(
				mutedStyle.Render("─ "+title+" ─") + "\n" + wrapContentDisplay(body, width-2),
			))
		}
	}
	return b.String()
}

func renderChatLog(messages []chatMessage, width int, collapse *collapseState, cache *renderCache, activeModel string) string {
	collapse.resetHits()
	if len(messages) == 0 {
		return dimStyle.Render(pipeStyle.Render("│ ") + "Send a message to start chatting.") + "\n" +
			dimStyle.Render(pipeStyle.Render("│ ") + "? help  ·  /commands  ·  @model  ·  Tab complete")
	}
	var parts []string
	for i, msg := range messages {
		model := msg.model
		if model == "" && msg.role == "assistant" {
			model = activeModel
		}
		parts = append(parts, renderMessageBlocks(msg.role, model, msg.content, width, i, collapse, msg.streaming, cache))
	}
	return strings.Join(parts, "\n\n")
}

func renderSuggestions(suggestions []suggestionItem, selected int, width int) string {
	if len(suggestions) == 0 {
		return ""
	}
	maxShow := 6
	show := suggestions
	if len(show) > maxShow {
		show = show[:maxShow]
	}
	var lines []string
	for i, s := range show {
		desc := ""
		if s.description != "" {
			desc = suggestDescStyle.Render(s.description)
		}
		descW := displayWidth(desc)
		maxText := width - descW - 6
		if maxText < 12 {
			maxText = 12
		}
		prefix := pipeStyle.Render("│ ")
		text := prefix + truncateDisplay(s.text, maxText-2)
		line := text
		if desc != "" {
			gap := width - displayWidth(text) - descW - 2
			if gap < 2 {
				gap = 2
			}
			line += strings.Repeat(" ", gap) + desc
		}
		if i == selected {
			line = suggestSelectedStyle.Render(line)
		} else {
			line = suggestStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func renderFooter(bar string, width int) string {
	if displayWidth(bar) > width {
		return footerStyle.Width(width).Render(truncateDisplay(bar, width))
	}
	return footerStyle.Render(bar)
}

func renderHeader(baseURL, model, sessionLabel string, width int) string {
	title := brandStyle.Render("owui") + dimStyle.Render("  ·  Open WebUI")

	host := strings.TrimPrefix(strings.TrimPrefix(baseURL, "https://"), "http://")
	host = truncateDisplay(host, 32)
	model = truncateDisplay(model, 28)

	sep := dimStyle.Render("  │  ")
	parts := []string{mutedStyle.Render(host), activeStyle.Render(model)}
	if sessionLabel != "" {
		parts = append(parts, dimStyle.Render(truncateDisplay(sessionLabel, 20)))
	}
	status := strings.Join(parts, sep)

	top := headerStyle.Width(width).Render(title)
	bottom := statusStyle.Width(width).Render(status)
	divider := dividerStyle.Render(strings.Repeat("─", width))
	return top + "\n" + bottom + "\n" + divider
}

func renderStatusBar(width int, thinking bool, elapsed float64, chars int, spinnerView string) string {
	spin := spinnerStyle.Render(spinnerView)
	label := mutedStyle.Render("thinking")
	if chars > 0 {
		label = mutedStyle.Render("streaming")
	} else if !thinking {
		label = mutedStyle.Render("working")
	}
	left := spin + "  " + label
	row := alignRow(left, formatTurnMetrics(elapsed, chars), width)
	return statusStyle.Width(width).Render(row)
}