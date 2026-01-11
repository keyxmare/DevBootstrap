// Package installers provides the installation framework for DevBootstrap.
package installers

import (
	"github.com/keyxmare/DevBootstrap/internal/cli"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// AppTag represents a category tag for applications.
type AppTag string

const (
	TagApp       AppTag = "app"
	TagConfig    AppTag = "config"
	TagEditor    AppTag = "editeur"
	TagShell     AppTag = "shell"
	TagContainer AppTag = "container"
	TagFont      AppTag = "police"
)

// AppStatus represents the installation status of an application.
type AppStatus int

const (
	StatusNotInstalled AppStatus = iota
	StatusInstalled
	StatusUpdateAvailable
)

// String returns the string representation of AppStatus.
func (s AppStatus) String() string {
	switch s {
	case StatusInstalled:
		return "installe"
	case StatusUpdateAvailable:
		return "mise a jour disponible"
	default:
		return "non installe"
	}
}

// InstallOptions contains options for installation.
type InstallOptions struct {
	DryRun        bool
	NoInteraction bool

	// Docker-specific
	InstallCompose       bool
	AddUserToDockerGroup bool
	StartOnBoot          bool

	// Neovim-specific
	ConfigPreset string // "minimal", "full", "custom"
	CustomConfig string

	// Zsh-specific
	InstallOhMyZsh bool
	ZshTheme       string
	ZshPlugins     []string

	// VSCode-specific
	Extensions []string

	// Font-specific
	FontFamily string
}

// InstallResult contains the result of an installation.
type InstallResult struct {
	Success  bool
	Message  string
	Version  string
	Path     string
	Errors   []string
	Warnings []string
}

// NewSuccessResult creates a successful result.
func NewSuccessResult(message string) *InstallResult {
	return &InstallResult{
		Success: true,
		Message: message,
	}
}

// NewFailureResult creates a failed result.
func NewFailureResult(message string, errors ...string) *InstallResult {
	return &InstallResult{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}

// AddWarning adds a warning to the result.
func (r *InstallResult) AddWarning(warning string) {
	r.Warnings = append(r.Warnings, warning)
}

// AddError adds an error to the result.
func (r *InstallResult) AddError(err string) {
	r.Errors = append(r.Errors, err)
}

// UninstallOptions contains options for uninstallation.
type UninstallOptions struct {
	DryRun        bool
	NoInteraction bool

	// Common options
	RemoveConfig bool
	RemoveCache  bool
	RemoveData   bool

	// Docker-specific
	RemoveImages  bool
	RemoveVolumes bool

	// Zsh-specific
	RemoveOhMyZsh bool
	RemovePlugins bool
	RemoveZshrc   bool
}

// UninstallResult contains the result of an uninstallation.
type UninstallResult struct {
	Success  bool
	Message  string
	Errors   []string
	Warnings []string
}

// NewUninstallSuccessResult creates a successful uninstall result.
func NewUninstallSuccessResult(message string) *UninstallResult {
	return &UninstallResult{
		Success: true,
		Message: message,
	}
}

// NewUninstallFailureResult creates a failed uninstall result.
func NewUninstallFailureResult(message string, errors ...string) *UninstallResult {
	return &UninstallResult{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}

// Installer is the interface for application installers.
type Installer interface {
	// Name returns the human-readable name of the application.
	Name() string

	// ID returns the unique identifier of the application.
	ID() string

	// Description returns a brief description of the application.
	Description() string

	// Tags returns the tags associated with the application.
	Tags() []AppTag

	// CheckExisting checks if the application is already installed.
	CheckExisting() (AppStatus, string)

	// Install installs the application with the given options.
	Install(opts *InstallOptions) *InstallResult

	// Verify verifies that the installation was successful.
	Verify() bool
}

// Uninstaller is the interface for application uninstallers.
type Uninstaller interface {
	// Uninstall removes the application with the given options.
	Uninstall(opts *UninstallOptions) *UninstallResult
}

// BaseInstaller provides common functionality for installers.
type BaseInstaller struct {
	CLI        *cli.CLI
	Runner     *runner.Runner
	SystemInfo *system.SystemInfo
}

// NewBaseInstaller creates a new BaseInstaller.
func NewBaseInstaller(c *cli.CLI, r *runner.Runner, sysInfo *system.SystemInfo) *BaseInstaller {
	return &BaseInstaller{
		CLI:        c,
		Runner:     r,
		SystemInfo: sysInfo,
	}
}
