// Package vscode provides Visual Studio Code installation strategies.
package vscode

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/installer/strategy"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
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

// Strategy implements VSCode installation.
type Strategy struct {
	strategy.BaseStrategy
}

// NewStrategy creates a new VSCode installer strategy.
func NewStrategy(deps strategy.Dependencies, platform *entity.Platform) *Strategy {
	return &Strategy{
		BaseStrategy: strategy.NewBaseStrategy(deps, platform),
	}
}

// CheckStatus checks if VSCode is already installed.
func (s *Strategy) CheckStatus(ctx context.Context) (valueobject.AppStatus, string, error) {
	if s.CommandExists("code") {
		version := s.GetCommandVersion("code")
		return valueobject.StatusInstalled, version, nil
	}

	// Check for macOS app
	if s.Platform.IsMacOS() {
		vscodeAppPaths := []string{
			"/Applications/Visual Studio Code.app",
			filepath.Join(s.Platform.HomeDir(), "Applications/Visual Studio Code.app"),
		}
		for _, path := range vscodeAppPaths {
			if s.Deps.FileSystem.Exists(path) {
				return valueobject.StatusInstalled, "(commande 'code' non configuree)", nil
			}
		}
	}

	return valueobject.StatusNotInstalled, "", nil
}

// Install installs VSCode.
func (s *Strategy) Install(ctx context.Context, opts primary.InstallOptions) (*result.InstallResult, error) {
	var installResult *result.InstallResult

	if s.Platform.IsMacOS() {
		installResult = s.installMacOS(ctx, opts)
	} else if s.Platform.IsDebian() {
		installResult = s.installUbuntu(ctx, opts)
	} else {
		return result.NewFailure("Systeme non supporte"), nil
	}

	if installResult.Success() {
		extensions := defaultExtensions
		if opts.VSCodeOptions != nil && len(opts.VSCodeOptions.Extensions) > 0 {
			extensions = opts.VSCodeOptions.Extensions
		}

		if !opts.NoInteraction {
			if s.Confirm("Installer les extensions recommandees?", true) {
				s.installExtensions(ctx, extensions)
			}
		} else {
			s.installExtensions(ctx, extensions)
		}
	}

	return installResult, nil
}

// installMacOS installs VSCode on macOS using Homebrew.
func (s *Strategy) installMacOS(ctx context.Context, opts primary.InstallOptions) *result.InstallResult {
	brewPath := s.getHomebrewPath()
	if brewPath == "" {
		return result.NewFailure("Homebrew est requis pour installer VSCode sur macOS")
	}

	s.Section("Installation de Visual Studio Code")
	s.Info("Installation via Homebrew...")

	res := s.Run(ctx,
		[]string{brewPath, "install", "--cask", "visual-studio-code"},
		secondary.WithDescription("Installation de VSCode"),
		secondary.WithTimeout(10*time.Minute),
	)

	if !res.Success {
		return result.NewFailure("Echec de l'installation de VSCode", res.Stderr)
	}

	s.Success("Visual Studio Code installe")
	return result.NewSuccess("VSCode installe avec succes")
}

// installUbuntu installs VSCode on Ubuntu/Debian.
func (s *Strategy) installUbuntu(ctx context.Context, opts primary.InstallOptions) *result.InstallResult {
	s.Section("Installation de Visual Studio Code")

	// Add Microsoft GPG key
	s.Info("Ajout de la cle GPG Microsoft...")
	s.Run(ctx,
		[]string{"bash", "-c", "wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > /tmp/packages.microsoft.gpg"},
	)
	s.Run(ctx,
		[]string{"mv", "/tmp/packages.microsoft.gpg", "/etc/apt/trusted.gpg.d/packages.microsoft.gpg"},
		secondary.WithSudo(),
	)

	// Add repository
	s.Info("Configuration du depot APT...")
	repoLine := "deb [arch=amd64,arm64,armhf] https://packages.microsoft.com/repos/code stable main"
	s.Run(ctx,
		[]string{"bash", "-c", "echo '" + repoLine + "' > /etc/apt/sources.list.d/vscode.list"},
		secondary.WithSudo(),
	)

	// Update and install
	s.Run(ctx, []string{"apt-get", "update"}, secondary.WithSudo())
	res := s.Run(ctx,
		[]string{"apt-get", "install", "-y", "code"},
		secondary.WithSudo(),
		secondary.WithDescription("Installation de VSCode"),
		secondary.WithTimeout(10*time.Minute),
	)

	if !res.Success {
		return result.NewFailure("Echec de l'installation de VSCode", res.Stderr)
	}

	s.Success("Visual Studio Code installe")
	return result.NewSuccess("VSCode installe avec succes")
}

// installExtensions installs VSCode extensions.
func (s *Strategy) installExtensions(ctx context.Context, extensions []string) {
	s.Section("Installation des extensions")

	for _, ext := range extensions {
		s.Info("Installation de " + ext + "...")
		res := s.Run(ctx,
			[]string{"code", "--install-extension", ext, "--force"},
			secondary.WithTimeout(2*time.Minute),
		)
		if res.Success {
			s.Success(ext + " installe")
		} else {
			s.Warning("Echec: " + ext)
		}
	}
}

// getHomebrewPath returns the path to Homebrew.
func (s *Strategy) getHomebrewPath() string {
	paths := []string{
		"/opt/homebrew/bin/brew",
		"/usr/local/bin/brew",
	}
	for _, p := range paths {
		if s.Deps.FileSystem.Exists(p) {
			return p
		}
	}
	return s.Deps.Executor.GetCommandPath("brew")
}

// Verify verifies the VSCode installation.
func (s *Strategy) Verify(ctx context.Context) bool {
	return s.CommandExists("code")
}

// Uninstall removes VSCode.
func (s *Strategy) Uninstall(ctx context.Context, opts primary.UninstallOptions) (*result.UninstallResult, error) {
	s.Section("Desinstallation de Visual Studio Code")

	if s.Platform.IsMacOS() {
		return s.uninstallMacOS(ctx, opts), nil
	} else if s.Platform.IsDebian() {
		return s.uninstallUbuntu(ctx, opts), nil
	}

	return result.NewUninstallFailure("Systeme non supporte"), nil
}

// uninstallMacOS uninstalls VSCode on macOS.
func (s *Strategy) uninstallMacOS(ctx context.Context, opts primary.UninstallOptions) *result.UninstallResult {
	brewPath := s.getHomebrewPath()
	if brewPath != "" {
		s.Run(ctx, []string{brewPath, "uninstall", "--cask", "visual-studio-code"})
	}

	// Remove app manually if exists
	s.Deps.FileSystem.RemoveAll("/Applications/Visual Studio Code.app")

	if opts.RemoveData {
		s.Info("Suppression des donnees...")
		s.Deps.FileSystem.RemoveAll(filepath.Join(s.Platform.HomeDir(), "Library/Application Support/Code"))
		s.Deps.FileSystem.RemoveAll(filepath.Join(s.Platform.HomeDir(), ".vscode"))
	}

	s.Success("VSCode desinstalle")
	return result.NewUninstallSuccess("VSCode desinstalle avec succes")
}

// uninstallUbuntu uninstalls VSCode on Ubuntu.
func (s *Strategy) uninstallUbuntu(ctx context.Context, opts primary.UninstallOptions) *result.UninstallResult {
	s.Run(ctx, []string{"apt-get", "purge", "-y", "code"}, secondary.WithSudo())

	if opts.RemoveData {
		s.Info("Suppression des donnees...")
		s.Deps.FileSystem.RemoveAll(filepath.Join(s.Platform.HomeDir(), ".config/Code"))
		s.Deps.FileSystem.RemoveAll(filepath.Join(s.Platform.HomeDir(), ".vscode"))
	}

	s.Success("VSCode desinstalle")
	return result.NewUninstallSuccess("VSCode desinstalle avec succes")
}

// NewVSCodeStrategy creates the VSCode installer strategy for the platform.
func NewVSCodeStrategy(deps strategy.Dependencies, platform *entity.Platform) (strategy.InstallerStrategy, error) {
	if !platform.IsMacOS() && !platform.IsDebian() {
		return nil, fmt.Errorf("unsupported platform for VSCode: %s", platform.OS())
	}
	return NewStrategy(deps, platform), nil
}
