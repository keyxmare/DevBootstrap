// Package tui provides the terminal UI using Bubble Tea.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette - soft minimal
var (
	colorAccent    = lipgloss.Color("#7dcfff") // soft cyan
	colorSuccess   = lipgloss.Color("#9ece6a") // soft green
	colorError     = lipgloss.Color("#f7768e") // soft red
	colorWarning   = lipgloss.Color("#e0af68") // soft yellow
	colorMuted     = lipgloss.Color("#565f89") // muted gray
	colorDim       = lipgloss.Color("#3b4261") // dim gray
	colorText      = lipgloss.Color("#c0caf5") // light text
	colorSubtle    = lipgloss.Color("#a9b1d6") // subtle text
	colorBg        = lipgloss.Color("#1a1b26") // dark background
	colorBgSubtle  = lipgloss.Color("#24283b") // subtle background
)

// Styles for the TUI
type Styles struct {
	// Layout
	App      lipgloss.Style
	Header   lipgloss.Style
	Content  lipgloss.Style
	Footer   lipgloss.Style

	// Header elements
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Divider  lipgloss.Style

	// Section
	SectionTitle lipgloss.Style
	SectionLine  lipgloss.Style

	// App list
	AppInstalled    lipgloss.Style
	AppNotInstalled lipgloss.Style
	AppName         lipgloss.Style
	AppNameDim      lipgloss.Style
	AppVersion      lipgloss.Style
	AppDescription  lipgloss.Style
	AppIcon         lipgloss.Style
	AppIconDim      lipgloss.Style

	// Info display
	Label lipgloss.Style
	Value lipgloss.Style

	// Status messages
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style

	// Interactive elements
	SelectedItem   lipgloss.Style
	UnselectedItem lipgloss.Style
	Cursor         lipgloss.Style

	// Misc
	Muted   lipgloss.Style
	Counter lipgloss.Style
	Help    lipgloss.Style
}

// NewStyles creates the styles for a given width.
func NewStyles(width int) Styles {
	// Calculate content width (centered with padding)
	contentWidth := width
	if contentWidth > 90 {
		contentWidth = 90
	}

	return Styles{
		// Layout
		App: lipgloss.NewStyle().
			Width(width).
			Padding(1, 0),

		Header: lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Center).
			Padding(1, 0).
			MarginBottom(1),

		Content: lipgloss.NewStyle().
			Width(contentWidth).
			Padding(0, 4),

		Footer: lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Center).
			Foreground(colorMuted).
			Padding(1, 0),

		// Header elements
		Title: lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true).
			MarginBottom(0),

		Subtitle: lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true),

		Divider: lipgloss.NewStyle().
			Foreground(colorDim),

		// Section
		SectionTitle: lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			MarginTop(1).
			MarginBottom(1),

		SectionLine: lipgloss.NewStyle().
			Foreground(colorDim),

		// App list
		AppInstalled: lipgloss.NewStyle().
			Foreground(colorSuccess),

		AppNotInstalled: lipgloss.NewStyle().
			Foreground(colorDim),

		AppName: lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true),

		AppNameDim: lipgloss.NewStyle().
			Foreground(colorSubtle),

		AppVersion: lipgloss.NewStyle().
			Foreground(colorMuted),

		AppDescription: lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginLeft(7),

		AppIcon: lipgloss.NewStyle().
			Foreground(colorSuccess).
			MarginRight(2),

		AppIconDim: lipgloss.NewStyle().
			Foreground(colorDim).
			MarginRight(2),

		// Info display
		Label: lipgloss.NewStyle().
			Foreground(colorAccent).
			Width(14),

		Value: lipgloss.NewStyle().
			Foreground(colorSubtle),

		// Status messages
		Success: lipgloss.NewStyle().
			Foreground(colorSuccess),

		Error: lipgloss.NewStyle().
			Foreground(colorError),

		Warning: lipgloss.NewStyle().
			Foreground(colorWarning),

		Info: lipgloss.NewStyle().
			Foreground(colorAccent),

		// Interactive elements
		SelectedItem: lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true),

		UnselectedItem: lipgloss.NewStyle().
			Foreground(colorSubtle),

		Cursor: lipgloss.NewStyle().
			Foreground(colorAccent),

		// Misc
		Muted: lipgloss.NewStyle().
			Foreground(colorMuted),

		Counter: lipgloss.NewStyle().
			Foreground(colorMuted).
			Align(lipgloss.Right),

		Help: lipgloss.NewStyle().
			Foreground(colorDim),
	}
}
