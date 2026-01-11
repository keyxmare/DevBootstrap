// Package strategy provides the installer strategy pattern.
package strategy

import (
	"context"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// InstallerStrategy defines the interface for application installation strategies.
type InstallerStrategy interface {
	// Install installs the application.
	Install(ctx context.Context, opts primary.InstallOptions) (*result.InstallResult, error)

	// Uninstall removes the application.
	Uninstall(ctx context.Context, opts primary.UninstallOptions) (*result.UninstallResult, error)

	// CheckStatus checks if the application is installed.
	CheckStatus(ctx context.Context) (valueobject.AppStatus, string, error)

	// Verify verifies the installation.
	Verify(ctx context.Context) bool
}

// Dependencies contains all dependencies needed by installer strategies.
type Dependencies struct {
	Executor   secondary.CommandExecutor
	FileSystem secondary.FileSystem
	HTTPClient secondary.HTTPClient
	Reporter   secondary.ProgressReporter
	Prompter   secondary.UserPrompter
}

// BaseStrategy provides common functionality for installer strategies.
type BaseStrategy struct {
	Deps     Dependencies
	Platform *entity.Platform
}

// NewBaseStrategy creates a new BaseStrategy instance.
func NewBaseStrategy(deps Dependencies, platform *entity.Platform) BaseStrategy {
	return BaseStrategy{
		Deps:     deps,
		Platform: platform,
	}
}

// CommandExists checks if a command exists in PATH.
func (s *BaseStrategy) CommandExists(name string) bool {
	return s.Deps.Executor.CommandExists(name)
}

// GetCommandVersion returns the version of a command.
func (s *BaseStrategy) GetCommandVersion(name string) string {
	return s.Deps.Executor.GetCommandVersion(name, "--version")
}

// Run executes a command with the given options.
func (s *BaseStrategy) Run(ctx context.Context, args []string, opts ...secondary.CommandOption) *secondary.CommandResult {
	return s.Deps.Executor.Execute(ctx, args, opts...)
}

// Info reports an informational message.
func (s *BaseStrategy) Info(message string) {
	if s.Deps.Reporter != nil {
		s.Deps.Reporter.Info(message)
	}
}

// Success reports a success message.
func (s *BaseStrategy) Success(message string) {
	if s.Deps.Reporter != nil {
		s.Deps.Reporter.Success(message)
	}
}

// Warning reports a warning message.
func (s *BaseStrategy) Warning(message string) {
	if s.Deps.Reporter != nil {
		s.Deps.Reporter.Warning(message)
	}
}

// Error reports an error message.
func (s *BaseStrategy) Error(message string) {
	if s.Deps.Reporter != nil {
		s.Deps.Reporter.Error(message)
	}
}

// Section starts a new section.
func (s *BaseStrategy) Section(title string) {
	if s.Deps.Reporter != nil {
		s.Deps.Reporter.Section(title)
	}
}

// Confirm asks the user for confirmation.
func (s *BaseStrategy) Confirm(question string, defaultValue bool) bool {
	if s.Deps.Prompter != nil {
		return s.Deps.Prompter.Confirm(question, defaultValue)
	}
	return defaultValue
}
