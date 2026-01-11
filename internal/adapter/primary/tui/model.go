// Package tui provides the terminal UI using Bubble Tea.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/keyxmare/DevBootstrap/internal/application/usecase"
	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
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
	iconSpinner      = "◐"
)

// Messages for async operations
type installDoneMsg struct {
	successCount int
	failureCount int
}
type tickMsg time.Time

// Model is the main TUI model.
type Model struct {
	// Dimensions
	width  int
	height int
	styles Styles

	// State
	state State
	err   error

	// Data
	platform *entity.Platform
	apps     []*entity.Application
	version  string

	// Mode selection
	uninstallMode bool
	modeIndex     int

	// App selection
	appIndex     int
	selectedApps map[int]bool

	// Installation progress
	installing   bool
	spinnerFrame int
	successCount int
	failureCount int

	// Use cases (injected)
	installUseCase   *usecase.InstallApplicationUseCase
	uninstallUseCase *usecase.UninstallApplicationUseCase

	// Context for operations
	ctx context.Context
}

// ContainerInterface defines what we need from the container.
type ContainerInterface interface {
	GetDryRun() bool
}

// NewModel creates a new TUI model.
func NewModel(
	platform *entity.Platform,
	apps []*entity.Application,
	version string,
	installUC *usecase.InstallApplicationUseCase,
	uninstallUC *usecase.UninstallApplicationUseCase,
	dryRun bool,
) Model {
	return Model{
		platform:         platform,
		apps:             apps,
		version:          version,
		state:            StateMain,
		modeIndex:        0,
		appIndex:         0,
		selectedApps:     make(map[int]bool),
		installUseCase:   installUC,
		uninstallUseCase: uninstallUC,
		ctx:              context.Background(),
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

	case tickMsg:
		if m.state == StateInstalling {
			m.spinnerFrame++
			return m, tickCmd()
		}
		return m, nil

	case installDoneMsg:
		m.installing = false
		m.successCount = msg.successCount
		m.failureCount = msg.failureCount
		m.state = StateResult
		return m, nil

	case tea.KeyMsg:
		// Don't handle keys during installation
		if m.state == StateInstalling {
			return m, nil
		}

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

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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
		// Start installation
		m.state = StateInstalling
		m.installing = true
		m.successCount = 0
		m.failureCount = 0
		return m, tea.Batch(tickCmd(), m.startInstallation())
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

func (m Model) getSelectedApps() []*entity.Application {
	apps := m.getSelectableApps()
	var selected []*entity.Application
	for i, app := range apps {
		if m.selectedApps[i] {
			selected = append(selected, app)
		}
	}
	return selected
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

func (m Model) startInstallation() tea.Cmd {
	// Capture values needed for the async operation
	selectedApps := m.getSelectedApps()
	uninstallMode := m.uninstallMode
	installUC := m.installUseCase
	uninstallUC := m.uninstallUseCase
	ctx := m.ctx

	return func() tea.Msg {
		if len(selectedApps) == 0 {
			return installDoneMsg{0, 0}
		}

		successCount := 0
		failureCount := 0

		for _, app := range selectedApps {
			appIDs := []valueobject.AppID{app.ID()}

			var success bool
			if uninstallMode {
				opts := primary.UninstallOptions{
					DryRun:        false,
					NoInteraction: true,
					RemoveConfig:  true,
					RemoveCache:   true,
					RemoveData:    true,
				}
				results, err := uninstallUC.ExecuteMultiple(ctx, appIDs, opts)
				if err == nil {
					// Check individual result
					if r, ok := results[app.ID().String()]; ok && r.Success() {
						success = true
					}
				}
			} else {
				opts := primary.InstallOptions{
					DryRun:        false,
					NoInteraction: true,
				}
				results, err := installUC.ExecuteMultiple(ctx, appIDs, opts)
				if err == nil {
					// Check individual result
					if r, ok := results[app.ID().String()]; ok && r.Success() {
						success = true
					}
				}
			}

			if success {
				successCount++
			} else {
				failureCount++
			}
		}

		return installDoneMsg{successCount, failureCount}
	}
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
	case StateInstalling:
		content = m.viewInstalling()
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

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")
	b.WriteString(m.renderSystemInfo())
	b.WriteString("\n\n")
	b.WriteString(m.renderAppList())
	b.WriteString("\n\n")
	b.WriteString(m.renderFooter("Entrée: continuer • q: quitter"))

	return m.styles.Content.Render(b.String())
}

func (m Model) viewModeSelect() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	b.WriteString(m.styles.SectionTitle.Render("Action"))
	b.WriteString("\n")
	b.WriteString(m.styles.SectionLine.Render(strings.Repeat("─", 40)))
	b.WriteString("\n\n")

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

	title := "Sélection"
	b.WriteString(m.styles.SectionTitle.Render(title))
	b.WriteString("\n")
	b.WriteString(m.styles.SectionLine.Render(strings.Repeat("─", 40)))
	b.WriteString("\n\n")

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

	count := 0
	for _, selected := range m.selectedApps {
		if selected {
			count++
		}
	}
	b.WriteString("\n")
	b.WriteString(m.styles.Muted.Render(fmt.Sprintf("%d sélectionnée(s)", count)))
	b.WriteString("\n\n")

	b.WriteString(m.renderFooter("↑/↓: naviguer • Espace: sélectionner • Entrée: continuer • Esc: retour"))

	return m.styles.Content.Render(b.String())
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	title := "Confirmer"
	if m.uninstallMode {
		title = "Confirmer la suppression"
	}
	b.WriteString(m.styles.SectionTitle.Render(title))
	b.WriteString("\n")
	b.WriteString(m.styles.SectionLine.Render(strings.Repeat("─", 40)))
	b.WriteString("\n\n")

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

func (m Model) viewInstalling() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	action := "Installation"
	if m.uninstallMode {
		action = "Désinstallation"
	}
	b.WriteString(m.styles.SectionTitle.Render(action + " en cours..."))
	b.WriteString("\n")
	b.WriteString(m.styles.SectionLine.Render(strings.Repeat("─", 40)))
	b.WriteString("\n\n")

	// Spinner animation
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]

	b.WriteString(fmt.Sprintf("  %s  ", m.styles.Info.Render(spinner)))
	b.WriteString(m.styles.Muted.Render("Veuillez patienter..."))
	b.WriteString("\n\n")

	// Show selected apps
	apps := m.getSelectedApps()
	for _, app := range apps {
		b.WriteString(fmt.Sprintf("  %s  %s\n", m.styles.Muted.Render("◌"), app.Name()))
	}

	b.WriteString("\n\n")
	b.WriteString(m.renderFooter("Opération en cours..."))

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

	// Summary
	total := m.successCount + m.failureCount
	if m.failureCount == 0 {
		b.WriteString(m.styles.Success.Render(fmt.Sprintf("  %s  %d/%d opération(s) réussie(s)", iconCheck, m.successCount, total)))
	} else {
		b.WriteString(m.styles.Warning.Render(fmt.Sprintf("  !  %d succès, %d échec(s)", m.successCount, m.failureCount)))
	}

	b.WriteString("\n\n")
	b.WriteString(m.renderFooter("Entrée: quitter"))

	return m.styles.Content.Render(b.String())
}

func (m Model) renderHeader() string {
	var b strings.Builder

	divider := m.styles.Divider.Render(strings.Repeat("─", 60))
	b.WriteString(divider)
	b.WriteString("\n\n")

	title := m.styles.Title.Render(fmt.Sprintf("DevBootstrap v%s", m.version))
	b.WriteString(lipgloss.PlaceHorizontal(60, lipgloss.Center, title))
	b.WriteString("\n")

	subtitle := m.styles.Subtitle.Render("Configuration de votre environnement de développement")
	b.WriteString(lipgloss.PlaceHorizontal(60, lipgloss.Center, subtitle))
	b.WriteString("\n\n")

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
