package config

import (
	"context"

	"github.com/keyxmare/DevBootstrap/internal/adapter/secondary/detector"
	"github.com/keyxmare/DevBootstrap/internal/adapter/secondary/executor"
	"github.com/keyxmare/DevBootstrap/internal/adapter/secondary/filesystem"
	httpAdapter "github.com/keyxmare/DevBootstrap/internal/adapter/secondary/http"
	"github.com/keyxmare/DevBootstrap/internal/adapter/secondary/prompter"
	"github.com/keyxmare/DevBootstrap/internal/adapter/secondary/reporter"
	"github.com/keyxmare/DevBootstrap/internal/adapter/secondary/repository"
	"github.com/keyxmare/DevBootstrap/internal/application/usecase"
	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/installer/strategy"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// Container holds all application dependencies.
type Container struct {
	// Configuration
	DryRun        bool
	NoInteraction bool
	UninstallMode bool

	// Secondary adapters (infrastructure)
	CommandExecutor  secondary.CommandExecutor
	FileSystem       secondary.FileSystem
	HTTPClient       secondary.HTTPClient
	UserPrompter     secondary.UserPrompter
	ProgressReporter secondary.ProgressReporter
	SystemDetector   secondary.SystemDetector
	AppRepository    secondary.AppRepository

	// Platform
	Platform *entity.Platform

	// Use cases
	InstallUseCase   *usecase.InstallApplicationUseCase
	UninstallUseCase *usecase.UninstallApplicationUseCase
	ListAppsUseCase  *usecase.ListApplicationsUseCase

	// Installer factory
	InstallerFactory *InstallerFactory
}

// NewContainer creates and wires all dependencies.
func NewContainer(dryRun, noInteraction, uninstallMode bool) *Container {
	c := &Container{
		DryRun:        dryRun,
		NoInteraction: noInteraction,
		UninstallMode: uninstallMode,
	}

	// Create progress reporter first (used by others)
	c.ProgressReporter = reporter.NewTerminalReporter(dryRun)

	// Create secondary adapters
	shellExecutor := executor.NewShellExecutor(dryRun, c.ProgressReporter)
	c.CommandExecutor = shellExecutor
	c.FileSystem = filesystem.NewLocalFileSystem(dryRun, c.ProgressReporter)
	c.HTTPClient = httpAdapter.NewDownloader(c.ProgressReporter)
	c.UserPrompter = prompter.NewSurveyPrompter(noInteraction, c.ProgressReporter)
	c.SystemDetector = detector.NewSystemDetector()

	// Detect platform
	c.Platform = c.SystemDetector.Detect()

	// Create repository
	c.AppRepository = repository.NewMemoryRepository()

	// Initialize repository with known apps
	c.initializeApps()

	// Create installer dependencies
	installerDeps := strategy.Dependencies{
		Executor:   c.CommandExecutor,
		FileSystem: c.FileSystem,
		HTTPClient: c.HTTPClient,
		Reporter:   c.ProgressReporter,
		Prompter:   c.UserPrompter,
	}

	// Create installer factory
	c.InstallerFactory = NewInstallerFactory(installerDeps, c.Platform)

	// Create use cases
	c.InstallUseCase = usecase.NewInstallApplicationUseCase(
		c.AppRepository,
		c.InstallerFactory,
		c.ProgressReporter,
		c.UserPrompter,
	)

	c.UninstallUseCase = usecase.NewUninstallApplicationUseCase(
		c.AppRepository,
		c.InstallerFactory,
		c.ProgressReporter,
		c.UserPrompter,
	)

	c.ListAppsUseCase = usecase.NewListApplicationsUseCase(
		c.AppRepository,
		c.InstallerFactory,
		c.ProgressReporter,
	)

	return c
}

// initializeApps registers all known applications in the repository.
func (c *Container) initializeApps() {
	ctx := context.Background()

	apps := []*entity.Application{
		entity.NewApplication("docker", "Docker", "Plateforme de conteneurisation",
			[]valueobject.AppTag{valueobject.TagApp, valueobject.TagContainer}),
		entity.NewApplication("vscode", "Visual Studio Code", "Editeur de code source leger et puissant",
			[]valueobject.AppTag{valueobject.TagApp, valueobject.TagEditor}),
		entity.NewApplication("neovim", "Neovim", "Editeur de texte moderne",
			[]valueobject.AppTag{valueobject.TagApp, valueobject.TagEditor}),
		entity.NewApplication("neovim-config", "Neovim Config", "Configuration et plugins pour Neovim",
			[]valueobject.AppTag{valueobject.TagConfig}),
		entity.NewApplication("zsh", "Zsh", "Shell Z moderne",
			[]valueobject.AppTag{valueobject.TagApp, valueobject.TagShell}),
		entity.NewApplication("oh-my-zsh", "Oh My Zsh", "Framework de configuration pour Zsh avec plugins",
			[]valueobject.AppTag{valueobject.TagConfig}),
		entity.NewApplication("nerd-font", "Nerd Font", "Polices avec icones pour terminal",
			[]valueobject.AppTag{valueobject.TagFont}),
	}

	for _, app := range apps {
		c.AppRepository.Save(ctx, app)
	}
}

// SetSudoAskpass sets the SUDO_ASKPASS script path on the command executor.
func (c *Container) SetSudoAskpass(path string) {
	if exec, ok := c.CommandExecutor.(*executor.ShellExecutor); ok {
		exec.SetSudoAskpass(path)
	}
}
