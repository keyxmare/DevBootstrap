// Package tui provides the terminal UI using Bubble Tea.
package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
)

// State represents the current UI state.
type State int

const (
	StateLoading State = iota
	StateMain
	StateModeSelect
	StateAppSelect
	StateConfirm
	StateInstalling
	StateResult
	StateQuit
)

// Icons
const (
	iconInstalled    = "●"
	iconNotInstalled = "○"
	iconSelected     = "◉"
	iconUnselected   = "○"
	iconArrow        = "→"
	iconCheck        = "✓"
	iconCross        = "✕"
)

// Model is the main TUI model.
type Model struct {
	// Dimensions
	width  int
	height int
	styles Styles

	// State
	state       State
	err         error

	// Data
	platform    *entity.Platform
	apps        []*entity.Application
	version     string

	// Mode selection
	uninstallMode bool
	modeIndex     int

	// App selection
	appIndex     int
	selectedApps map[int]bool

	// Installation
	installing    bool
	installIndex  int
	results       map[string]bool

	// Context for operations
	ctx context.Context
}

// NewModel creates a new TUI model.
func NewModel(platform *entity.Platform, apps []*entity.Application, version string) Model {
	return Model{
		platform:     platform,
		apps:         apps,
		version:      version,
		state:        StateMain,
		modeIndex:    0,
		appIndex:     0,
		selectedApps: make(map[int]bool),
		results:      make(map[string]bool),
		ctx:          context.Background(),
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.styles = NewStyles(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.state = StateQuit
			return m, tea.Quit

		case "enter":
			return m.handleEnter()

		case "up", "k":
			return m.handleUp()

		case "down", "j":
			return m.handleDown()

		case " ":
			return m.handleSpace()

		case "esc":
			return m.handleEsc()
		}
	}

	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMain:
		m.state = StateModeSelect
	case StateModeSelect:
		m.uninstallMode = m.modeIndex == 1
		m.initAppSelection()
		m.state = StateAppSelect
	case StateAppSelect:
		if m.hasSelection() {
			m.state = StateConfirm
		}
	case StateConfirm:
		// Would trigger installation here
		m.state = StateResult
	case StateResult:
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleUp() (tea.Model, tea.Cmd) {
	switch m.state {
	case StateModeSelect:
		if m.modeIndex > 0 {
			m.modeIndex--
		}
	case StateAppSelect:
		if m.appIndex > 0 {
			m.appIndex--
		}
	}
	return m, nil
}

func (m Model) handleDown() (tea.Model, tea.Cmd) {
	switch m.state {
	case StateModeSelect:
		if m.modeIndex < 1 {
			m.modeIndex++
		}
	case StateAppSelect:
		if m.appIndex < len(m.getSelectableApps())-1 {
			m.appIndex++
		}
	}
	return m, nil
}

func (m Model) handleSpace() (tea.Model, tea.Cmd) {
	if m.state == StateAppSelect {
		m.selectedApps[m.appIndex] = !m.selectedApps[m.appIndex]
	}
	return m, nil
}

func (m Model) handleEsc() (tea.Model, tea.Cmd) {
	switch m.state {
	case StateModeSelect:
		m.state = StateMain
	case StateAppSelect:
		m.state = StateModeSelect
	case StateConfirm:
		m.state = StateAppSelect
	}
	return m, nil
}

func (m *Model) initAppSelection() {
	m.selectedApps = make(map[int]bool)
	apps := m.getSelectableApps()
	for i, app := range apps {
		if m.uninstallMode {
			m.selectedApps[i] = false
		} else {
			m.selectedApps[i] = !app.IsInstalled()
		}
	}
}

func (m Model) getSelectableApps() []*entity.Application {
	if m.uninstallMode {
		var installed []*entity.Application
		for _, app := range m.apps {
			if app.IsInstalled() {
				installed = append(installed, app)
			}
		}
		return installed
	}
	return m.apps
}

func (m Model) hasSelection() bool {
	for _, selected := range m.selectedApps {
		if selected {
			return true
		}
	}
	return false
}

func (m Model) countInstalled() int {
	count := 0
	for _, app := range m.apps {
		if app.IsInstalled() {
			count++
		}
	}
	return count
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string

	switch m.state {
	case StateMain:
		content = m.viewMain()
	case StateModeSelect:
		content = m.viewModeSelect()
	case StateAppSelect:
		content = m.viewAppSelect()
	case StateConfirm:
		content = m.viewConfirm()
	case StateResult:
		content = m.viewResult()
	default:
		content = m.viewMain()
	}

	// Center everything
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m Model) viewMain() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// System info
	b.WriteString(m.renderSystemInfo())
	b.WriteString("\n\n")

	// Applications list
	b.WriteString(m.renderAppList())
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderFooter("Appuyez sur Entrée pour continuer"))

	return m.styles.Content.Render(b.String())
}

func (m Model) viewModeSelect() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Section title
	b.WriteString(m.styles.SectionTitle.Render("Action"))
	b.WriteString("\n")
	b.WriteString(m.styles.SectionLine.Render(strings.Repeat("─", 40)))
	b.WriteString("\n\n")

	// Options
	options := []struct {
		title string
		desc  string
	}{
		{"Installer", "Ajouter des applications"},
		{"Désinstaller", "Supprimer des applications"},
	}

	for i, opt := range options {
		cursor := "  "
		style := m.styles.UnselectedItem
		if i == m.modeIndex {
			cursor = m.styles.Cursor.Render("▸ ")
			style = m.styles.SelectedItem
		}

		line := fmt.Sprintf("%s%s", cursor, style.Render(opt.title))
		b.WriteString(line)
		b.WriteString("   ")
		b.WriteString(m.styles.Muted.Render(opt.desc))
		b.WriteString("\n\n")
	}

	b.WriteString("\n")
	b.WriteString(m.renderFooter("↑/↓: naviguer • Entrée: confirmer • Esc: retour"))

	return m.styles.Content.Render(b.String())
}

func (m Model) viewAppSelect() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Section title
	title := "Installation"
	if m.uninstallMode {
		title = "Désinstallation"
	}
	b.WriteString(m.styles.SectionTitle.Render(title))
	b.WriteString("\n")
	b.WriteString(m.styles.SectionLine.Render(strings.Repeat("─", 40)))
	b.WriteString("\n\n")

	// App list
	apps := m.getSelectableApps()
	for i, app := range apps {
		cursor := "  "
		if i == m.appIndex {
			cursor = m.styles.Cursor.Render("▸ ")
		}

		checkbox := iconUnselected
		if m.selectedApps[i] {
			checkbox = m.styles.Success.Render(iconSelected)
		} else {
			checkbox = m.styles.Muted.Render(iconUnselected)
		}

		name := app.Name()
		if m.selectedApps[i] {
			name = m.styles.AppName.Render(name)
		} else {
			name = m.styles.AppNameDim.Render(name)
		}

		line := fmt.Sprintf("%s%s  %s", cursor, checkbox, name)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Selection count
	count := 0
	for _, selected := range m.selectedApps {
		if selected {
			count++
		}
	}
	b.WriteString("\n")
	b.WriteString(m.styles.Muted.Render(fmt.Sprintf("%d sélectionnée(s)", count)))
	b.WriteString("\n\n")

	b.WriteString(m.renderFooter("↑/↓: naviguer • Espace: sélectionner • Entrée: confirmer • Esc: retour"))

	return m.styles.Content.Render(b.String())
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Section title
	title := "Confirmer l'installation"
	if m.uninstallMode {
		title = "Confirmer la désinstallation"
	}
	b.WriteString(m.styles.SectionTitle.Render(title))
	b.WriteString("\n")
	b.WriteString(m.styles.SectionLine.Render(strings.Repeat("─", 40)))
	b.WriteString("\n\n")

	// Selected apps
	apps := m.getSelectableApps()
	for i, app := range apps {
		if m.selectedApps[i] {
			icon := m.styles.Success.Render(iconArrow)
			if m.uninstallMode {
				icon = m.styles.Error.Render(iconCross)
			}
			b.WriteString(fmt.Sprintf("  %s  %s\n", icon, m.styles.AppName.Render(app.Name())))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(m.renderFooter("Entrée: confirmer • Esc: retour"))

	return m.styles.Content.Render(b.String())
}

func (m Model) viewResult() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	b.WriteString(m.styles.SectionTitle.Render("Terminé"))
	b.WriteString("\n")
	b.WriteString(m.styles.SectionLine.Render(strings.Repeat("─", 40)))
	b.WriteString("\n\n")

	b.WriteString(m.styles.Success.Render("✓  Opération terminée avec succès"))
	b.WriteString("\n\n")

	b.WriteString(m.renderFooter("Entrée: quitter"))

	return m.styles.Content.Render(b.String())
}

func (m Model) renderHeader() string {
	var b strings.Builder

	// Divider
	divider := m.styles.Divider.Render(strings.Repeat("─", 60))
	b.WriteString(divider)
	b.WriteString("\n\n")

	// Title
	title := m.styles.Title.Render(fmt.Sprintf("DevBootstrap v%s", m.version))
	b.WriteString(lipgloss.PlaceHorizontal(60, lipgloss.Center, title))
	b.WriteString("\n")

	// Subtitle
	subtitle := m.styles.Subtitle.Render("Configuration de votre environnement de développement")
	b.WriteString(lipgloss.PlaceHorizontal(60, lipgloss.Center, subtitle))
	b.WriteString("\n\n")

	// Divider
	b.WriteString(divider)

	return b.String()
}

func (m Model) renderSystemInfo() string {
	var b strings.Builder

	b.WriteString(m.styles.SectionTitle.Render("Système"))
	b.WriteString("\n")
	b.WriteString(m.styles.SectionLine.Render(strings.Repeat("─", 40)))
	b.WriteString("\n\n")

	info := []struct {
		label string
		value string
	}{
		{"OS", m.platform.OSName()},
		{"Version", m.platform.OSVersion()},
		{"Architecture", m.platform.Arch().String()},
	}

	for _, item := range info {
		label := m.styles.Label.Render(item.label)
		value := m.styles.Value.Render(item.value)
		b.WriteString(fmt.Sprintf("  %s  %s\n", label, value))
	}

	return b.String()
}

func (m Model) renderAppList() string {
	var b strings.Builder

	b.WriteString(m.styles.SectionTitle.Render("Applications"))
	b.WriteString("\n")
	b.WriteString(m.styles.SectionLine.Render(strings.Repeat("─", 40)))
	b.WriteString("\n\n")

	for _, app := range m.apps {
		var icon, name string

		if app.IsInstalled() {
			icon = m.styles.AppIcon.Render(iconInstalled)
			name = m.styles.AppName.Render(app.Name())

			if !app.Version().IsEmpty() {
				ver := cleanVersion(app.Version().String())
				if ver != "" {
					name += "  " + m.styles.AppVersion.Render(ver)
				}
			}
		} else {
			icon = m.styles.AppIconDim.Render(iconNotInstalled)
			name = m.styles.AppNameDim.Render(app.Name())
		}

		b.WriteString(fmt.Sprintf("  %s %s\n", icon, name))
		b.WriteString(m.styles.AppDescription.Render(app.Description()))
		b.WriteString("\n\n")
	}

	// Counter
	counter := m.styles.Counter.Render(fmt.Sprintf("%d/%d installées", m.countInstalled(), len(m.apps)))
	b.WriteString(counter)

	return b.String()
}

func (m Model) renderFooter(help string) string {
	return m.styles.Help.Render(help)
}

// cleanVersion cleans and shortens version strings.
func cleanVersion(version string) string {
	version = strings.TrimPrefix(version, "Docker version ")
	version = strings.TrimPrefix(version, "NVIM ")
	version = strings.TrimPrefix(version, "zsh ")

	if strings.Contains(version, "commande 'code' non") {
		return ""
	}

	if strings.Contains(version, "Nerd Font") {
		parts := strings.Split(version, " ")
		if len(parts) > 0 {
			return parts[0]
		}
	}

	if idx := strings.Index(version, ","); idx > 0 {
		version = version[:idx]
	}
	if idx := strings.Index(version, "+g"); idx > 0 {
		version = version[:idx]
	}
	if idx := strings.Index(version, "-dev"); idx > 0 {
		version = version[:idx]
	}
	if idx := strings.Index(version, " ("); idx > 0 {
		version = version[:idx]
	}

	if len(version) > 20 {
		return version[:17] + "..."
	}

	return version
}
