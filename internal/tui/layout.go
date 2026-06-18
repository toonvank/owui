package tui

import (
	"fmt"
	"strings"
)

// alignRow places right at the right margin of width terminal cells.
func alignRow(left, right string, width int) string {
	if right == "" {
		return left
	}
	gap := width - displayWidth(left) - displayWidth(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func prefixLines(text, prefix string) string {
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line == "" {
			lines[i] = prefix
			continue
		}
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func formatTurnMetrics(elapsed float64, chars int) string {
	if chars > 0 {
		return mutedStyle.Render(fmt.Sprintf("[turn: %.1fs  ↕%s]", elapsed, formatCount(chars)))
	}
	return mutedStyle.Render(fmt.Sprintf("[turn: %.1fs]", elapsed))
}

func formatCount(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 10000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fk", float64(n)/1000)
}