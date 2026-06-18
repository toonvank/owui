package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Hierarchy: bright = active, mid = content, dim = metadata.
	colorBrand   = lipgloss.Color("69")
	colorActive  = lipgloss.Color("255")
	colorMuted   = lipgloss.Color("245")
	colorDim     = lipgloss.Color("238")
	colorText    = lipgloss.Color("252")
	colorUser    = lipgloss.Color("117")
	colorAssist  = lipgloss.Color("114")
	colorDone    = lipgloss.Color("42")
	colorError   = lipgloss.Color("203")
	colorCode    = lipgloss.Color("252")
	colorCodeBG  = lipgloss.Color("236")
	colorCardBG  = lipgloss.Color("235")
	colorBorder  = lipgloss.Color("238")
	colorSelect  = lipgloss.Color("63")
	colorSpinner = lipgloss.Color("69")

	brandStyle   = lipgloss.NewStyle().Bold(true).Foreground(colorBrand)
	activeStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorActive)
	mutedStyle   = lipgloss.NewStyle().Foreground(colorMuted)
	dimStyle     = lipgloss.NewStyle().Foreground(colorDim)
	textStyle    = lipgloss.NewStyle().Foreground(colorText)
	doneStyle    = lipgloss.NewStyle().Foreground(colorDone)
	pipeStyle    = dimStyle
	chevronStyle = brandStyle

	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(colorBrand).Padding(0, 1)
	statusStyle = lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 1)

	userCardStyle = lipgloss.NewStyle().
			Background(colorCardBG).
			Foreground(colorActive).
			Padding(0, 1)

	assistBodyStyle = lipgloss.NewStyle().Foreground(colorText)
	infoLineStyle   = dimStyle

	errorStyle = lipgloss.NewStyle().Foreground(colorError).Bold(true)

	codeBlockStyle = lipgloss.NewStyle().
			Foreground(colorCode).
			Background(colorCodeBG).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Background(colorCardBG).
			Padding(0, 1)

	inputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBrand).
				Background(colorCardBG).
				Padding(0, 1)

	footerStyle = lipgloss.NewStyle().Foreground(colorDim).Padding(0, 1)

	suggestStyle = lipgloss.NewStyle().Foreground(colorText).Padding(0, 1)

	suggestSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(colorSelect).
				Padding(0, 1)

	suggestDescStyle = lipgloss.NewStyle().Foreground(colorMuted)

	helpTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorBrand)

	spinnerStyle = lipgloss.NewStyle().Foreground(colorSpinner)
	cursorStyle  = lipgloss.NewStyle().Foreground(colorBrand)

	dividerStyle = lipgloss.NewStyle().Foreground(colorDim)

	collapseHeaderStyle = lipgloss.NewStyle().Foreground(colorBrand).Bold(true)
	collapseTitleStyle  = mutedStyle

	metaTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorBrand)
	metaHintStyle  = dimStyle
	metaCursorStyle = chevronStyle
	metaItemStyle  = textStyle
	metaDescStyle  = mutedStyle
	metaCurrentStyle = lipgloss.NewStyle().Foreground(colorAssist).Italic(true)
	metaSelectedRowStyle = lipgloss.NewStyle().
				Background(colorSelect).
				Foreground(lipgloss.Color("255"))
)