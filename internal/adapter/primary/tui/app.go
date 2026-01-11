// Package tui provides the terminal UI using Bubble Tea.
package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/keyxmare/DevBootstrap/internal/config"
)

const appVersion = "3.0.0"

// App is the TUI application.
type App struct {
	container *config.Container
}

// NewApp creates a new TUI application.
func NewApp(container *config.Container) *App {
	return &App{
		container: container,
	}
}

// Run starts the TUI application.
func (a *App) Run() int {
	ctx := context.Background()

	// Get all applications with status
	apps, err := a.container.ListAppsUseCase.Execute(ctx)
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		return 1
	}

	// Create model with use cases injected
	model := NewModel(
		a.container.Platform,
		apps,
		appVersion,
		a.container.InstallUseCase,
		a.container.UninstallUseCase,
		a.container.DryRun,
	)

	// Run Bubble Tea program with alt screen
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

	return 0
}
