package usecase

import (
	"context"
	"fmt"

	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// UninstallApplicationUseCase handles application uninstallation.
type UninstallApplicationUseCase struct {
	repository secondary.AppRepository
	factory    InstallerFactory
	reporter   secondary.ProgressReporter
	prompter   secondary.UserPrompter
}

// NewUninstallApplicationUseCase creates a new UninstallApplicationUseCase.
func NewUninstallApplicationUseCase(
	repo secondary.AppRepository,
	factory InstallerFactory,
	reporter secondary.ProgressReporter,
	prompter secondary.UserPrompter,
) *UninstallApplicationUseCase {
	return &UninstallApplicationUseCase{
		repository: repo,
		factory:    factory,
		reporter:   reporter,
		prompter:   prompter,
	}
}

// Execute uninstalls an application by ID.
func (uc *UninstallApplicationUseCase) Execute(
	ctx context.Context,
	appID valueobject.AppID,
	opts primary.UninstallOptions,
) (*result.UninstallResult, error) {
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

	// Check if installed
	status, _, err := installer.CheckStatus(ctx)
	if err != nil {
		uc.reporter.Warning(fmt.Sprintf("Could not check status: %v", err))
	}

	if !status.IsInstalled() {
		uc.reporter.Warning(fmt.Sprintf("%s n'est pas installe", app.Name()))
		return result.NewUninstallSuccess("Application non installee"), nil
	}

	// Confirm uninstallation
	if !opts.NoInteraction {
		if !uc.prompter.Confirm(fmt.Sprintf("Etes-vous sur de vouloir desinstaller %s?", app.Name()), false) {
			return result.NewUninstallSuccess("Desinstallation annulee"), nil
		}
	}

	// Run uninstallation
	uninstallResult, err := installer.Uninstall(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Update repository status
	if uninstallResult.Success() {
		uc.repository.UpdateStatus(ctx, appID, valueobject.StatusNotInstalled, valueobject.Version{})
	}

	return uninstallResult, nil
}

// ExecuteMultiple uninstalls multiple applications.
func (uc *UninstallApplicationUseCase) ExecuteMultiple(
	ctx context.Context,
	appIDs []valueobject.AppID,
	opts primary.UninstallOptions,
) (map[string]*result.UninstallResult, error) {
	results := make(map[string]*result.UninstallResult)

	for i, appID := range appIDs {
		uc.reporter.Step(i+1, len(appIDs), fmt.Sprintf("Desinstallation de %s", appID))

		uninstallResult, err := uc.Execute(ctx, appID, opts)
		if err != nil {
			results[appID.String()] = result.NewUninstallFailure(err.Error())
		} else {
			results[appID.String()] = uninstallResult
		}
	}

	return results, nil
}
