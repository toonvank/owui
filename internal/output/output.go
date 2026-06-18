package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	brandStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	okStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

func Banner(version string) {
	title := brandStyle.Render("owui")
	sub := dimStyle.Render(" — Open WebUI CLI v" + version)
	line := strings.Repeat("─", 40)
	fmt.Println(title + sub)
	fmt.Println(borderStyle.Render(line))
}

func Error(msg string) {
	fmt.Fprintln(os.Stderr, errStyle.Render("✗ error")+dimStyle.Render(" · ")+msg)
}

func Success(msg string) {
	fmt.Println(okStyle.Render("✓ ") + msg)
}

func Info(msg string) {
	fmt.Println(dimStyle.Render("→ ") + msg)
}

func Table(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = lipgloss.Width(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				w := lipgloss.Width(cell)
				if w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	headerParts := make([]string, len(headers))
	for i, h := range headers {
		headerParts[i] = headerStyle.Render(padDisplay(h, widths[i]))
	}
	fmt.Println(strings.Join(headerParts, "  "))
	fmt.Println(borderStyle.Render(strings.Repeat("─", sumWidths(widths)+2*(len(headers)-1))))

	for _, row := range rows {
		parts := make([]string, len(row))
		for i, cell := range row {
			if i < len(widths) {
				parts[i] = padDisplay(cell, widths[i])
			}
		}
		fmt.Println(strings.Join(parts, "  "))
	}
}

func padDisplay(s string, w int) string {
	pad := w - lipgloss.Width(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
}

func sumWidths(widths []int) int {
	n := 0
	for _, w := range widths {
		n += w
	}
	return n
}

func JSON(data string) {
	fmt.Println(data)
}

func Dim(msg string) string {
	return dimStyle.Render(msg)
}