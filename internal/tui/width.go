package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func displayWidth(s string) int {
	return lipgloss.Width(s)
}

func truncateDisplay(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if displayWidth(s) <= max {
		return s
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > max-1 {
			b.WriteString("…")
			break
		}
		b.WriteRune(r)
		w += rw
	}
	return b.String()
}

func wrapLineDisplay(line string, width int) []string {
	if displayWidth(line) <= width {
		return []string{line}
	}
	var lines []string
	remaining := line
	for displayWidth(remaining) > width {
		cut := findBreak(remaining, width)
		lines = append(lines, remaining[:cut])
		remaining = strings.TrimLeft(remaining[cut:], " ")
	}
	if remaining != "" {
		lines = append(lines, remaining)
	}
	return lines
}

func findBreak(s string, width int) int {
	if width >= len(s) {
		return len(s)
	}
	best := 0
	w := 0
	for i, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > width {
			break
		}
		w += rw
		best = i + len(string(r))
		if r == ' ' && w > width/2 {
			return best
		}
	}
	if best == 0 {
		return len(s)
	}
	return best
}

func wrapContentDisplay(text string, width int) string {
	if width < 20 {
		width = 20
	}
	lines := strings.Split(text, "\n")
	var out []string
	for _, line := range lines {
		out = append(out, wrapLineDisplay(line, width)...)
	}
	return strings.Join(out, "\n")
}