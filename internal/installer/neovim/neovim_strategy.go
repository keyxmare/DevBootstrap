// Package neovim provides Neovim installation strategies.
package neovim

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

// Strategy implements Neovim installation.
type Strategy struct {
	strategy.BaseStrategy
}

// NewStrategy creates a new Neovim installer strategy.
func NewStrategy(deps strategy.Dependencies, platform *entity.Platform) *Strategy {
	return &Strategy{
		BaseStrategy: strategy.NewBaseStrategy(deps, platform),
	}
}

// CheckStatus checks if Neovim is already installed.
func (s *Strategy) CheckStatus(ctx context.Context) (valueobject.AppStatus, string, error) {
	if s.CommandExists("nvim") {
		version := s.GetCommandVersion("nvim")
		return valueobject.StatusInstalled, version, nil
	}
	return valueobject.StatusNotInstalled, "", nil
}

// Install installs Neovim.
func (s *Strategy) Install(ctx context.Context, opts primary.InstallOptions) (*result.InstallResult, error) {
	if s.Platform.IsMacOS() {
		return s.installMacOS(ctx, opts), nil
	} else if s.Platform.IsDebian() {
		return s.installUbuntu(ctx, opts), nil
	}
	return result.NewFailure("Systeme non supporte"), nil
}

// installMacOS installs Neovim on macOS using Homebrew.
func (s *Strategy) installMacOS(ctx context.Context, opts primary.InstallOptions) *result.InstallResult {
	brewPath := s.getHomebrewPath()
	if brewPath == "" {
		return result.NewFailure("Homebrew est requis pour installer Neovim sur macOS")
	}

	s.Section("Installation de Neovim")
	s.Info("Installation via Homebrew...")

	res := s.Run(ctx,
		[]string{brewPath, "install", "neovim"},
		secondary.WithDescription("Installation de Neovim"),
		secondary.WithTimeout(10*time.Minute),
	)

	if !res.Success {
		return result.NewFailure("Echec de l'installation de Neovim", res.Stderr)
	}

	s.installDependencies(ctx)
	s.Success("Neovim installe")
	return result.NewSuccess("Neovim installe avec succes")
}

// installUbuntu installs Neovim on Ubuntu/Debian.
func (s *Strategy) installUbuntu(ctx context.Context, opts primary.InstallOptions) *result.InstallResult {
	s.Section("Installation de Neovim")

	s.Info("Ajout du PPA Neovim...")
	s.Run(ctx, []string{"add-apt-repository", "-y", "ppa:neovim-ppa/unstable"}, secondary.WithSudo())
	s.Run(ctx, []string{"apt-get", "update"}, secondary.WithSudo())

	res := s.Run(ctx,
		[]string{"apt-get", "install", "-y", "neovim"},
		secondary.WithSudo(),
		secondary.WithDescription("Installation de Neovim"),
		secondary.WithTimeout(10*time.Minute),
	)

	if !res.Success {
		return result.NewFailure("Echec de l'installation de Neovim", res.Stderr)
	}

	s.installDependencies(ctx)
	s.Success("Neovim installe")
	return result.NewSuccess("Neovim installe avec succes")
}

// installDependencies installs common dependencies for Neovim.
func (s *Strategy) installDependencies(ctx context.Context) {
	s.Section("Installation des dependances")

	deps := []string{"ripgrep", "fd"}
	if s.Platform.IsMacOS() {
		deps = append(deps, "fzf", "lazygit")
	}

	for _, dep := range deps {
		s.Info("Installation de " + dep + "...")
		if s.Platform.IsMacOS() {
			brewPath := s.getHomebrewPath()
			s.Run(ctx, []string{brewPath, "install", dep})
		} else {
			pkgName := dep
			if dep == "fd" {
				pkgName = "fd-find"
			}
			s.Run(ctx, []string{"apt-get", "install", "-y", pkgName}, secondary.WithSudo())
		}
	}
}

// getHomebrewPath returns the path to Homebrew.
func (s *Strategy) getHomebrewPath() string {
	paths := []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"}
	for _, p := range paths {
		if s.Deps.FileSystem.Exists(p) {
			return p
		}
	}
	return s.Deps.Executor.GetCommandPath("brew")
}

// Verify verifies the Neovim installation.
func (s *Strategy) Verify(ctx context.Context) bool {
	return s.CommandExists("nvim")
}

// Uninstall removes Neovim.
func (s *Strategy) Uninstall(ctx context.Context, opts primary.UninstallOptions) (*result.UninstallResult, error) {
	s.Section("Desinstallation de Neovim")

	if s.Platform.IsMacOS() {
		brewPath := s.getHomebrewPath()
		if brewPath != "" {
			s.Run(ctx, []string{brewPath, "uninstall", "neovim"})
		}
	} else if s.Platform.IsDebian() {
		s.Run(ctx, []string{"apt-get", "purge", "-y", "neovim"}, secondary.WithSudo())
	}

	if opts.RemoveConfig {
		s.Info("Suppression de la configuration...")
		s.Deps.FileSystem.RemoveAll(filepath.Join(s.Platform.HomeDir(), ".config/nvim"))
	}

	if opts.RemoveData {
		s.Info("Suppression des donnees...")
		s.Deps.FileSystem.RemoveAll(filepath.Join(s.Platform.HomeDir(), ".local/share/nvim"))
		s.Deps.FileSystem.RemoveAll(filepath.Join(s.Platform.HomeDir(), ".local/state/nvim"))
	}

	if opts.RemoveCache {
		s.Info("Suppression du cache...")
		s.Deps.FileSystem.RemoveAll(filepath.Join(s.Platform.HomeDir(), ".cache/nvim"))
	}

	s.Success("Neovim desinstalle")
	return result.NewUninstallSuccess("Neovim desinstalle avec succes"), nil
}

// NewNeovimStrategy creates the Neovim installer strategy.
func NewNeovimStrategy(deps strategy.Dependencies, platform *entity.Platform) (strategy.InstallerStrategy, error) {
	if !platform.IsMacOS() && !platform.IsDebian() {
		return nil, fmt.Errorf("unsupported platform for Neovim: %s", platform.OS())
	}
	return NewStrategy(deps, platform), nil
}
