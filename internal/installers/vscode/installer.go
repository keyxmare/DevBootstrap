// Package vscode provides Visual Studio Code installation functionality.
package vscode

import (
	"os"
	"path/filepath"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/cli"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// Default extensions to install
var defaultExtensions = []string{
	"ms-python.python",
	"esbenp.prettier-vscode",
	"dbaeumer.vscode-eslint",
	"ms-vscode.vscode-typescript-next",
	"bradlc.vscode-tailwindcss",
	"eamodio.gitlens",
	"PKief.material-icon-theme",
}

// Installer handles VSCode installation.
type Installer struct {
	*installers.BaseInstaller
}

// NewInstaller creates a new VSCode installer.
func NewInstaller(c *cli.CLI, r *runner.Runner, sysInfo *system.SystemInfo) *Installer {
	return &Installer{
		BaseInstaller: installers.NewBaseInstaller(c, r, sysInfo),
	}
}

// Name returns the application name.
func (i *Installer) Name() string {
	return "Visual Studio Code"
}

// ID returns the application ID.
func (i *Installer) ID() string {
	return "vscode"
}

// Description returns the application description.
func (i *Installer) Description() string {
	return "Editeur de code source leger et puissant"
}

// Tags returns the application tags.
func (i *Installer) Tags() []installers.AppTag {
	return []installers.AppTag{installers.TagApp, installers.TagEditor}
}

// CheckExisting checks if VSCode is already installed.
func (i *Installer) CheckExisting() (installers.AppStatus, string) {
	// Check if code command exists
	if i.Runner.CommandExists("code") {
		version := i.Runner.GetCommandVersion("code", "--version")
		return installers.StatusInstalled, version
	}

	// Check for macOS app
	if i.SystemInfo.IsMacOS() {
		vscodeAppPaths := []string{
			"/Applications/Visual Studio Code.app",
			filepath.Join(i.SystemInfo.HomeDir, "Applications/Visual Studio Code.app"),
		}
		for _, path := range vscodeAppPaths {
			if _, err := os.Stat(path); err == nil {
				return installers.StatusInstalled, "(commande 'code' non configuree)"
			}
		}
	}

	return installers.StatusNotInstalled, ""
}

// Install installs VSCode.
func (i *Installer) Install(opts *installers.InstallOptions) *installers.InstallResult {
	// Check if already installed
	status, version := i.CheckExisting()
	if status == installers.StatusInstalled {
		i.CLI.PrintInfo("Visual Studio Code est deja installe: " + version)
		if !i.CLI.AskYesNo("Voulez-vous reinstaller?", false) {
			return installers.NewSuccessResult("VSCode deja installe")
		}
	}

	var result *installers.InstallResult

	if i.SystemInfo.IsMacOS() {
		result = i.installMacOS(opts)
	} else if i.SystemInfo.IsDebian() {
		result = i.installUbuntu(opts)
	} else {
		return installers.NewFailureResult("Systeme non supporte")
	}

	if result.Success && len(opts.Extensions) > 0 {
		i.installExtensions(opts.Extensions)
	} else if result.Success {
		// Ask about default extensions
		if i.CLI.AskYesNo("Installer les extensions recommandees?", true) {
			i.installExtensions(defaultExtensions)
		}
	}

	return result
}

// installMacOS installs VSCode on macOS using Homebrew.
func (i *Installer) installMacOS(opts *installers.InstallOptions) *installers.InstallResult {
	brewPath := i.getHomebrewPath()
	if brewPath == "" {
		return installers.NewFailureResult("Homebrew est requis pour installer VSCode sur macOS")
	}

	i.CLI.PrintSection("Installation de Visual Studio Code")
	i.CLI.PrintInfo("Installation via Homebrew...")

	result := i.Runner.Run(
		[]string{brewPath, "install", "--cask", "visual-studio-code"},
		runner.WithDescription("Installation de VSCode"),
		runner.WithTimeout(10*time.Minute),
	)

	if !result.Success {
		return installers.NewFailureResult("Echec de l'installation de VSCode", result.Stderr)
	}

	i.CLI.PrintSuccess("Visual Studio Code installe")
	return installers.NewSuccessResult("VSCode installe avec succes")
}

// installUbuntu installs VSCode on Ubuntu/Debian.
func (i *Installer) installUbuntu(opts *installers.InstallOptions) *installers.InstallResult {
	i.CLI.PrintSection("Installation de Visual Studio Code")

	// Add Microsoft GPG key
	i.CLI.PrintInfo("Ajout de la cle GPG Microsoft...")
	i.Runner.Run(
		[]string{"bash", "-c", "wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > /tmp/packages.microsoft.gpg"},
	)
	i.Runner.Run(
		[]string{"mv", "/tmp/packages.microsoft.gpg", "/etc/apt/trusted.gpg.d/packages.microsoft.gpg"},
		runner.WithSudo(),
	)

	// Add repository
	i.CLI.PrintInfo("Configuration du depot APT...")
	repoLine := "deb [arch=amd64,arm64,armhf] https://packages.microsoft.com/repos/code stable main"
	i.Runner.Run(
		[]string{"bash", "-c", "echo '" + repoLine + "' > /etc/apt/sources.list.d/vscode.list"},
		runner.WithSudo(),
	)

	// Update and install
	i.Runner.Run([]string{"apt-get", "update"}, runner.WithSudo())
	result := i.Runner.Run(
		[]string{"apt-get", "install", "-y", "code"},
		runner.WithSudo(),
		runner.WithDescription("Installation de VSCode"),
		runner.WithTimeout(10*time.Minute),
	)

	if !result.Success {
		return installers.NewFailureResult("Echec de l'installation de VSCode", result.Stderr)
	}

	i.CLI.PrintSuccess("Visual Studio Code installe")
	return installers.NewSuccessResult("VSCode installe avec succes")
}

// installExtensions installs VSCode extensions.
func (i *Installer) installExtensions(extensions []string) {
	i.CLI.PrintSection("Installation des extensions")

	for _, ext := range extensions {
		i.CLI.PrintInfo("Installation de " + ext + "...")
		result := i.Runner.Run(
			[]string{"code", "--install-extension", ext, "--force"},
			runner.WithTimeout(2*time.Minute),
		)
		if result.Success {
			i.CLI.PrintSuccess(ext + " installe")
		} else {
			i.CLI.PrintWarning("Echec: " + ext)
		}
	}
}

// getHomebrewPath returns the path to Homebrew.
func (i *Installer) getHomebrewPath() string {
	paths := []string{
		"/opt/homebrew/bin/brew",
		"/usr/local/bin/brew",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return i.Runner.GetCommandPath("brew")
}

// Verify verifies the VSCode installation.
func (i *Installer) Verify() bool {
	return i.Runner.CommandExists("code")
}

// Uninstall removes VSCode.
func (i *Installer) Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult {
	i.CLI.PrintSection("Desinstallation de Visual Studio Code")

	if i.SystemInfo.IsMacOS() {
		return i.uninstallMacOS(opts)
	} else if i.SystemInfo.IsDebian() {
		return i.uninstallUbuntu(opts)
	}

	return installers.NewUninstallFailureResult("Systeme non supporte")
}

// uninstallMacOS uninstalls VSCode on macOS.
func (i *Installer) uninstallMacOS(opts *installers.UninstallOptions) *installers.UninstallResult {
	brewPath := i.getHomebrewPath()
	if brewPath != "" {
		i.Runner.Run([]string{brewPath, "uninstall", "--cask", "visual-studio-code"})
	}

	// Remove app manually if exists
	i.Runner.RemoveAll("/Applications/Visual Studio Code.app")

	if opts.RemoveData {
		i.CLI.PrintInfo("Suppression des donnees...")
		i.Runner.RemoveAll(filepath.Join(i.SystemInfo.HomeDir, "Library/Application Support/Code"))
		i.Runner.RemoveAll(filepath.Join(i.SystemInfo.HomeDir, ".vscode"))
	}

	i.CLI.PrintSuccess("VSCode desinstalle")
	return installers.NewUninstallSuccessResult("VSCode desinstalle avec succes")
}

// uninstallUbuntu uninstalls VSCode on Ubuntu.
func (i *Installer) uninstallUbuntu(opts *installers.UninstallOptions) *installers.UninstallResult {
	i.Runner.Run([]string{"apt-get", "purge", "-y", "code"}, runner.WithSudo())

	if opts.RemoveData {
		i.CLI.PrintInfo("Suppression des donnees...")
		i.Runner.RemoveAll(filepath.Join(i.SystemInfo.HomeDir, ".config/Code"))
		i.Runner.RemoveAll(filepath.Join(i.SystemInfo.HomeDir, ".vscode"))
	}

	i.CLI.PrintSuccess("VSCode desinstalle")
	return installers.NewUninstallSuccessResult("VSCode desinstalle avec succes")
}
