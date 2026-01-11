// Package tui provides the terminal user interface for DevBootstrap.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#06B6D4") // Cyan
	successColor   = lipgloss.Color("#10B981") // Green
	warningColor   = lipgloss.Color("#F59E0B") // Yellow
	errorColor     = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	textColor      = lipgloss.Color("#F9FAFB") // White
	dimColor       = lipgloss.Color("#9CA3AF") // Light gray
)

// Styles
var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			MarginBottom(1)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// List item styles
	ItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(primaryColor).
				Bold(true)

	CheckedStyle = lipgloss.NewStyle().
			Foreground(successColor)

	UncheckedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Status styles
	InstalledStyle = lipgloss.NewStyle().
			Foreground(successColor)

	NotInstalledStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	// Tag styles
	TagStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginRight(1)

	TagAppStyle = TagStyle.
			Foreground(lipgloss.Color("#3B82F6")). // Blue
			Background(lipgloss.Color("#1E3A5F"))

	TagConfigStyle = TagStyle.
			Foreground(lipgloss.Color("#A855F7")). // Purple
			Background(lipgloss.Color("#3B1F5C"))

	TagEditorStyle = TagStyle.
			Foreground(lipgloss.Color("#10B981")). // Green
			Background(lipgloss.Color("#134E3A"))

	TagShellStyle = TagStyle.
			Foreground(lipgloss.Color("#F59E0B")). // Yellow
			Background(lipgloss.Color("#5C4813"))

	TagContainerStyle = TagStyle.
				Foreground(lipgloss.Color("#EF4444")). // Red
				Background(lipgloss.Color("#5C1313"))

	TagFontStyle = TagStyle.
			Foreground(lipgloss.Color("#EC4899")). // Pink
			Background(lipgloss.Color("#5C1340"))

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			MarginTop(1)

	// Progress styles
	ProgressStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Section styles
	SectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			MarginTop(1).
			MarginBottom(1)

	// Info box
	InfoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1).
			MarginBottom(1)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Warning message
	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	// Spinner
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	// Description
	DescriptionStyle = lipgloss.NewStyle().
				Foreground(dimColor).
				Italic(true)
)

// Icons
const (
	IconCheck     = "✓"
	IconCross     = "✗"
	IconArrow     = "→"
	IconBullet    = "•"
	IconSelected  = "◉"
	IconUnselect  = "○"
	IconCheckbox  = "☑"
	IconUnchecked = "☐"
	IconSpinner   = "⠋"
	IconInfo      = "ℹ"
	IconWarning   = "⚠"
	IconError     = "✖"
	IconSuccess   = "✔"
	IconDocker    = "🐳"
	IconCode      = "📝"
	IconTerminal  = "💻"
	IconFont      = "🔤"
	IconGear      = "⚙"
	IconRocket    = "🚀"
)
