package tui

import "strings"

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(renderHeader(m.repl.BaseURL(), m.repl.CurrentModel(), m.repl.LocalSessionLabel(), m.width))
	b.WriteString("\n")
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	if m.chatActive {
		b.WriteString(renderStatusBar(m.width, m.streamChars == 0, m.elapsedThinking(), m.streamChars, m.spinner.View()))
		b.WriteString("\n")
	}

	if m.metaOverlay != metaNone {
		b.WriteString(renderMetaOverlay(m.metaOverlay, m.metaItems, m.metaSelected, m.metaScroll, m.width))
		b.WriteString("\n")
	} else if m.showSuggestions && len(m.suggestions) > 0 {
		b.WriteString(renderSuggestions(m.suggestions, m.selectedSuggestion, m.width))
		b.WriteString("\n")
	}

	b.WriteString(m.renderInput())
	b.WriteString("\n")
	b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")

	b.WriteString(renderFooter(m.repl.ShortcutBarLine(), m.width))

	return b.String()
}

func metaFilterText(overlay metaOverlayKind, line string) string {
	switch overlay {
	case metaModelPicker:
		if active, filter := parseModelCommand(line); active {
			return filter
		}
	case metaChatPicker:
		if active, filter := parseChatsCommand(line); active {
			return filter
		}
	}
	return ""
}

func (m Model) renderInput() string {
	style := inputStyle
	if m.textinput.Focused() {
		style = inputFocusedStyle
	}

	if m.metaOverlay != metaNone {
		filter := metaFilterText(m.metaOverlay, m.textinput.Value())
		if filter != "" {
			inner := chevronStyle.Render("❯ ") +
				textStyle.Render(filter) +
				cursorStyle.Render("▏")
			return style.Width(m.width).Render(inner)
		}
		hint := "↑↓ pick · Enter confirm · Esc cancel"
		if m.metaOverlay == metaChatPicker {
			hint = "↑↓ pick · Enter resume · Esc cancel"
		}
		return style.Width(m.width).Render(chevronStyle.Render("❯ ") + mutedStyle.Render(hint))
	}

	if m.textinput.Value() == "" {
		inner := chevronStyle.Render("❯ ") +
			cursorStyle.Render("▏") +
			mutedStyle.Render(m.textinput.Placeholder)
		return style.Width(m.width).Render(inner)
	}

	return style.Width(m.width).Render(m.textinput.View())
}