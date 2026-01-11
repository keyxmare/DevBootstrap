// Package usecase contains application use cases.
package usecase

import (
	"context"
	"fmt"

	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/installer/strategy"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// InstallerFactory creates installer strategies for applications.
type InstallerFactory interface {
	GetInstaller(appID valueobject.AppID) (strategy.InstallerStrategy, error)
}

// InstallApplicationUseCase handles application installation.
type InstallApplicationUseCase struct {
	repository secondary.AppRepository
	factory    InstallerFactory
	reporter   secondary.ProgressReporter
	prompter   secondary.UserPrompter
}

// NewInstallApplicationUseCase creates a new InstallApplicationUseCase.
func NewInstallApplicationUseCase(
	repo secondary.AppRepository,
	factory InstallerFactory,
	reporter secondary.ProgressReporter,
	prompter secondary.UserPrompter,
) *InstallApplicationUseCase {
	return &InstallApplicationUseCase{
		repository: repo,
		factory:    factory,
		reporter:   reporter,
		prompter:   prompter,
	}
}

// Execute installs an application by ID.
func (uc *InstallApplicationUseCase) Execute(
	ctx context.Context,
	appID valueobject.AppID,
	opts primary.InstallOptions,
) (*result.InstallResult, error) {
	// Get application from repository
	app, err := uc.repository.GetByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("application not found: %s", appID)
	}

	// Get installer strategy
	installer, err := uc.factory.GetInstaller(appID)
	if err != nil {
		return nil, fmt.Errorf("no installer for %s: %w", appID, err)
	}

	// Check if already installed
	status, version, err := installer.CheckStatus(ctx)
	if err != nil {
		uc.reporter.Warning(fmt.Sprintf("Could not check status: %v", err))
	}

	if status.IsInstalled() && !opts.NoInteraction {
		uc.reporter.Info(fmt.Sprintf("%s est deja installe: %s", app.Name(), version))
		if !uc.prompter.Confirm("Voulez-vous reinstaller?", false) {
			return result.NewSuccess("Deja installe").WithVersion(version), nil
		}
	}

	// Run installation
	installResult, err := installer.Install(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Update repository status
	if installResult.Success() {
		uc.repository.UpdateStatus(ctx, appID, valueobject.StatusInstalled, valueobject.NewVersion(installResult.Version()))
	}

	return installResult, nil
}

// ExecuteMultiple installs multiple applications.
func (uc *InstallApplicationUseCase) ExecuteMultiple(
	ctx context.Context,
	appIDs []valueobject.AppID,
	opts primary.InstallOptions,
) (map[string]*result.InstallResult, error) {
	results := make(map[string]*result.InstallResult)

	for i, appID := range appIDs {
		uc.reporter.Step(i+1, len(appIDs), fmt.Sprintf("Installation de %s", appID))

		installResult, err := uc.Execute(ctx, appID, opts)
		if err != nil {
			results[appID.String()] = result.NewFailure(err.Error())
		} else {
			results[appID.String()] = installResult
		}
	}

	return results, nil
}
