// Package app provides the main application logic for DevBootstrap.
package app

import (
	"github.com/keyxmare/DevBootstrap/internal/cli"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/installers/docker"
	"github.com/keyxmare/DevBootstrap/internal/installers/font"
	"github.com/keyxmare/DevBootstrap/internal/installers/neovim"
	"github.com/keyxmare/DevBootstrap/internal/installers/vscode"
	"github.com/keyxmare/DevBootstrap/internal/installers/zsh"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// AppEntry represents an application in the registry.
type AppEntry struct {
	ID          string
	Name        string
	Description string
	Tags        []installers.AppTag
	Installer   installers.Installer
	Uninstaller installers.Uninstaller
}

// Registry holds all available applications.
type Registry struct {
	Apps []*AppEntry
}

// NewRegistry creates a new application registry with all available installers.
func NewRegistry(c *cli.CLI, r *runner.Runner, sysInfo *system.SystemInfo) *Registry {
	reg := &Registry{
		Apps: make([]*AppEntry, 0),
	}

	// Docker
	dockerInstaller := docker.NewInstaller(c, r, sysInfo)
	reg.Apps = append(reg.Apps, &AppEntry{
		ID:          "docker",
		Name:        "Docker",
		Description: "Plateforme de conteneurisation",
		Tags:        []installers.AppTag{installers.TagApp, installers.TagContainer},
		Installer:   dockerInstaller,
		Uninstaller: dockerInstaller,
	})

	// VSCode
	vscodeInstaller := vscode.NewInstaller(c, r, sysInfo)
	reg.Apps = append(reg.Apps, &AppEntry{
		ID:          "vscode",
		Name:        "Visual Studio Code",
		Description: "Editeur de code source leger et puissant",
		Tags:        []installers.AppTag{installers.TagApp, installers.TagEditor},
		Installer:   vscodeInstaller,
		Uninstaller: vscodeInstaller,
	})

	// Neovim
	neovimInstaller := neovim.NewInstaller(c, r, sysInfo)
	reg.Apps = append(reg.Apps, &AppEntry{
		ID:          "neovim",
		Name:        "Neovim",
		Description: "Editeur de texte moderne",
		Tags:        []installers.AppTag{installers.TagApp, installers.TagEditor},
		Installer:   neovimInstaller,
		Uninstaller: neovimInstaller,
	})

	// Neovim Config
	neovimConfigInstaller := neovim.NewConfigInstaller(c, r, sysInfo)
	reg.Apps = append(reg.Apps, &AppEntry{
		ID:          "neovim-config",
		Name:        "Neovim Config",
		Description: "Configuration et plugins pour Neovim",
		Tags:        []installers.AppTag{installers.TagConfig},
		Installer:   neovimConfigInstaller,
		Uninstaller: neovimConfigInstaller,
	})

	// Zsh
	zshInstaller := zsh.NewInstaller(c, r, sysInfo)
	reg.Apps = append(reg.Apps, &AppEntry{
		ID:          "zsh",
		Name:        "Zsh",
		Description: "Shell Z moderne",
		Tags:        []installers.AppTag{installers.TagApp, installers.TagShell},
		Installer:   zshInstaller,
		Uninstaller: zshInstaller,
	})

	// Oh My Zsh
	ohmyzshInstaller := zsh.NewOhMyZshInstaller(c, r, sysInfo)
	reg.Apps = append(reg.Apps, &AppEntry{
		ID:          "oh-my-zsh",
		Name:        "Oh My Zsh",
		Description: "Framework de configuration pour Zsh avec plugins",
		Tags:        []installers.AppTag{installers.TagConfig},
		Installer:   ohmyzshInstaller,
		Uninstaller: ohmyzshInstaller,
	})

	// Nerd Font
	fontInstaller := font.NewInstaller(c, r, sysInfo)
	reg.Apps = append(reg.Apps, &AppEntry{
		ID:          "nerd-font",
		Name:        "Nerd Font",
		Description: "Polices avec icones pour terminal",
		Tags:        []installers.AppTag{installers.TagFont},
		Installer:   fontInstaller,
		Uninstaller: fontInstaller,
	})

	return reg
}

// GetByID returns an application by its ID.
func (r *Registry) GetByID(id string) *AppEntry {
	for _, app := range r.Apps {
		if app.ID == id {
			return app
		}
	}
	return nil
}

// GetAll returns all applications.
func (r *Registry) GetAll() []*AppEntry {
	return r.Apps
}
