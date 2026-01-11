package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// App is the main TUI application
type App struct {
	model   *Model
	runner  *runner.Runner
	sysInfo *system.SystemInfo
	dryRun  bool
}

// NewApp creates a new TUI application
func NewApp(sysInfo *system.SystemInfo, r *runner.Runner, uninstall, noInteract, dryRun bool) *App {
	model := NewModel(sysInfo, uninstall, noInteract, dryRun)

	return &App{
		model:   model,
		runner:  r,
		sysInfo: sysInfo,
		dryRun:  dryRun,
	}
}

// AddApp adds an application to the list
func (a *App) AddApp(id, name, description string, tags []installers.AppTag, status installers.AppStatus, version string, inst installers.Installer, uninst installers.Uninstaller) {
	item := AppItem{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Status:      status,
		Version:     version,
		Selected:    false,
		Installer:   inst,
		Uninstaller: uninst,
	}
	a.model.items = append(a.model.items, item)
}

// Run starts the TUI application
func (a *App) Run() error {
	// Set callbacks
	a.model.SetInstallCallback(a.installCallback)
	a.model.SetUninstallCallback(a.uninstallCallback)

	// Run the program
	p := tea.NewProgram(a.model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (a *App) installCallback(items []AppItem) tea.Cmd {
	return func() tea.Msg {
		opts := &installers.InstallOptions{
			DryRun:        a.dryRun,
			NoInteraction: true,
		}

		for _, item := range items {
			if item.Installer == nil {
				continue
			}

			// Send progress message
			tea.Println(fmt.Sprintf("Installation de %s...", item.Name))

			// Install
			result := item.Installer.Install(opts)
			if !result.Success {
				return InstallCompleteMsg{
					Error: fmt.Errorf("échec de l'installation de %s: %s", item.Name, result.Message),
				}
			}
		}

		return InstallCompleteMsg{Error: nil}
	}
}

func (a *App) uninstallCallback(items []AppItem) tea.Cmd {
	return func() tea.Msg {
		opts := &installers.UninstallOptions{
			DryRun:        a.dryRun,
			NoInteraction: true,
			RemoveConfig:  true,
			RemoveCache:   true,
			RemoveData:    true,
		}

		for _, item := range items {
			if item.Uninstaller == nil {
				continue
			}

			// Send progress message
			tea.Println(fmt.Sprintf("Désinstallation de %s...", item.Name))

			// Uninstall
			result := item.Uninstaller.Uninstall(opts)
			if !result.Success {
				return InstallCompleteMsg{
					Error: fmt.Errorf("échec de la désinstallation de %s: %s", item.Name, result.Message),
				}
			}
		}

		return InstallCompleteMsg{Error: nil}
	}
}
