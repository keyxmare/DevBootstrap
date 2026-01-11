package usecase

import (
	"context"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// ListApplicationsUseCase handles listing all applications.
type ListApplicationsUseCase struct {
	repository secondary.AppRepository
	factory    InstallerFactory
	reporter   secondary.ProgressReporter
}

// NewListApplicationsUseCase creates a new ListApplicationsUseCase.
func NewListApplicationsUseCase(
	repo secondary.AppRepository,
	factory InstallerFactory,
	reporter secondary.ProgressReporter,
) *ListApplicationsUseCase {
	return &ListApplicationsUseCase{
		repository: repo,
		factory:    factory,
		reporter:   reporter,
	}
}

// Execute returns all available applications with their current status.
func (uc *ListApplicationsUseCase) Execute(ctx context.Context) ([]*entity.Application, error) {
	apps, err := uc.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Update status for each application
	for _, app := range apps {
		installer, err := uc.factory.GetInstaller(app.ID())
		if err != nil {
			continue
		}

		status, version, _ := installer.CheckStatus(ctx)
		if status.IsInstalled() {
			app.MarkInstalled(valueobject.NewVersion(version))
		} else {
			app.MarkNotInstalled()
		}
	}

	return apps, nil
}

// ExecuteInstalled returns only installed applications.
func (uc *ListApplicationsUseCase) ExecuteInstalled(ctx context.Context) ([]*entity.Application, error) {
	apps, err := uc.Execute(ctx)
	if err != nil {
		return nil, err
	}

	installed := make([]*entity.Application, 0)
	for _, app := range apps {
		if app.IsInstalled() {
			installed = append(installed, app)
		}
	}

	return installed, nil
}

// ExecuteNotInstalled returns only not-installed applications.
func (uc *ListApplicationsUseCase) ExecuteNotInstalled(ctx context.Context) ([]*entity.Application, error) {
	apps, err := uc.Execute(ctx)
	if err != nil {
		return nil, err
	}

	notInstalled := make([]*entity.Application, 0)
	for _, app := range apps {
		if !app.IsInstalled() {
			notInstalled = append(notInstalled, app)
		}
	}

	return notInstalled, nil
}
