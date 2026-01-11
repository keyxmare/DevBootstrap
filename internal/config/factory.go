// Package config provides dependency injection and configuration.
package config

import (
	"fmt"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/installer/docker"
	"github.com/keyxmare/DevBootstrap/internal/installer/font"
	"github.com/keyxmare/DevBootstrap/internal/installer/neovim"
	"github.com/keyxmare/DevBootstrap/internal/installer/strategy"
	"github.com/keyxmare/DevBootstrap/internal/installer/vscode"
	"github.com/keyxmare/DevBootstrap/internal/installer/zsh"
)

// InstallerFactory creates installer strategies based on app ID.
type InstallerFactory struct {
	deps     strategy.Dependencies
	platform *entity.Platform
}

// NewInstallerFactory creates a new InstallerFactory.
func NewInstallerFactory(deps strategy.Dependencies, platform *entity.Platform) *InstallerFactory {
	return &InstallerFactory{
		deps:     deps,
		platform: platform,
	}
}

// GetInstaller returns the installer strategy for the given app ID.
func (f *InstallerFactory) GetInstaller(appID valueobject.AppID) (strategy.InstallerStrategy, error) {
	switch appID.String() {
	case "docker":
		return docker.NewStrategy(f.deps, f.platform)
	case "vscode":
		return vscode.NewVSCodeStrategy(f.deps, f.platform)
	case "neovim":
		return neovim.NewNeovimStrategy(f.deps, f.platform)
	case "neovim-config":
		return neovim.NewNeovimConfigStrategy(f.deps, f.platform)
	case "zsh":
		return zsh.NewZshStrategy(f.deps, f.platform)
	case "oh-my-zsh":
		return zsh.NewOhMyZshInstallerStrategy(f.deps, f.platform)
	case "nerd-font":
		return font.NewFontStrategy(f.deps, f.platform)
	default:
		return nil, fmt.Errorf("unknown application: %s", appID)
	}
}
