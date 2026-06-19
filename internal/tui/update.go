package tui

import (
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		PrewarmMarkdown(m.width)
		m.refreshViewportForce(true)

	case readyMsg:
		if !m.ready {
			if !m.hydrated {
				m.hydrateFromSession()
			}
			m.ready = true
			m.layout()
			m.refreshViewportForce(true)
		}

	case streamStartedMsg:
		m.streamCh = msg.ch
		return m, m.listenStream()

	case streamEventMsg:
		cmd := m.handleStreamEvent(streamEvent(msg))
		if m.chatActive {
			cmds = append(cmds, m.thinkingTick())
		}
		return m, tea.Batch(cmd)

	case thinkingTickMsg:
		if m.chatActive {
			cmds = append(cmds, m.thinkingTick())
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			m.handleMouseClick(int(msg.Y))
			return m, nil
		}
		prevBottom := m.viewport.AtBottom()
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		if m.streaming && prevBottom && !m.viewport.AtBottom() {
			m.streamFollow = false
		}
		return m, cmd

	case tea.KeyMsg:
		if m.showHelp && msg.String() != "?" {
			m.showHelp = false
			m.refreshViewportForce(true)
		}

		// Meta overlay captures navigation before the text input.
		if m.metaOverlay == metaModelPicker || m.metaOverlay == metaChatPicker ||
			m.metaOverlay == metaSessionPicker || m.metaOverlay == metaKnowledgePicker ||
			m.metaOverlay == metaProfilePicker ||
			m.metaOverlay == metaFilterPicker || m.metaOverlay == metaToolPicker {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.cancelMetaOverlay()
				return m, nil
			case "up":
				if len(m.metaItems) > 0 {
					m.metaSelected--
					if m.metaSelected < 0 {
						m.metaSelected = len(m.metaItems) - 1
					}
					m.clampMetaScroll()
				}
				return m, nil
			case "down":
				if len(m.metaItems) > 0 {
					m.metaSelected++
					if m.metaSelected >= len(m.metaItems) {
						m.metaSelected = 0
					}
					m.clampMetaScroll()
				}
				return m, nil
			case "enter", "tab":
				m.confirmMetaPick()
				return m, nil
			case "p":
				if m.metaOverlay == metaChatPicker {
					m.toggleChatPinPick()
					return m, nil
				}
			default:
				// Allow typing to filter while picker is open.
				if msg.Type == tea.KeyRunes {
					var cmd tea.Cmd
					m.textinput, cmd = m.textinput.Update(msg)
					prevOverlay := m.metaOverlay
					m.syncMetaOverlay()
					if m.metaOverlay != prevOverlay {
						m.layout()
					}
					return m, cmd
				}
				if msg.String() == "backspace" || msg.String() == "ctrl+h" {
					var cmd tea.Cmd
					m.textinput, cmd = m.textinput.Update(msg)
					m.syncMetaOverlay()
					return m, cmd
				}
			}
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+u":
			m.cancelMetaOverlay()
			return m, nil
		case "esc":
			if m.showSuggestions {
				m.showSuggestions = false
				m.suggestions = nil
				m.layout()
				return m, nil
			}
			if m.showHelp {
				m.showHelp = false
				m.refreshViewportForce(true)
				return m, nil
			}
		case "pgup", "pgdown", "home", "end":
			if m.streaming && (msg.String() == "pgup" || msg.String() == "home") {
				m.streamFollow = false
			}
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		case "c":
			// Only when input is empty — otherwise "c" must reach the text field.
			if m.metaOverlay == metaNone && !m.showSuggestions && !m.chatActive && m.textinput.Value() == "" {
				m.toggleLastAssistantBlocks()
				return m, nil
			}
		case "up":
			if m.showSuggestions && len(m.suggestions) > 0 {
				m.selectedSuggestion--
				if m.selectedSuggestion < 0 {
					m.selectedSuggestion = len(m.suggestions) - 1
				}
				return m, nil
			}
			if m.metaOverlay == metaNone && !m.chatActive && m.textinput.Value() == "" {
				if recalled := m.repl.RecallInput(-1); recalled != "" {
					m.textinput.SetValue(recalled)
					return m, nil
				}
				if last := m.repl.LastUserMessage(); last != "" {
					m.textinput.SetValue(last)
					return m, nil
				}
			}
		case "down":
			if m.showSuggestions && len(m.suggestions) > 0 {
				m.selectedSuggestion++
				if m.selectedSuggestion >= len(m.suggestions) {
					m.selectedSuggestion = 0
				}
				return m, nil
			}
			if m.metaOverlay == metaNone && !m.chatActive && m.repl.HistoryBrowsing() {
				m.textinput.SetValue(m.repl.RecallInput(1))
				return m, nil
			}
		case "j":
			if m.repl.VimKeysEnabled() && m.metaOverlay == metaNone && !m.chatActive && m.textinput.Value() == "" {
				m.viewport.LineDown(1)
				return m, nil
			}
		case "k":
			if m.repl.VimKeysEnabled() && m.metaOverlay == metaNone && !m.chatActive && m.textinput.Value() == "" {
				m.viewport.LineUp(1)
				return m, nil
			}
		case "tab":
			if m.showSuggestions && len(m.suggestions) > 0 {
				m.acceptSuggestion()
				m.layout()
				return m, nil
			}
		case "enter":
			if m.chatActive {
				return m, nil
			}
			line := m.textinput.Value()
			if active, _ := parseModelCommand(line); active {
				m.syncMetaOverlay()
				m.confirmMetaPick()
				return m, nil
			}
			if active, _ := parseChatsCommand(line); active {
				m.syncMetaOverlay()
				m.confirmMetaPick()
				return m, nil
			}
			if active, _ := parseSessionsCommand(line); active {
				m.syncMetaOverlay()
				m.confirmMetaPick()
				return m, nil
			}
			if active, _ := parseKnowledgeCommand(line); active {
				m.syncMetaOverlay()
				m.confirmMetaPick()
				return m, nil
			}
			if active, _ := parseProfileCommand(line); active {
				m.syncMetaOverlay()
				m.confirmMetaPick()
				return m, nil
			}
			if active, _ := parseFiltersCommand(line); active {
				m.syncMetaOverlay()
				if m.isToggleOverlay() {
					m.toggleMetaPick()
				} else {
					m.confirmMetaPick()
				}
				return m, nil
			}
			if active, _ := parseToolsCommand(line); active {
				m.syncMetaOverlay()
				if m.isToggleOverlay() {
					m.toggleMetaPick()
				} else {
					m.confirmMetaPick()
				}
				return m, nil
			}
			m.textinput.SetValue("")
			m.metaOverlay = metaNone
			m.metaItems = nil
			m.showSuggestions = false
			m.suggestions = nil
			return m, m.submitLine(line)
		}

		var cmd tea.Cmd
		m.textinput, cmd = m.textinput.Update(msg)
		prevOverlay := m.metaOverlay
		prevSuggest := m.showSuggestions
		m.syncMetaOverlay()
		m.syncSuggestions()
		if m.metaOverlay != prevOverlay || m.showSuggestions != prevSuggest {
			m.layout()
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.chatActive {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	case cursor.BlinkMsg:
		var cmd tea.Cmd
		m.textinput, cmd = m.textinput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}