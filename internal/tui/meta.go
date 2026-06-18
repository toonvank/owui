package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type metaOverlayKind int

const (
	metaNone metaOverlayKind = iota
	metaModelPicker
	metaChatPicker
)

const metaPickerMaxRows = 12

type metaItem struct {
	id      string
	title   string
	desc    string
	current bool
}

func parseModelCommand(line string) (active bool, filter string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "/model" {
		return true, ""
	}
	if strings.HasPrefix(trimmed, "/model ") {
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/model"))
	}
	return false, ""
}

func parseChatsCommand(line string) (active bool, filter string) {
	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "/chats":
		return true, ""
	case strings.HasPrefix(trimmed, "/chats "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/chats"))
	case trimmed == "/resume":
		return true, ""
	case strings.HasPrefix(trimmed, "/resume "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/resume"))
	case trimmed == "/load":
		return true, ""
	case strings.HasPrefix(trimmed, "/load "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/load"))
	default:
		return false, ""
	}
}

func (m *Model) syncMetaOverlay() {
	if active, filter := parseModelCommand(m.textinput.Value()); active {
		m.syncModelOverlay(filter)
		return
	}
	if active, filter := parseChatsCommand(m.textinput.Value()); active {
		m.syncChatOverlay(filter)
		return
	}
	m.metaOverlay = metaNone
	m.metaItems = nil
	m.metaSelected = 0
	m.metaScroll = 0
}

func (m *Model) syncModelOverlay(filter string) {
	picks := m.repl.SearchModels(filter, 50)
	if len(picks) == 0 && !m.repl.ModelsReady() {
		m.metaOverlay = metaModelPicker
		m.metaItems = []metaItem{{id: "", title: "Loading models…", desc: ""}}
		m.metaSelected = 0
		m.metaScroll = 0
		return
	}

	items := make([]metaItem, len(picks))
	for i, p := range picks {
		items[i] = metaItem{
			id:      p.ID,
			title:   p.ID,
			desc:    p.Kind,
			current: p.Current,
		}
	}
	m.metaOverlay = metaModelPicker
	m.metaItems = items
	if m.metaSelected >= len(m.metaItems) {
		m.metaSelected = 0
	}
	m.clampMetaScroll()
}

func (m *Model) syncChatOverlay(filter string) {
	opening := m.metaOverlay != metaChatPicker
	prevID := ""
	if !opening && m.metaSelected < len(m.metaItems) {
		prevID = m.metaItems[m.metaSelected].id
	}

	picks := m.repl.SearchChats(filter, 50)
	if len(picks) == 0 && !m.repl.ChatsReady() {
		m.metaOverlay = metaChatPicker
		m.metaItems = []metaItem{{id: "", title: "Loading chats…", desc: ""}}
		m.metaSelected = 0
		m.metaScroll = 0
		return
	}
	if len(picks) == 0 {
		m.metaOverlay = metaChatPicker
		m.metaItems = []metaItem{{id: "", title: "No chats found", desc: "keep typing to filter"}}
		m.metaSelected = 0
		m.metaScroll = 0
		return
	}

	items := make([]metaItem, len(picks))
	for i, p := range picks {
		items[i] = metaItem{
			id:      p.ID,
			title:   p.Title,
			desc:    p.ShortID,
			current: p.Current,
		}
	}
	m.metaOverlay = metaChatPicker
	m.metaItems = items
	m.metaSelected = 0
	for i, item := range items {
		if item.id != "" && item.id == prevID {
			m.metaSelected = i
			break
		}
	}
	m.clampMetaScroll()
}

func (m *Model) clampMetaScroll() {
	if len(m.metaItems) == 0 {
		m.metaScroll = 0
		return
	}
	visible := metaPickerVisibleRows(len(m.metaItems))
	if m.metaSelected < m.metaScroll {
		m.metaScroll = m.metaSelected
	}
	if m.metaSelected >= m.metaScroll+visible {
		m.metaScroll = m.metaSelected - visible + 1
	}
}

func metaPickerVisibleRows(total int) int {
	if total < metaPickerMaxRows {
		return total
	}
	return metaPickerMaxRows
}

func (m *Model) metaVisibleCount() int {
	if m.metaOverlay == metaNone || len(m.metaItems) == 0 {
		return 0
	}
	return metaPickerVisibleRows(len(m.metaItems))
}

func (m *Model) confirmMetaPick() {
	switch m.metaOverlay {
	case metaModelPicker:
		m.confirmModelPick()
	case metaChatPicker:
		m.confirmChatPick()
	}
}

func (m *Model) confirmModelPick() {
	if m.metaOverlay != metaModelPicker || len(m.metaItems) == 0 {
		return
	}
	item := m.metaItems[m.metaSelected]
	if item.id == "" {
		return
	}

	m.repl.SetModelID(item.id)
	m.textinput.SetValue("")
	m.metaOverlay = metaNone
	m.metaItems = nil
	m.metaSelected = 0
	m.metaScroll = 0
	m.showSuggestions = false
	m.suggestions = nil

	msgCount := len(m.repl.SessionMessages())
	m.appendInfo(fmt.Sprintf("Using %s · %s", item.id, preservedMessagesPhrase(msgCount)))
	m.layout()
	m.refreshViewportForce(true)
}

func (m *Model) confirmChatPick() {
	if m.metaOverlay != metaChatPicker || len(m.metaItems) == 0 {
		return
	}
	item := m.metaItems[m.metaSelected]
	if item.id == "" {
		return
	}

	result := m.repl.ResumeChatByID(item.id)

	m.metaOverlay = metaNone
	m.metaItems = nil
	m.metaSelected = 0
	m.metaScroll = 0
	m.showSuggestions = false
	m.suggestions = nil
	m.textinput.SetValue("")

	if result.Err != nil {
		m.appendError(result.Err.Error())
	} else if result.ReloadMessages {
		m.reloadFromSession()
		if len(m.messages) == 0 {
			m.appendError("chat resumed but no messages were found")
		} else if result.Output != "" {
			m.appendInfo(result.Output)
		}
	} else if result.Output != "" {
		m.appendInfo(result.Output)
	}

	m.layout()
	m.refreshViewportForce(true)
	m.viewport.GotoBottom()
}

func (m *Model) cancelMetaOverlay() {
	m.textinput.SetValue("")
	m.metaOverlay = metaNone
	m.metaItems = nil
	m.metaSelected = 0
	m.metaScroll = 0
	m.showSuggestions = false
	m.suggestions = nil
	m.layout()
}

func renderMetaOverlay(kind metaOverlayKind, items []metaItem, selected, scroll, width int) string {
	if kind == metaNone || len(items) == 0 {
		return ""
	}

	var b strings.Builder
	switch kind {
	case metaModelPicker:
		b.WriteString(metaTitleStyle.Render(pipeStyle.Render("│ ") + "Switch model"))
		b.WriteString("\n")
		b.WriteString(metaHintStyle.Render(pipeStyle.Render("│ ") + "↑↓ navigate · Enter confirm · Esc cancel"))
	case metaChatPicker:
		b.WriteString(metaTitleStyle.Render(pipeStyle.Render("│ ") + "Resume chat"))
		b.WriteString("\n")
		b.WriteString(metaHintStyle.Render(pipeStyle.Render("│ ") + "↑↓ navigate · Enter resume · Esc cancel"))
	}
	b.WriteString("\n")

	visible := metaPickerVisibleRows(len(items))
	end := scroll + visible
	if end > len(items) {
		end = len(items)
	}

	for i := scroll; i < end; i++ {
		item := items[i]
		cursor := pipeStyle.Render("│ ")
		if i == selected {
			cursor = metaCursorStyle.Render("❯ ")
		}
		title := item.title
		if item.current {
			title += metaCurrentStyle.Render("  ← active")
		}
		desc := ""
		if item.desc != "" {
			desc = metaDescStyle.Render(item.desc)
		}
		descW := lipgloss.Width(desc)
		maxTitle := width - lipgloss.Width(cursor) - descW - 4
		if maxTitle < 8 {
			maxTitle = 8
		}
		line := cursor + metaItemStyle.Render(truncateDisplay(title, maxTitle))
		if desc != "" {
			gap := width - lipgloss.Width(line) - descW
			if gap < 1 {
				gap = 1
			}
			line += strings.Repeat(" ", gap) + desc
		}
		if i == selected {
			line = metaSelectedRowStyle.Width(width).Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	if len(items) > visible {
		fmt.Fprintf(&b, "%s  %d–%d of %d\n",
			metaHintStyle.Render(""),
			scroll+1, end, len(items))
	}
	return strings.TrimRight(b.String(), "\n")
}