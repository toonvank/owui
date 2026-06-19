package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/toonvank/owui/internal/repl"
)

type chatMessage struct {
	role      string
	content   string
	model     string
	streaming bool
}

type suggestionItem struct {
	text        string
	description string
}

type streamEvent struct {
	delta string
	done  bool
	reply string
	err   error
}

type streamEventMsg streamEvent

type streamStartedMsg struct {
	ch chan streamEvent
}

type thinkingTickMsg time.Time

type readyMsg struct{}

const (
	inputHeight      = 1
	headerHeight     = 3 // title + status + divider
	footerHeight     = 1
	statusHeight     = 1
	dividerHeight    = 1
	suggestMaxRows   = 6
	viewportTop      = headerHeight
	defaultWidth     = 100
	defaultHeight    = 30
	streamRefreshMin = 50 * time.Millisecond
)

type Model struct {
	repl *repl.REPL

	width  int
	height int

	viewport  viewport.Model
	textinput textinput.Model
	spinner   spinner.Model
	collapse  *collapseState
	cache     *renderCache

	messages           []chatMessage
	chatActive         bool
	thinkingSince      time.Time
	streamChars        int
	streaming          bool
	streamFollow       bool
	streamIdx          int
	streamCh           chan streamEvent
	lastStreamRefresh  time.Time
	suggestions        []suggestionItem
	selectedSuggestion int
	showSuggestions    bool
	showHelp           bool
	ready              bool
	hydrated           bool

	metaOverlay  metaOverlayKind
	metaItems    []metaItem
	metaSelected int
	metaScroll   int
}

func New(r *repl.REPL) Model {
	ti := textinput.New()
	ti.Placeholder = "Ask anything…  (/model · ? help)"
	ti.CharLimit = 32000
	ti.Width = defaultWidth - 8
	ti.Prompt = "❯ "
	ti.PromptStyle = chevronStyle
	ti.TextStyle = textStyle
	ti.PlaceholderStyle = mutedStyle
	ti.Cursor.Style = cursorStyle
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	vp := viewport.New(defaultWidth, 20)
	vp.SetContent("")

	return Model{
		repl:      r,
		width:     defaultWidth,
		height:    defaultHeight,
		textinput: ti,
		spinner:   sp,
		viewport:  vp,
		collapse:  newCollapseState(),
		cache:     newRenderCache(),
	}
}

func (m *Model) hydrateFromSession() {
	model := m.repl.CurrentModel()
	for i, msg := range m.repl.SessionMessages() {
		role := msg.Role
		if role == "" {
			role = "system"
		}
		if role == "system" && i == 0 {
			continue
		}
		cm := chatMessage{role: role, content: msg.Content}
		if role == "assistant" {
			cm.model = model
		}
		m.messages = append(m.messages, cm)
	}
	m.hydrated = true
}

func (m *Model) reloadFromSession() {
	m.messages = nil
	m.hydrated = false
	m.hydrateFromSession()
	m.collapse = newCollapseState()
	m.cache.clear()
}

func (m Model) Init() tea.Cmd {
	PrewarmMarkdown(defaultWidth)
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
		tea.WindowSize(),
		m.deferredReady(),
	)
}

func (m Model) deferredReady() tea.Cmd {
	return func() tea.Msg { return readyMsg{} }
}

func (m *Model) layout() {
	if m.width == 0 || m.height == 0 {
		m.width = defaultWidth
		m.height = defaultHeight
	}

	overlayH := 0
	if m.metaOverlay != metaNone {
		overlayH = m.metaVisibleCount() + 2 // title + hint rows
	} else if m.showSuggestions && len(m.suggestions) > 0 {
		n := len(m.suggestions)
		if n > suggestMaxRows {
			n = suggestMaxRows
		}
		overlayH = n
	}

	activeStatusH := 0
	if m.chatActive {
		activeStatusH = statusHeight
	}

	chatH := m.height - headerHeight - footerHeight - inputHeight - activeStatusH - overlayH - dividerHeight
	if chatH < 4 {
		chatH = 4
	}

	m.viewport.Width = m.width
	m.viewport.Height = chatH
	m.textinput.Width = m.width - 8
}

func (m *Model) refreshViewport() {
	m.refreshViewportForce(false)
}

func (m *Model) refreshViewportForce(force bool) {
	if m.showHelp {
		m.viewport.SetContent(repl.ShortcutsPanel())
		m.viewport.GotoTop()
		return
	}

	if m.streaming && !force {
		if since := time.Since(m.lastStreamRefresh); since < streamRefreshMin {
			return
		}
	}
	m.lastStreamRefresh = time.Now()

	atBottom := m.viewport.AtBottom()

	content := renderChatLog(m.messages, m.width, m.collapse, m.cache, m.repl.CurrentModel())
	m.viewport.SetContent(content)

	if (m.streaming && m.streamFollow) || atBottom {
		m.viewport.GotoBottom()
	}
}

func (m *Model) syncSuggestions() {
	if m.metaOverlay != metaNone {
		m.showSuggestions = false
		m.suggestions = nil
		return
	}
	line := m.textinput.Value()
	if line == "" || line == "?" {
		m.showSuggestions = false
		m.suggestions = nil
		return
	}
	if strings.HasPrefix(line, "/model") || strings.HasPrefix(line, "/chats") ||
		strings.HasPrefix(line, "/resume") || strings.HasPrefix(line, "/load") ||
		strings.HasPrefix(line, "/sessions") || strings.HasPrefix(line, "/session load") ||
		strings.HasPrefix(line, "/knowledge") || strings.HasPrefix(line, "/kb") ||
		strings.HasPrefix(line, "/filters") || strings.HasPrefix(line, "/functions") ||
		strings.HasPrefix(line, "/tools") ||
		strings.HasPrefix(line, "/profile") {
		m.showSuggestions = false
		m.suggestions = nil
		return
	}
	if !strings.HasPrefix(line, "/") && !strings.HasPrefix(line, "@") {
		m.showSuggestions = false
		m.suggestions = nil
		return
	}

	sugs := m.repl.GetSuggestions(line)
	if len(sugs) == 0 {
		m.showSuggestions = false
		m.suggestions = nil
		m.selectedSuggestion = 0
		return
	}

	m.suggestions = make([]suggestionItem, len(sugs))
	for i, s := range sugs {
		m.suggestions[i] = suggestionItem{text: s.Text, description: s.Description}
	}
	if m.selectedSuggestion >= len(m.suggestions) {
		m.selectedSuggestion = 0
	}
	m.showSuggestions = true
}

func (m *Model) acceptSuggestion() {
	if !m.showSuggestions || len(m.suggestions) == 0 {
		return
	}
	s := m.suggestions[m.selectedSuggestion]
	m.textinput.SetValue(s.text)
	m.showSuggestions = false
	m.suggestions = nil
	m.selectedSuggestion = 0
}

func (m *Model) appendSystem(text string) {
	if text == "" {
		return
	}
	m.messages = append(m.messages, chatMessage{role: "system", content: text})
	m.cache.invalidateFrom(len(m.messages) - 1)
}

func (m *Model) appendInfo(text string) {
	if text == "" {
		return
	}
	m.messages = append(m.messages, chatMessage{role: "info", content: text})
	m.cache.invalidateFrom(len(m.messages) - 1)
}

func (m *Model) appendError(text string) {
	m.messages = append(m.messages, chatMessage{role: "error", content: text})
	m.cache.invalidateFrom(len(m.messages) - 1)
}

func (m *Model) beginChat(prompt string) tea.Cmd {
	m.messages = append(m.messages, chatMessage{role: "user", content: prompt})
	m.cache.invalidateFrom(len(m.messages) - 1)
	m.chatActive = true
	m.thinkingSince = time.Now()
	m.streamChars = 0
	m.streamFollow = true
	m.streaming = m.repl.StreamEnabled()
	if m.streaming {
		m.messages = append(m.messages, chatMessage{
			role:      "assistant",
			content:   "",
			model:     m.repl.CurrentModel(),
			streaming: true,
		})
		m.streamIdx = len(m.messages) - 1
	}
	m.layout()
	m.refreshViewportForce(true)
	return tea.Batch(m.spinner.Tick, m.thinkingTick(), m.startStream(prompt))
}

func (m Model) thinkingTick() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return thinkingTickMsg(t)
	})
}

func (m Model) startStream(prompt string) tea.Cmd {
	ch := make(chan streamEvent, 64)
	return func() tea.Msg {
		go func() {
			streaming := m.repl.StreamEnabled()
			var onDelta func(string)
			if streaming {
				onDelta = func(d string) {
					ch <- streamEvent{delta: d}
				}
			}
			reply, err := m.repl.ChatUserMessage(prompt, onDelta)
			ch <- streamEvent{done: true, reply: reply, err: err}
			close(ch)
		}()
		return streamStartedMsg{ch: ch}
	}
}

func (m Model) listenStream() tea.Cmd {
	ch := m.streamCh
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return streamEventMsg{done: true}
		}
		return streamEventMsg(ev)
	}
}

func (m *Model) handleStreamEvent(ev streamEvent) tea.Cmd {
	if !ev.done && ev.delta != "" {
		if m.streamIdx >= 0 && m.streamIdx < len(m.messages) {
			m.messages[m.streamIdx].content += ev.delta
			m.streamChars += len(ev.delta)
		}
		m.refreshViewport()
		return m.listenStream()
	}

	m.chatActive = false
	m.layout()
	if ev.err != nil {
		if m.streaming && m.streamIdx >= 0 {
			m.messages = m.messages[:m.streamIdx]
		}
		m.appendError(ev.err.Error())
		m.streaming = false
		m.streamIdx = -1
		m.streamCh = nil
		m.refreshViewportForce(true)
		return nil
	}

	if m.streaming && m.streamIdx >= 0 {
		m.messages[m.streamIdx].content = ev.reply
		m.messages[m.streamIdx].streaming = false
		if m.messages[m.streamIdx].model == "" {
			m.messages[m.streamIdx].model = m.repl.CurrentModel()
		}
		m.cache.invalidateFrom(m.streamIdx)
	} else {
		m.messages = append(m.messages, chatMessage{
			role:    "assistant",
			content: ev.reply,
			model:   m.repl.CurrentModel(),
		})
	}
	m.streaming = false
	m.streamIdx = -1
	m.streamCh = nil
	m.refreshViewportForce(true)
	return nil
}

func (m *Model) handleMouseClick(y int) {
	if m.showHelp {
		return
	}
	row := y - viewportTop
	if row < 0 || row >= m.viewport.Height {
		return
	}
	contentLine := m.viewport.YOffset + row
	if key, ok := m.collapse.hitAt(contentLine); ok {
		m.collapse.toggleByKey(key)
		m.cache.clear()
		m.refreshViewportForce(true)
	}
}

func (m *Model) toggleLastAssistantBlocks() {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].role == "assistant" && m.messages[i].content != "" {
			m.collapse.toggleAllInMessage(i, ParseBlocks(m.messages[i].content))
			m.cache.clear()
			m.refreshViewportForce(true)
			return
		}
	}
}

func (m *Model) submitLine(line string) tea.Cmd {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	if line == "?" {
		m.showHelp = !m.showHelp
		m.refreshViewportForce(true)
		return nil
	}

	m.showHelp = false

	if strings.HasPrefix(line, "/") {
		result := m.repl.RunSlashCommand(line)
		if result.Quit {
			return tea.Quit
		}
		if result.Err != nil {
			m.appendError(result.Err.Error())
		} else if result.Output != "" {
			m.appendSystem(result.Output)
		}
		if result.Cleared {
			m.messages = nil
			m.collapse = newCollapseState()
			m.cache.clear()
		} else if result.ReloadMessages {
			m.reloadFromSession()
		}
		if result.ResendPrompt != "" {
			m.refreshViewportForce(true)
			return m.beginChat(result.ResendPrompt)
		}
		if result.ScrollToMsg >= 0 {
			m.scrollToMessage(result.ScrollToMsg)
		}
		m.refreshViewportForce(true)
		return nil
	}

	prompt, modelChanged, err := m.repl.ParseAtPrefix(line)
	if err != nil {
		m.appendError(err.Error())
		m.refreshViewportForce(true)
		return nil
	}
	if modelChanged != "" {
		m.appendInfo("Using " + modelChanged)
	}
	if prompt == "" {
		m.refreshViewportForce(true)
		return nil
	}

	m.repl.PushInputHistory(prompt)
	return m.beginChat(prompt)
}

func (m *Model) scrollToMessage(msgIdx int) {
	if msgIdx < 0 || msgIdx >= len(m.messages) {
		return
	}
	content := renderChatLog(m.messages, m.width, m.collapse, m.cache, m.repl.CurrentModel())
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return
	}
	target := (msgIdx * len(lines)) / len(m.messages)
	if target >= len(lines) {
		target = len(lines) - 1
	}
	m.viewport.SetYOffset(target)
}

func (m Model) elapsedThinking() float64 {
	if m.thinkingSince.IsZero() {
		return 0
	}
	return time.Since(m.thinkingSince).Seconds()
}