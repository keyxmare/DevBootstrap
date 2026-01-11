package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// View represents the current view/screen
type View int

const (
	ViewMain View = iota
	ViewInstalling
	ViewUninstalling
	ViewComplete
	ViewError
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
	view        View
	items       []AppItem
	cursor      int
	width       int
	height      int
	quitting    bool
	uninstall   bool
	noInteract  bool
	dryRun      bool

	// System info
	sysInfo *system.SystemInfo

	// Installation state
	installing     bool
	currentApp     string
	installLog     []string
	installError   string
	installSuccess bool

	// Components
	spinner spinner.Model

	// Callbacks
	onInstall   func(items []AppItem) tea.Cmd
	onUninstall func(items []AppItem) tea.Cmd
}

// KeyMap defines keybindings
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	All      key.Binding
	None     key.Binding
	Confirm  key.Binding
	Quit     key.Binding
	Help     key.Binding
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
		key.WithKeys(" ", "enter"),
		key.WithHelp("espace", "sélectionner"),
	),
	All: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "tout sélectionner"),
	),
	None: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "tout désélectionner"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("c", "enter"),
		key.WithHelp("c", "confirmer"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quitter"),
	),
}

// NewModel creates a new TUI model
func NewModel(sysInfo *system.SystemInfo, uninstall, noInteract, dryRun bool) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return &Model{
		view:       ViewMain,
		items:      make([]AppItem, 0),
		cursor:     0,
		sysInfo:    sysInfo,
		uninstall:  uninstall,
		noInteract: noInteract,
		dryRun:     dryRun,
		spinner:    s,
		installLog: make([]string, 0),
	}
}

// SetItems sets the application items
func (m *Model) SetItems(items []AppItem) {
	m.items = items
}

// SetInstallCallback sets the installation callback
func (m *Model) SetInstallCallback(cb func(items []AppItem) tea.Cmd) {
	m.onInstall = cb
}

// SetUninstallCallback sets the uninstallation callback
func (m *Model) SetUninstallCallback(cb func(items []AppItem) tea.Cmd) {
	m.onUninstall = cb
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
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

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case InstallProgressMsg:
		m.currentApp = msg.AppName
		m.installLog = append(m.installLog, msg.Message)
		return m, nil

	case InstallCompleteMsg:
		m.installing = false
		if msg.Error != nil {
			m.view = ViewError
			m.installError = msg.Error.Error()
		} else {
			m.view = ViewComplete
			m.installSuccess = true
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.view == ViewInstalling {
		// During installation, only allow quit
		if key.Matches(msg, keys.Quit) {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	}

	if m.view == ViewComplete || m.view == ViewError {
		// Any key to quit from complete/error view
		m.quitting = true
		return m, tea.Quit
	}

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
		return m.startInstallation()
	}

	return m, nil
}

func (m *Model) startInstallation() (tea.Model, tea.Cmd) {
	// Get selected items
	selected := m.getSelectedItems()
	if len(selected) == 0 {
		return m, nil
	}

	m.view = ViewInstalling
	m.installing = true
	m.installLog = []string{}

	if m.uninstall {
		if m.onUninstall != nil {
			return m, m.onUninstall(selected)
		}
	} else {
		if m.onInstall != nil {
			return m, m.onInstall(selected)
		}
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

	switch m.view {
	case ViewMain:
		return m.viewMain()
	case ViewInstalling:
		return m.viewInstalling()
	case ViewComplete:
		return m.viewComplete()
	case ViewError:
		return m.viewError()
	default:
		return m.viewMain()
	}
}

func (m *Model) viewMain() string {
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

	// Help
	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m *Model) viewInstalling() string {
	var b strings.Builder

	action := "Installation"
	if m.uninstall {
		action = "Désinstallation"
	}

	b.WriteString(m.renderHeader(action + " en cours..."))
	b.WriteString("\n\n")

	// Spinner and current app
	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	if m.currentApp != "" {
		b.WriteString(ProgressStyle.Render(m.currentApp))
	}
	b.WriteString("\n\n")

	// Log
	maxLogs := 10
	start := 0
	if len(m.installLog) > maxLogs {
		start = len(m.installLog) - maxLogs
	}
	for _, log := range m.installLog[start:] {
		b.WriteString(DescriptionStyle.Render("  " + log))
		b.WriteString("\n")
	}

	return b.String()
}

func (m *Model) viewComplete() string {
	var b strings.Builder

	b.WriteString(m.renderHeader("Terminé !"))
	b.WriteString("\n\n")

	action := "installées"
	if m.uninstall {
		action = "désinstallées"
	}

	b.WriteString(SuccessStyle.Render(fmt.Sprintf("%s Applications %s avec succès !", IconSuccess, action)))
	b.WriteString("\n\n")

	// Show what was installed
	for _, item := range m.items {
		if item.Selected {
			b.WriteString(fmt.Sprintf("  %s %s\n", CheckedStyle.Render(IconCheck), item.Name))
		}
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Appuyez sur une touche pour quitter..."))

	return b.String()
}

func (m *Model) viewError() string {
	var b strings.Builder

	b.WriteString(m.renderHeader("Erreur"))
	b.WriteString("\n\n")

	b.WriteString(ErrorStyle.Render(fmt.Sprintf("%s Une erreur s'est produite:", IconError)))
	b.WriteString("\n\n")
	b.WriteString(DescriptionStyle.Render(m.installError))
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("Appuyez sur une touche pour quitter..."))

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
		keys.Up.Help().Key + "/" + keys.Down.Help().Key + " naviguer",
		"espace sélectionner",
		"a tout",
		"n rien",
		"c confirmer",
		"q quitter",
	}

	return HelpStyle.Render(strings.Join(help, " • "))
}

// Messages
type InstallProgressMsg struct {
	AppName string
	Message string
}

type InstallCompleteMsg struct {
	Error error
}
