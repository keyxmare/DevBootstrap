package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// AppItem represents an application in the list
type AppItem struct {
	ID          string
	Name        string
	Description string
	Tags        []installers.AppTag
	Status      installers.AppStatus
	Version     string
	Selected    bool
	Installer   installers.Installer
	Uninstaller installers.Uninstaller
}

// Model is the main TUI model
type Model struct {
	// State
	items      []AppItem
	cursor     int
	width      int
	height     int
	quitting   bool
	uninstall  bool
	noInteract bool
	dryRun     bool

	// System info
	sysInfo *system.SystemInfo

	// Callback when user confirms selection
	onConfirm func(items []AppItem)
}

// KeyMap defines keybindings
type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Toggle  key.Binding
	All     key.Binding
	None    key.Binding
	Confirm key.Binding
	Quit    key.Binding
}

var keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "monter"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "descendre"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("espace", "sélectionner"),
	),
	All: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "tout"),
	),
	None: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "rien"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter", "c"),
		key.WithHelp("entrée/c", "confirmer"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quitter"),
	),
}

// NewModel creates a new TUI model
func NewModel(sysInfo *system.SystemInfo, uninstall, noInteract, dryRun bool) *Model {
	return &Model{
		items:      make([]AppItem, 0),
		cursor:     0,
		sysInfo:    sysInfo,
		uninstall:  uninstall,
		noInteract: noInteract,
		dryRun:     dryRun,
	}
}

// SetConfirmCallback sets the callback when user confirms selection
func (m *Model) SetConfirmCallback(cb func(items []AppItem)) {
	m.onConfirm = cb
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}

	case key.Matches(msg, keys.Toggle):
		if len(m.items) > 0 {
			m.items[m.cursor].Selected = !m.items[m.cursor].Selected
		}

	case key.Matches(msg, keys.All):
		for i := range m.items {
			m.items[i].Selected = true
		}

	case key.Matches(msg, keys.None):
		for i := range m.items {
			m.items[i].Selected = false
		}

	case key.Matches(msg, keys.Confirm):
		selected := m.getSelectedItems()
		if len(selected) > 0 && m.onConfirm != nil {
			m.onConfirm(selected)
		}
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) getSelectedItems() []AppItem {
	var selected []AppItem
	for _, item := range m.items {
		if item.Selected {
			selected = append(selected, item)
		}
	}
	return selected
}

// View renders the UI
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Header
	title := "DevBootstrap"
	if m.uninstall {
		title += " - Désinstallation"
	}
	b.WriteString(m.renderHeader(title))
	b.WriteString("\n\n")

	// System info
	b.WriteString(m.renderSystemInfo())
	b.WriteString("\n\n")

	// App list
	action := "installer"
	if m.uninstall {
		action = "désinstaller"
	}
	b.WriteString(SectionStyle.Render(fmt.Sprintf("Applications à %s", action)))
	b.WriteString("\n\n")

	for i, item := range m.items {
		b.WriteString(m.renderItem(item, i == m.cursor))
		b.WriteString("\n")
	}

	// Selected count
	selectedCount := len(m.getSelectedItems())
	if selectedCount > 0 {
		b.WriteString("\n")
		countStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)
		b.WriteString(countStyle.Render(fmt.Sprintf("%d application(s) sélectionnée(s)", selectedCount)))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m *Model) renderHeader(title string) string {
	width := 60
	if m.width > 0 && m.width < width {
		width = m.width - 4
	}

	titleText := lipgloss.NewStyle().
		Bold(true).
		Foreground(textColor).
		Render(title)

	version := lipgloss.NewStyle().
		Foreground(dimColor).
		Render("v2.0.0")

	header := lipgloss.JoinHorizontal(lipgloss.Center, titleText, "  ", version)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(0, 2).
		Width(width).
		Align(lipgloss.Center).
		Render(header)

	return box
}

func (m *Model) renderSystemInfo() string {
	osName := m.sysInfo.OSName
	if osName == "" {
		osName = "Unknown"
	}

	arch := m.sysInfo.Arch.String()

	info := fmt.Sprintf("%s %s • %s", IconTerminal, osName, arch)
	return InfoBoxStyle.Render(info)
}

func (m *Model) renderItem(item AppItem, selected bool) string {
	// Checkbox
	checkbox := UncheckedStyle.Render(IconUnchecked)
	if item.Selected {
		checkbox = CheckedStyle.Render(IconCheckbox)
	}

	// Cursor
	cursor := "  "
	if selected {
		cursor = SelectedItemStyle.Render(IconArrow + " ")
	}

	// Name
	name := item.Name
	if selected {
		name = SelectedItemStyle.Render(name)
	}

	// Status
	status := ""
	if item.Status == installers.StatusInstalled {
		status = InstalledStyle.Render(" [installé]")
		if item.Version != "" {
			status = InstalledStyle.Render(fmt.Sprintf(" [%s]", item.Version))
		}
	}

	// Tags
	tags := m.renderTags(item.Tags)

	// Description
	desc := DescriptionStyle.Render(item.Description)

	// Build line
	line1 := fmt.Sprintf("%s%s %s%s %s", cursor, checkbox, name, status, tags)
	line2 := fmt.Sprintf("      %s", desc)

	return line1 + "\n" + line2
}

func (m *Model) renderTags(tags []installers.AppTag) string {
	var parts []string
	for _, tag := range tags {
		var style lipgloss.Style
		switch tag {
		case installers.TagApp:
			style = TagAppStyle
		case installers.TagConfig:
			style = TagConfigStyle
		case installers.TagEditor:
			style = TagEditorStyle
		case installers.TagShell:
			style = TagShellStyle
		case installers.TagContainer:
			style = TagContainerStyle
		case installers.TagFont:
			style = TagFontStyle
		default:
			style = TagStyle.Foreground(dimColor)
		}
		parts = append(parts, style.Render(string(tag)))
	}
	return strings.Join(parts, " ")
}

func (m *Model) renderHelp() string {
	help := []string{
		"↑/↓ naviguer",
		"espace sélectionner",
		"a tout",
		"n rien",
		"entrée confirmer",
		"q quitter",
	}

	return HelpStyle.Render(strings.Join(help, " • "))
}
