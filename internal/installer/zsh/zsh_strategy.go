// Package zsh provides Zsh and Oh My Zsh installation strategies.
package zsh

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

// Strategy implements Zsh installation.
type Strategy struct {
	strategy.BaseStrategy
}

// NewStrategy creates a new Zsh installer strategy.
func NewStrategy(deps strategy.Dependencies, platform *entity.Platform) *Strategy {
	return &Strategy{
		BaseStrategy: strategy.NewBaseStrategy(deps, platform),
	}
}

// CheckStatus checks if Zsh is already installed.
func (s *Strategy) CheckStatus(ctx context.Context) (valueobject.AppStatus, string, error) {
	if s.CommandExists("zsh") {
		version := s.GetCommandVersion("zsh")
		return valueobject.StatusInstalled, version, nil
	}
	return valueobject.StatusNotInstalled, "", nil
}

// Install installs Zsh.
func (s *Strategy) Install(ctx context.Context, opts primary.InstallOptions) (*result.InstallResult, error) {
	if s.Platform.IsMacOS() {
		return s.installMacOS(ctx, opts), nil
	} else if s.Platform.IsDebian() {
		return s.installUbuntu(ctx, opts), nil
	}
	return result.NewFailure("Systeme non supporte"), nil
}

// installMacOS installs Zsh on macOS (usually pre-installed).
func (s *Strategy) installMacOS(ctx context.Context, opts primary.InstallOptions) *result.InstallResult {
	if s.CommandExists("zsh") {
		s.Success("Zsh est deja installe sur macOS")
		return result.NewSuccess("Zsh deja installe")
	}

	brewPath := s.getHomebrewPath()
	if brewPath == "" {
		return result.NewFailure("Homebrew est requis pour installer Zsh sur macOS")
	}

	s.Section("Installation de Zsh")
	res := s.Run(ctx,
		[]string{brewPath, "install", "zsh"},
		secondary.WithDescription("Installation de Zsh"),
		secondary.WithTimeout(5*time.Minute),
	)

	if !res.Success {
		return result.NewFailure("Echec de l'installation de Zsh", res.Stderr)
	}

	s.Success("Zsh installe")
	return result.NewSuccess("Zsh installe avec succes")
}

// installUbuntu installs Zsh on Ubuntu/Debian.
func (s *Strategy) installUbuntu(ctx context.Context, opts primary.InstallOptions) *result.InstallResult {
	s.Section("Installation de Zsh")

	res := s.Run(ctx,
		[]string{"apt-get", "install", "-y", "zsh"},
		secondary.WithSudo(),
		secondary.WithDescription("Installation de Zsh"),
		secondary.WithTimeout(5*time.Minute),
	)

	if !res.Success {
		return result.NewFailure("Echec de l'installation de Zsh", res.Stderr)
	}

	s.Success("Zsh installe")
	return result.NewSuccess("Zsh installe avec succes")
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

// Verify verifies the Zsh installation.
func (s *Strategy) Verify(ctx context.Context) bool {
	return s.CommandExists("zsh")
}

// Uninstall removes Zsh.
func (s *Strategy) Uninstall(ctx context.Context, opts primary.UninstallOptions) (*result.UninstallResult, error) {
	s.Section("Desinstallation de Zsh")

	if s.Platform.IsMacOS() {
		s.Warning("Zsh est un composant systeme sur macOS, desinstallation non recommandee")
		return result.NewUninstallSuccess("Zsh conserve (composant systeme)"), nil
	}

	if s.Platform.IsDebian() {
		s.Run(ctx, []string{"apt-get", "purge", "-y", "zsh"}, secondary.WithSudo())
	}

	if opts.RemoveZshrc {
		s.Deps.FileSystem.RemoveAll(filepath.Join(s.Platform.HomeDir(), ".zshrc"))
	}

	s.Success("Zsh desinstalle")
	return result.NewUninstallSuccess("Zsh desinstalle avec succes"), nil
}

// NewZshStrategy creates the Zsh installer strategy.
func NewZshStrategy(deps strategy.Dependencies, platform *entity.Platform) (strategy.InstallerStrategy, error) {
	if !platform.IsMacOS() && !platform.IsDebian() {
		return nil, fmt.Errorf("unsupported platform for Zsh: %s", platform.OS())
	}
	return NewStrategy(deps, platform), nil
}
