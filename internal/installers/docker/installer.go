// Package docker provides Docker installation functionality.
package docker

import (
	"github.com/keyxmare/DevBootstrap/internal/cli"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// Installer handles Docker installation.
type Installer struct {
	*installers.BaseInstaller
	platform platformInstaller
}

// platformInstaller defines platform-specific installation methods.
type platformInstaller interface {
	Install(opts *installers.InstallOptions) *installers.InstallResult
	Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult
	CheckExisting() (installers.AppStatus, string)
	Verify() bool
}

// NewInstaller creates a new Docker installer.
func NewInstaller(c *cli.CLI, r *runner.Runner, sysInfo *system.SystemInfo) *Installer {
	base := installers.NewBaseInstaller(c, r, sysInfo)
	inst := &Installer{
		BaseInstaller: base,
	}

	// Select platform-specific installer
	if sysInfo.IsMacOS() {
		inst.platform = newMacOSInstaller(base)
	} else if sysInfo.IsDebian() {
		inst.platform = newUbuntuInstaller(base)
	}

	return inst
}

// Name returns the application name.
func (i *Installer) Name() string {
	return "Docker"
}

// ID returns the application ID.
func (i *Installer) ID() string {
	return "docker"
}

// Description returns the application description.
func (i *Installer) Description() string {
	return "Plateforme de conteneurisation"
}

// Tags returns the application tags.
func (i *Installer) Tags() []installers.AppTag {
	return []installers.AppTag{installers.TagApp, installers.TagContainer}
}

// CheckExisting checks if Docker is already installed.
func (i *Installer) CheckExisting() (installers.AppStatus, string) {
	if i.platform == nil {
		return installers.StatusNotInstalled, ""
	}
	return i.platform.CheckExisting()
}

// Install installs Docker.
func (i *Installer) Install(opts *installers.InstallOptions) *installers.InstallResult {
	if i.platform == nil {
		return installers.NewFailureResult("Systeme non supporte")
	}

	// Check if already installed
	status, version := i.CheckExisting()
	if status == installers.StatusInstalled {
		i.CLI.PrintInfo("Docker est deja installe: " + version)
		if !i.CLI.AskYesNo("Voulez-vous reinstaller?", false) {
			return installers.NewSuccessResult("Docker deja installe")
		}
	}

	// Run platform-specific installation
	return i.platform.Install(opts)
}

// Verify verifies the Docker installation.
func (i *Installer) Verify() bool {
	if i.platform == nil {
		return false
	}
	return i.platform.Verify()
}

// Uninstall removes Docker.
func (i *Installer) Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult {
	if i.platform == nil {
		return installers.NewUninstallFailureResult("Systeme non supporte")
	}
	return i.platform.Uninstall(opts)
}
