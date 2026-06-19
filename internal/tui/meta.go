package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/toonvank/owui/internal/repl"
)

type metaOverlayKind int

const (
	metaNone metaOverlayKind = iota
	metaModelPicker
	metaChatPicker
	metaSessionPicker
	metaKnowledgePicker
	metaFilterPicker
	metaToolPicker
	metaProfilePicker
)

const metaPickerMaxRows = 12

type metaItem struct {
	id      string
	title   string
	desc    string
	current bool
	enabled bool
	toggle  bool
	pinned  bool
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

func parseProfileCommand(line string) (active bool, filter string) {
	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "/profile":
		return true, ""
	case strings.HasPrefix(trimmed, "/profile "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/profile"))
	default:
		return false, ""
	}
}

func parseFiltersCommand(line string) (active bool, filter string) {
	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "/filters", trimmed == "/functions":
		return true, ""
	case strings.HasPrefix(trimmed, "/filters "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/filters"))
	case strings.HasPrefix(trimmed, "/functions "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/functions"))
	default:
		return false, ""
	}
}

func parseToolsCommand(line string) (active bool, filter string) {
	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "/tools":
		return true, ""
	case strings.HasPrefix(trimmed, "/tools "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/tools"))
	default:
		return false, ""
	}
}

func parseKnowledgeCommand(line string) (active bool, filter string) {
	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "/knowledge", trimmed == "/kb":
		return true, ""
	case strings.HasPrefix(trimmed, "/knowledge "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/knowledge"))
	case strings.HasPrefix(trimmed, "/kb "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/kb"))
	default:
		return false, ""
	}
}

func parseSessionsCommand(line string) (active bool, filter string) {
	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "/sessions":
		return true, ""
	case strings.HasPrefix(trimmed, "/sessions "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/sessions"))
	case trimmed == "/session load":
		return true, ""
	case strings.HasPrefix(trimmed, "/session load "):
		return true, strings.TrimSpace(strings.TrimPrefix(trimmed, "/session load "))
	default:
		return false, ""
	}
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
	if active, filter := parseSessionsCommand(m.textinput.Value()); active {
		m.syncSessionOverlay(filter)
		return
	}
	if active, filter := parseKnowledgeCommand(m.textinput.Value()); active {
		m.syncKnowledgeOverlay(filter)
		return
	}
	if active, filter := parseProfileCommand(m.textinput.Value()); active {
		m.syncProfileOverlay(filter)
		return
	}
	if active, filter := parseFiltersCommand(m.textinput.Value()); active {
		m.syncFilterOverlay(filter)
		return
	}
	if active, filter := parseToolsCommand(m.textinput.Value()); active {
		m.syncToolOverlay(filter)
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
		title := p.Title
		if p.Pinned {
			title = "📌 " + title
		}
		items[i] = metaItem{
			id:      p.ID,
			title:   title,
			desc:    p.ShortID,
			current: p.Current,
			pinned:  p.Pinned,
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

func (m *Model) syncSessionOverlay(filter string) {
	opening := m.metaOverlay != metaSessionPicker
	prevID := ""
	if !opening && m.metaSelected < len(m.metaItems) {
		prevID = m.metaItems[m.metaSelected].id
	}

	picks := m.repl.SearchLocalSessions(filter, 50)
	if len(picks) == 0 {
		m.metaOverlay = metaSessionPicker
		m.metaItems = []metaItem{{id: "", title: "No local sessions", desc: "start chatting to create one"}}
		m.metaSelected = 0
		m.metaScroll = 0
		return
	}

	items := make([]metaItem, len(picks))
	for i, p := range picks {
		desc := fmt.Sprintf("%d msgs", p.MsgCount)
		if p.Age != "" {
			desc += " · " + p.Age
		}
		items[i] = metaItem{
			id:      p.ID,
			title:   p.Title,
			desc:    desc,
			current: p.Current,
		}
	}
	m.metaOverlay = metaSessionPicker
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

func (m *Model) confirmMetaPick() {
	switch m.metaOverlay {
	case metaModelPicker:
		m.confirmModelPick()
	case metaChatPicker:
		m.confirmChatPick()
	case metaSessionPicker:
		m.confirmSessionPick()
	case metaKnowledgePicker:
		m.confirmKnowledgePick()
	case metaProfilePicker:
		m.confirmProfilePick()
	case metaFilterPicker, metaToolPicker:
		m.toggleMetaPick()
	}
}

func (m *Model) syncProfileOverlay(filter string) {
	opening := m.metaOverlay != metaProfilePicker
	prevID := ""
	if !opening && m.metaSelected < len(m.metaItems) {
		prevID = m.metaItems[m.metaSelected].id
	}

	picks, err := m.repl.SearchProfiles(filter)
	if err != nil {
		m.metaOverlay = metaProfilePicker
		m.metaItems = []metaItem{{id: "", title: "Error loading profiles", desc: err.Error()}}
		m.metaSelected = 0
		m.metaScroll = 0
		return
	}
	if len(picks) == 0 {
		m.metaOverlay = metaProfilePicker
		m.metaItems = []metaItem{{id: "", title: "No profiles", desc: "owui config profile add <name>"}}
		m.metaSelected = 0
		m.metaScroll = 0
		return
	}

	items := make([]metaItem, len(picks))
	for i, p := range picks {
		desc := p.URL
		if len(desc) > 36 {
			desc = desc[:33] + "..."
		}
		items[i] = metaItem{
			id:      p.Name,
			title:   p.Name,
			desc:    desc,
			current: p.Current,
		}
	}
	m.metaOverlay = metaProfilePicker
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

func (m *Model) confirmProfilePick() {
	if m.metaOverlay != metaProfilePicker || len(m.metaItems) == 0 {
		return
	}
	item := m.metaItems[m.metaSelected]
	if item.id == "" {
		return
	}

	result := m.repl.SwitchProfile(item.id)

	m.metaOverlay = metaNone
	m.metaItems = nil
	m.metaSelected = 0
	m.metaScroll = 0
	m.showSuggestions = false
	m.suggestions = nil
	m.textinput.SetValue("")

	if result.Err != nil {
		m.appendError(result.Err.Error())
	} else {
		if result.Cleared {
			m.messages = nil
			m.collapse = newCollapseState()
			m.cache.clear()
		}
		if result.ReloadMessages {
			m.reloadFromSession()
		}
		if result.Output != "" {
			m.appendInfo(result.Output)
		}
	}
	m.layout()
	m.refreshViewportForce(true)
}

func (m *Model) isToggleOverlay() bool {
	return m.metaOverlay == metaFilterPicker || m.metaOverlay == metaToolPicker
}

func (m *Model) syncFilterOverlay(filter string) {
	m.syncFunctionOverlay(metaFilterPicker, filter, m.repl.SearchFilters)
}

func (m *Model) syncToolOverlay(filter string) {
	m.syncFunctionOverlay(metaToolPicker, filter, m.repl.SearchTools)
}

func (m *Model) syncFunctionOverlay(kind metaOverlayKind, filter string, search func(string, int) []repl.FunctionPick) {
	opening := m.metaOverlay != kind
	prevID := ""
	if !opening && m.metaSelected < len(m.metaItems) {
		prevID = m.metaItems[m.metaSelected].id
	}

	picks := search(filter, 50)
	if len(picks) == 0 && !m.repl.FunctionsReady() {
		m.metaOverlay = kind
		m.metaItems = []metaItem{{id: "", title: "Loading…", desc: ""}}
		m.metaSelected = 0
		m.metaScroll = 0
		return
	}
	if len(picks) == 0 {
		m.metaOverlay = kind
		m.metaItems = []metaItem{{id: "", title: "No matches", desc: "keep typing to filter"}}
		m.metaSelected = 0
		m.metaScroll = 0
		return
	}

	items := make([]metaItem, len(picks))
	for i, p := range picks {
		state := "off"
		if p.Enabled {
			state = "on"
		}
		items[i] = metaItem{
			id:      p.ID,
			title:   p.Name,
			desc:    fmt.Sprintf("[%s] %s", state, p.Scope),
			enabled: p.Enabled,
			toggle:  true,
		}
	}
	m.metaOverlay = kind
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

func (m *Model) toggleMetaPick() {
	if !m.isToggleOverlay() || len(m.metaItems) == 0 {
		return
	}
	item := m.metaItems[m.metaSelected]
	if item.id == "" {
		return
	}

	var enabled bool
	switch m.metaOverlay {
	case metaFilterPicker:
		enabled = m.repl.ToggleFilter(item.id)
		if active, filter := parseFiltersCommand(m.textinput.Value()); active {
			m.syncFilterOverlay(filter)
		}
	case metaToolPicker:
		enabled = m.repl.ToggleTool(item.id)
		if active, filter := parseToolsCommand(m.textinput.Value()); active {
			m.syncToolOverlay(filter)
		}
	}

	state := "off"
	if enabled {
		state = "on"
	}
	m.appendInfo(fmt.Sprintf("%s → %s", item.title, state))
	m.layout()
	m.refreshViewportForce(true)
}

func (m *Model) syncKnowledgeOverlay(filter string) {
	opening := m.metaOverlay != metaKnowledgePicker
	prevID := ""
	if !opening && m.metaSelected < len(m.metaItems) {
		prevID = m.metaItems[m.metaSelected].id
	}

	picks := m.repl.SearchKnowledge(filter, 50)
	if len(picks) == 0 && !m.repl.KnowledgeReady() {
		m.metaOverlay = metaKnowledgePicker
		m.metaItems = []metaItem{{id: "", title: "Loading collections…", desc: ""}}
		m.metaSelected = 0
		m.metaScroll = 0
		return
	}
	if len(picks) == 0 {
		m.metaOverlay = metaKnowledgePicker
		m.metaItems = []metaItem{{id: "", title: "No collections found", desc: "keep typing to filter"}}
		m.metaSelected = 0
		m.metaScroll = 0
		return
	}

	items := make([]metaItem, len(picks))
	for i, p := range picks {
		items[i] = metaItem{
			id:      p.ID,
			title:   p.Name,
			desc:    p.Desc,
			current: p.Current,
		}
	}
	m.metaOverlay = metaKnowledgePicker
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

func (m *Model) toggleChatPinPick() {
	if m.metaOverlay != metaChatPicker || len(m.metaItems) == 0 {
		return
	}
	item := m.metaItems[m.metaSelected]
	if item.id == "" {
		return
	}

	pinned, err := m.repl.TogglePickerChatPin(item.id)
	if err != nil {
		m.appendError(err.Error())
		return
	}

	state := "unpinned"
	if pinned {
		state = "pinned"
	}
	m.appendInfo(fmt.Sprintf("%s → %s", strings.TrimPrefix(item.title, "📌 "), state))

	if active, filter := parseChatsCommand(m.textinput.Value()); active {
		m.syncChatOverlay(filter)
	}
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

func (m *Model) confirmSessionPick() {
	if m.metaOverlay != metaSessionPicker || len(m.metaItems) == 0 {
		return
	}
	item := m.metaItems[m.metaSelected]
	if item.id == "" {
		return
	}

	result := m.repl.LoadLocalSessionByID(item.id)

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
		if result.Output != "" {
			m.appendInfo(result.Output)
		}
	} else if result.Output != "" {
		m.appendInfo(result.Output)
	}

	m.layout()
	m.refreshViewportForce(true)
	m.viewport.GotoBottom()
}

func (m *Model) confirmKnowledgePick() {
	if m.metaOverlay != metaKnowledgePicker || len(m.metaItems) == 0 {
		return
	}
	item := m.metaItems[m.metaSelected]
	if item.id == "" {
		return
	}

	m.repl.SetKnowledgeCollection(item.id, item.title)

	m.metaOverlay = metaNone
	m.metaItems = nil
	m.metaSelected = 0
	m.metaScroll = 0
	m.showSuggestions = false
	m.suggestions = nil
	m.textinput.SetValue("")

	m.appendInfo(fmt.Sprintf("Using knowledge collection %q", item.title))
	m.layout()
	m.refreshViewportForce(true)
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
		b.WriteString(metaHintStyle.Render(pipeStyle.Render("│ ") + "↑↓ navigate · Enter resume · p pin · Esc cancel"))
	case metaSessionPicker:
		b.WriteString(metaTitleStyle.Render(pipeStyle.Render("│ ") + "Local sessions"))
		b.WriteString("\n")
		b.WriteString(metaHintStyle.Render(pipeStyle.Render("│ ") + "↑↓ navigate · Enter load · Esc cancel"))
	case metaKnowledgePicker:
		b.WriteString(metaTitleStyle.Render(pipeStyle.Render("│ ") + "Knowledge collection"))
		b.WriteString("\n")
		b.WriteString(metaHintStyle.Render(pipeStyle.Render("│ ") + "↑↓ navigate · Enter select · Esc cancel"))
	case metaFilterPicker:
		b.WriteString(metaTitleStyle.Render(pipeStyle.Render("│ ") + "Toggle filters"))
		b.WriteString("\n")
		b.WriteString(metaHintStyle.Render(pipeStyle.Render("│ ") + "↑↓ navigate · Enter toggle · Esc done"))
	case metaToolPicker:
		b.WriteString(metaTitleStyle.Render(pipeStyle.Render("│ ") + "Toggle tools"))
		b.WriteString("\n")
		b.WriteString(metaHintStyle.Render(pipeStyle.Render("│ ") + "↑↓ navigate · Enter toggle · Esc done"))
	case metaProfilePicker:
		b.WriteString(metaTitleStyle.Render(pipeStyle.Render("│ ") + "Switch profile"))
		b.WriteString("\n")
		b.WriteString(metaHintStyle.Render(pipeStyle.Render("│ ") + "↑↓ navigate · Enter switch · Esc cancel"))
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