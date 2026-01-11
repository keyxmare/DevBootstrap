// Package tui provides the terminal UI using Bubble Tea.
package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/keyxmare/DevBootstrap/internal/application/usecase"
	"github.com/keyxmare/DevBootstrap/internal/config"
	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
)

const appVersion = "3.0.0"

// App is the TUI application.
type App struct {
	container        *config.Container
	installUseCase   *usecase.InstallApplicationUseCase
	uninstallUseCase *usecase.UninstallApplicationUseCase
	listAppsUseCase  *usecase.ListApplicationsUseCase
}

// NewApp creates a new TUI application.
func NewApp(container *config.Container) *App {
	return &App{
		container:        container,
		installUseCase:   container.InstallUseCase,
		uninstallUseCase: container.UninstallUseCase,
		listAppsUseCase:  container.ListAppsUseCase,
	}
}

// Run starts the TUI application.
func (a *App) Run() int {
	ctx := context.Background()

	// Get all applications with status
	apps, err := a.listAppsUseCase.Execute(ctx)
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return 1
	}

	// Create model
	model := NewModel(a.container.Platform, apps, appVersion)

	// Run Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return 1
	}

	// Handle final state
	m := finalModel.(Model)
	if m.state == StateQuit {
		return 0
	}

	// Execute operations if confirmed
	if m.state == StateResult && m.hasSelection() {
		return a.executeOperations(ctx, &m)
	}

	return 0
}

func (a *App) executeOperations(ctx context.Context, m *Model) int {
	apps := m.getSelectableApps()
	var selectedApps []*entity.Application

	for i, app := range apps {
		if m.selectedApps[i] {
			selectedApps = append(selectedApps, app)
		}
	}

	if len(selectedApps) == 0 {
		return 0
	}

	appIDs := make([]valueobject.AppID, len(selectedApps))
	for i, app := range selectedApps {
		appIDs[i] = app.ID()
	}

	if m.uninstallMode {
		opts := primary.UninstallOptions{
			DryRun:        a.container.DryRun,
			NoInteraction: a.container.NoInteraction,
			RemoveConfig:  true,
			RemoveCache:   true,
			RemoveData:    true,
		}
		_, err := a.uninstallUseCase.ExecuteMultiple(ctx, appIDs, opts)
		if err != nil {
			return 1
		}
	} else {
		opts := primary.InstallOptions{
			DryRun:        a.container.DryRun,
			NoInteraction: a.container.NoInteraction,
		}
		_, err := a.installUseCase.ExecuteMultiple(ctx, appIDs, opts)
		if err != nil {
			return 1
		}
	}

	return 0
}
