package tui

import (
	"fmt"
	"regexp"
	"strings"
)

type BlockKind int

const (
	BlockText BlockKind = iota
	BlockCode
	BlockDiff
	BlockThinking
	BlockTool
)

type Block struct {
	Kind      BlockKind
	ID        string
	Title     string
	Lang      string
	Content   string
	Collapsed bool
	Lines     int
}

var (
	thinkingTagRe = regexp.MustCompile(`(?is)<think(?:ing)?>(.*?)</think(?:ing)?>`)
	reasoningRe   = regexp.MustCompile(`(?is)<reasoning>(.*?)</reasoning>`)
	toolHeaderRe  = regexp.MustCompile(`(?m)^#{1,3}\s*(?:Tool|Action|Running)(?::\s*(.+))?$`)
)

func ParseBlocks(content string) []Block {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	var blocks []Block
	rest := content
	seq := 0

	extractTagged := func(re *regexp.Regexp, kind BlockKind, title string, collapseDefault bool) {
		for {
			loc := re.FindStringSubmatchIndex(rest)
			if loc == nil {
				break
			}
			before := strings.TrimSpace(rest[:loc[0]])
			if before != "" {
				blocks = append(blocks, splitTextAndFences(before, &seq)...)
			}
			body := strings.TrimSpace(rest[loc[2]:loc[3]])
			id := fmt.Sprintf("b%d", seq)
			seq++
			blocks = append(blocks, Block{
				Kind:      kind,
				ID:        id,
				Title:     title,
				Content:   body,
				Collapsed: collapseDefault,
				Lines:     countLines(body),
			})
			rest = strings.TrimSpace(rest[loc[1]:])
		}
	}

	extractTagged(thinkingTagRe, BlockThinking, "Thinking", true)
	extractTagged(reasoningRe, BlockThinking, "Reasoning", true)

	if strings.TrimSpace(rest) != "" {
		blocks = append(blocks, splitTextAndFences(rest, &seq)...)
	}
	return blocks
}

func splitTextAndFences(content string, seq *int) []Block {
	var blocks []Block
	parts := strings.Split(content, "```")
	for i, part := range parts {
		if i%2 == 0 {
			text := strings.TrimSpace(part)
			if text == "" {
				continue
			}
			for _, chunk := range splitToolSections(text, seq) {
				blocks = append(blocks, chunk)
			}
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
		id := fmt.Sprintf("b%d", *seq)
		*seq++

		langLower := strings.ToLower(lang)
		switch {
		case langLower == "diff" || looksLikeDiff(body):
			blocks = append(blocks, Block{
				Kind:    BlockDiff,
				ID:      id,
				Title:   "Diff",
				Lang:    lang,
				Content: body,
				Lines:   countLines(body),
			})
		case langLower == "thinking" || langLower == "reasoning":
			title := "Thinking"
			if langLower == "reasoning" {
				title = "Reasoning"
			}
			blocks = append(blocks, Block{
				Kind:      BlockThinking,
				ID:        id,
				Title:     title,
				Content:   body,
				Collapsed: true,
				Lines:     countLines(body),
			})
		default:
			title := "Code"
			if lang != "" {
				title = "Code (" + lang + ")"
			}
			lines := countLines(body)
			b := Block{
				Kind:    BlockCode,
				ID:      id,
				Title:   title,
				Lang:    lang,
				Content: body,
				Lines:   lines,
			}
			if lines > 20 {
				b.Collapsed = true
			}
			blocks = append(blocks, b)
		}
	}
	return blocks
}

func splitToolSections(text string, seq *int) []Block {
	lines := strings.Split(text, "\n")
	var blocks []Block
	var textBuf strings.Builder
	flushText := func() {
		t := strings.TrimSpace(textBuf.String())
		textBuf.Reset()
		if t == "" {
			return
		}
		id := fmt.Sprintf("b%d", *seq)
		*seq++
		blocks = append(blocks, Block{
			Kind:    BlockText,
			ID:      id,
			Content: t,
			Lines:   countLines(t),
		})
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trim := strings.TrimSpace(line)

		if isToolHeader(trim) {
			flushText()
			var body []string
			title := toolTitle(trim)
			maxBody := 200
			for j := i + 1; j < len(lines) && len(body) < maxBody; j++ {
				next := strings.TrimSpace(lines[j])
				if isToolHeader(next) || strings.HasPrefix(next, "```") {
					i = j - 1
					break
				}
				body = append(body, lines[j])
				if j == len(lines)-1 {
					i = j
				}
			}
			content := strings.TrimSpace(strings.Join(body, "\n"))
			id := fmt.Sprintf("b%d", *seq)
			*seq++
			blocks = append(blocks, Block{
				Kind:      BlockTool,
				ID:        id,
				Title:     title,
				Content:   content,
				Collapsed: true,
				Lines:     countLines(content),
			})
			continue
		}

		if textBuf.Len() > 0 {
			textBuf.WriteByte('\n')
		}
		textBuf.WriteString(line)
	}
	flushText()
	return blocks
}

func isToolHeader(line string) bool {
	if line == "" {
		return false
	}
	trim := strings.TrimSpace(line)
	if toolHeaderRe.MatchString(trim) {
		return true
	}
	lower := strings.ToLower(trim)
	switch {
	case strings.HasPrefix(trim, "🔧"):
		return true
	case lower == "tool:" || strings.HasPrefix(lower, "tool: "):
		return true
	case strings.HasPrefix(lower, "**tool:**"), strings.HasPrefix(lower, "**tool: "):
		return true
	case strings.HasPrefix(lower, "**running:**"), strings.HasPrefix(lower, "**running: "):
		return true
	default:
		return false
	}
}

func toolTitle(line string) string {
	line = strings.TrimPrefix(line, "🔧")
	line = strings.TrimSpace(line)
	if m := toolHeaderRe.FindStringSubmatch(line); len(m) > 1 && m[1] != "" {
		return strings.TrimSpace(m[1])
	}
	line = strings.Trim(line, "*# ")
	if idx := strings.Index(line, ":"); idx >= 0 {
		return strings.TrimSpace(line[idx+1:])
	}
	return line
}

func looksLikeDiff(body string) bool {
	lines := strings.Split(body, "\n")
	hits := 0
	for i, l := range lines {
		if i > 8 {
			break
		}
		if strings.HasPrefix(l, "+++ ") || strings.HasPrefix(l, "--- ") ||
			strings.HasPrefix(l, "@@") ||
			strings.HasPrefix(l, "+") || strings.HasPrefix(l, "-") {
			hits++
		}
	}
	return hits >= 2
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func blockSummary(b Block) string {
	switch b.Kind {
	case BlockThinking:
		return fmt.Sprintf("%s · %d lines", b.Title, b.Lines)
	case BlockTool:
		if b.Lines > 0 {
			return fmt.Sprintf("%s · %d lines", b.Title, b.Lines)
		}
		return b.Title
	case BlockCode:
		return fmt.Sprintf("%s · %d lines", b.Title, b.Lines)
	case BlockDiff:
		return fmt.Sprintf("Diff · %d lines", b.Lines)
	default:
		return ""
	}
}

func isCollapsible(b Block) bool {
	return b.Kind == BlockThinking || b.Kind == BlockTool || b.Kind == BlockCode
}