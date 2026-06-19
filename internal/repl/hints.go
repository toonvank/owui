package repl

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	hintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	keyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
	descStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Bold(true)
	sepStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(16)
	cmdStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	panelDesc  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type shortcut struct {
	key  string
	desc string
}

func hintKey(key string) string {
	return keyStyle.Render(key)
}

func hintSep() string {
	return sepStyle.Render(" │ ")
}

func hintItem(key, desc string) string {
	return hintKey(key) + descStyle.Render(":"+desc)
}

func (r *REPL) printShortcutBar() {
	fmt.Println()
	fmt.Println(r.shortcutBarLine())
}

func (r *REPL) shortcutBarLine() string {
	parts := []string{
		hintItem("Enter", "send"),
		hintItem("↑", "recall"),
		hintItem("Tab", "complete"),
		hintItem("?", "help"),
		hintItem("/model", "switch"),
		hintItem("/sessions", "local"),
		hintItem("/knowledge", "RAG"),
		hintItem("/filters", "toggle"),
		hintItem("C", "collapse"),
		hintItem("Ctrl+C", "quit"),
	}
	if r.session.ChatID != "" {
		title := r.session.Title
		if title == "" {
			title = r.session.ChatID[:8]
		}
		if len(title) > 20 {
			title = title[:17] + "..."
		}
		parts = append([]string{hintItem("/resume", title)}, parts...)
	}
	if !r.cfg.Stream {
		parts = append(parts, hintItem("/stream", "enable"))
	}
	if d := r.LastTurnLatency(); d > 0 {
		parts = append(parts, descStyle.Render(fmt.Sprintf("%.1fs", d.Seconds())))
	}
	return strings.Join(parts, hintSep())
}

// ShortcutsPanel returns the full help panel text.
func ShortcutsPanel() string {
	return shortcutsPanel()
}

func shortcutsPanel() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Keyboard shortcuts"))
	b.WriteString("\n")
	for _, row := range []shortcut{
		{"Enter", "Send message or confirm /model pick"},
		{"/model", "Interactive model picker (type to filter, ↑↓ navigate)"},
		{"Tab / ↑↓", "Navigate autocomplete"},
		{"Esc", "Close menu or help"},
		{"Ctrl+U", "Clear input"},
		{"PgUp/PgDn", "Scroll chat history"},
		{"↑ (empty input)", "Recall last message or input history"},
		{"j/k", "Scroll chat (when vim_keys enabled)"},
		{"C", "Collapse/expand blocks in last reply (empty input)"},
		{"Click ▸/▾", "Toggle thinking, tools, long code"},
		{"Ctrl+C", "Exit"},
		{"?", "Toggle this help panel"},
	} {
		b.WriteString("  ")
		b.WriteString(labelStyle.Render(row.key))
		b.WriteString(panelDesc.Render(row.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Slash commands"))
	b.WriteString("\n")
	for _, row := range []shortcut{
		{"/help", "Show commands and shortcuts"},
		{"/model [filter]", "Interactive model picker — preserves chat history"},
		{"/models [filter]", "List models"},
		{"/chats [filter]", "Browse and resume server chats (↑↓ pick · p pin)"},
		{"/sessions [filter]", "Browse local sessions (↑↓ pick)"},
		{"/knowledge [filter]", "Pick knowledge collection for RAG (↑↓ pick)"},
		{"/file upload <path>", "Upload and attach a file for RAG"},
		{"/file list", "Show attached files"},
		{"/file clear", "Detach all files"},
		{"/knowledge clear", "Detach knowledge collection"},
		{"/profile [filter]", "Switch server profile (↑↓ pick)"},
		{"/profile list", "List configured profiles"},
		{"/server", "Show server URL · run owui setup to change"},
		{"/resume [filter]", "Resume a server chat"},
		{"/session list", "List local sessions (auto-saved)"},
		{"/session new", "Start a fresh local session"},
		{"/session load <id>", "Restore a local session"},
		{"/filters [filter]", "Toggle filters interactively (↑↓ · Enter)"},
		{"/tools [filter]", "Toggle tools interactively"},
		{"/filters list", "List filters with on/off state"},
		{"/filters clear", "Reset to model defaults"},
		{"/model info", "Show model capabilities & integrations"},
		{"/clear", "Clear conversation history"},
		{"/system [prompt]", "Show or set system prompt"},
		{"/title <name>", "Rename local session"},
		{"/export [md|json] [path]", "Export conversation to file"},
		{"/copy", "Copy last assistant reply to clipboard"},
		{"/regen", "Regenerate last response"},
		{"/search <term>", "Search history · /search next to cycle"},
		{"/fork", "Fork conversation (new local session)"},
		{"/delete", "Delete linked server chat"},
		{"/pin", "Pin or unpin linked server chat"},
		{"/stream", "Toggle streaming"},
		{"/quit", "Exit"},
	} {
		b.WriteString("  ")
		b.WriteString(cmdStyle.Render(row.key))
		b.WriteString(panelDesc.Render("  " + row.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Rendering"))
	b.WriteString("\n")
	for _, row := range []shortcut{
		{"Markdown", "Assistant replies render with full MD support"},
		{"Code blocks", "Syntax-aware fenced blocks; long blocks start collapsed"},
		{"Diff blocks", "```diff or +/- patches are colorized"},
		{"<think> blocks", "Reasoning sections collapse by default"},
	} {
		b.WriteString("  ")
		b.WriteString(cmdStyle.Render(row.key))
		b.WriteString(panelDesc.Render("  " + row.desc))
		b.WriteString("\n")
	}

	return b.String()
}

func helpText() string {
	return shortcutsPanel()
}