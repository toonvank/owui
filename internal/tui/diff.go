package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	diffAddStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	diffDelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	diffHunkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	diffFileStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Bold(true)
	diffCtxStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	diffBorderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(colorDim).Padding(0, 1)
)

func renderDiff(content string, width int) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			out = append(out, diffFileStyle.Render(truncateDisplay(line, width-4)))
		case strings.HasPrefix(line, "@@"):
			out = append(out, diffHunkStyle.Render(truncateDisplay(line, width-4)))
		case strings.HasPrefix(line, "+"):
			out = append(out, diffAddStyle.Render(truncateDisplay(line, width-4)))
		case strings.HasPrefix(line, "-"):
			out = append(out, diffDelStyle.Render(truncateDisplay(line, width-4)))
		default:
			out = append(out, diffCtxStyle.Render(truncateDisplay(line, width-4)))
		}
	}
	body := strings.Join(out, "\n")
	return diffBorderStyle.Width(width).Render(body)
}