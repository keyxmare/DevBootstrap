// Package zsh provides Zsh and Oh My Zsh installation strategies.
package zsh

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/installer/strategy"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// OhMyZshStrategy implements Oh My Zsh installation.
type OhMyZshStrategy struct {
	strategy.BaseStrategy
}

// NewOhMyZshStrategy creates a new Oh My Zsh installer strategy.
func NewOhMyZshStrategy(deps strategy.Dependencies, platform *entity.Platform) *OhMyZshStrategy {
	return &OhMyZshStrategy{
		BaseStrategy: strategy.NewBaseStrategy(deps, platform),
	}
}

// CheckStatus checks if Oh My Zsh is already installed.
func (s *OhMyZshStrategy) CheckStatus(ctx context.Context) (valueobject.AppStatus, string, error) {
	ohmyzshDir := filepath.Join(s.Platform.HomeDir(), ".oh-my-zsh")
	if s.Deps.FileSystem.Exists(ohmyzshDir) {
		return valueobject.StatusInstalled, "", nil
	}
	return valueobject.StatusNotInstalled, "", nil
}

// Install installs Oh My Zsh.
func (s *OhMyZshStrategy) Install(ctx context.Context, opts primary.InstallOptions) (*result.InstallResult, error) {
	// Check if Zsh is installed
	if !s.CommandExists("zsh") {
		return result.NewFailure("Zsh doit etre installe avant Oh My Zsh"), nil
	}

	// Check if already installed
	status, _, _ := s.CheckStatus(ctx)
	if status.IsInstalled() {
		s.Info("Oh My Zsh est deja installe")
		// In non-interactive mode, force reinstall
		if !opts.NoInteraction {
			if !s.Confirm("Voulez-vous reinstaller?", false) {
				return result.NewSuccess("Oh My Zsh deja installe"), nil
			}
		}
		// Remove existing installation for reinstall
		ohmyzshDir := filepath.Join(s.Platform.HomeDir(), ".oh-my-zsh")
		s.Deps.FileSystem.RemoveAll(ohmyzshDir)
	}

	s.Section("Installation de Oh My Zsh")

	// Clone Oh My Zsh
	ohmyzshDir := filepath.Join(s.Platform.HomeDir(), ".oh-my-zsh")
	res := s.Run(ctx,
		[]string{"git", "clone", "--depth=1", "https://github.com/ohmyzsh/ohmyzsh.git", ohmyzshDir},
		secondary.WithDescription("Telechargement de Oh My Zsh"),
		secondary.WithTimeout(5*time.Minute),
	)

	if !res.Success {
		return result.NewFailure("Echec du telechargement de Oh My Zsh", res.Stderr), nil
	}

	// Create .zshrc from template
	s.createZshrc(opts)

	// Install plugins
	var plugins []string
	if opts.ZshOptions != nil && len(opts.ZshOptions.Plugins) > 0 {
		plugins = opts.ZshOptions.Plugins
	}
	s.installPlugins(ctx, plugins)

	// Set Zsh as default shell
	if opts.NoInteraction {
		s.setDefaultShell(ctx)
	} else if s.Confirm("Definir Zsh comme shell par defaut?", true) {
		s.setDefaultShell(ctx)
	}

	s.Success("Oh My Zsh installe")
	return result.NewSuccess("Oh My Zsh installe avec succes"), nil
}

// createZshrc creates a .zshrc file.
func (s *OhMyZshStrategy) createZshrc(opts primary.InstallOptions) {
	theme := "robbyrussell"
	if opts.ZshOptions != nil && opts.ZshOptions.Theme != "" {
		theme = opts.ZshOptions.Theme
	}

	plugins := []string{"git", "zsh-autosuggestions", "zsh-syntax-highlighting"}
	if opts.ZshOptions != nil && len(opts.ZshOptions.Plugins) > 0 {
		plugins = opts.ZshOptions.Plugins
	}

	zshrcContent := fmt.Sprintf(`# Oh My Zsh configuration
export ZSH="$HOME/.oh-my-zsh"

# Theme
ZSH_THEME="%s"

# Plugins
plugins=(%s)

source $ZSH/oh-my-zsh.sh

# User configuration
export LANG=en_US.UTF-8
export EDITOR='nvim'

# Aliases
alias ll='ls -la'
alias la='ls -A'
alias l='ls -CF'
`, theme, strings.Join(plugins, " "))

	zshrcPath := filepath.Join(s.Platform.HomeDir(), ".zshrc")

	// Backup existing .zshrc
	if s.Deps.FileSystem.Exists(zshrcPath) {
		s.Deps.FileSystem.Copy(zshrcPath, zshrcPath+".backup")
	}

	s.Deps.FileSystem.WriteFile(zshrcPath, []byte(zshrcContent), 0644)
	s.Success(".zshrc cree")
}

// installPlugins installs Oh My Zsh plugins.
func (s *OhMyZshStrategy) installPlugins(ctx context.Context, plugins []string) {
	s.Section("Installation des plugins")
	customDir := filepath.Join(s.Platform.HomeDir(), ".oh-my-zsh/custom/plugins")

	// Essential plugins
	essentialPlugins := map[string]string{
		"zsh-autosuggestions":     "https://github.com/zsh-users/zsh-autosuggestions",
		"zsh-syntax-highlighting": "https://github.com/zsh-users/zsh-syntax-highlighting",
		"zsh-completions":         "https://github.com/zsh-users/zsh-completions",
	}

	for name, url := range essentialPlugins {
		pluginDir := filepath.Join(customDir, name)
		if s.Deps.FileSystem.Exists(pluginDir) {
			continue
		}

		s.Info("Installation de " + name + "...")
		res := s.Run(ctx,
			[]string{"git", "clone", "--depth=1", url, pluginDir},
			secondary.WithTimeout(2*time.Minute),
		)
		if res.Success {
			s.Success(name + " installe")
		} else {
			s.Warning("Echec: " + name)
		}
	}
}

// setDefaultShell sets Zsh as the default shell.
func (s *OhMyZshStrategy) setDefaultShell(ctx context.Context) {
	zshPath := s.Deps.Executor.GetCommandPath("zsh")
	if zshPath == "" {
		s.Warning("Impossible de trouver le chemin de Zsh")
		return
	}

	username := s.Platform.Username()
	if username == "" {
		s.Warning("Impossible de determiner le nom d'utilisateur")
		return
	}

	s.Info("Configuration de Zsh comme shell par defaut...")

	var res *secondary.CommandResult

	if s.Platform.IsMacOS() {
		// On macOS, use dscl to change shell
		res = s.Run(ctx,
			[]string{"dscl", ".", "-create", "/Users/" + username, "UserShell", zshPath},
			secondary.WithSudo(),
			secondary.WithDescription("Configuration du shell par defaut"),
		)
	} else {
		// On Linux, use sudo chsh
		res = s.Run(ctx,
			[]string{"chsh", "-s", zshPath, username},
			secondary.WithSudo(),
			secondary.WithDescription("Configuration du shell par defaut"),
		)
	}

	if res.Success {
		s.Success("Zsh defini comme shell par defaut")
	} else {
		s.Warning("Impossible de definir Zsh comme shell par defaut")
		if s.Platform.IsMacOS() {
			s.Info("Executez manuellement: sudo dscl . -create /Users/" + username + " UserShell " + zshPath)
		} else {
			s.Info("Executez manuellement: sudo chsh -s " + zshPath + " " + username)
		}
	}
}

// Verify verifies the Oh My Zsh installation.
func (s *OhMyZshStrategy) Verify(ctx context.Context) bool {
	ohmyzshDir := filepath.Join(s.Platform.HomeDir(), ".oh-my-zsh")
	return s.Deps.FileSystem.Exists(ohmyzshDir)
}

// Uninstall removes Oh My Zsh.
func (s *OhMyZshStrategy) Uninstall(ctx context.Context, opts primary.UninstallOptions) (*result.UninstallResult, error) {
	s.Section("Desinstallation de Oh My Zsh")

	if opts.RemoveOhMyZsh {
		ohmyzshDir := filepath.Join(s.Platform.HomeDir(), ".oh-my-zsh")
		s.Deps.FileSystem.RemoveAll(ohmyzshDir)
		s.Success("Oh My Zsh supprime")
	}

	if opts.RemoveZshrc {
		zshrcPath := filepath.Join(s.Platform.HomeDir(), ".zshrc")
		s.Deps.FileSystem.RemoveAll(zshrcPath)

		// Restore backup if exists
		backupPath := zshrcPath + ".backup"
		if s.Deps.FileSystem.Exists(backupPath) {
			s.Deps.FileSystem.Copy(backupPath, zshrcPath)
			s.Deps.FileSystem.RemoveAll(backupPath)
			s.Info(".zshrc restaure depuis la sauvegarde")
		}
	}

	return result.NewUninstallSuccess("Oh My Zsh desinstalle avec succes"), nil
}

// NewOhMyZshInstallerStrategy creates the Oh My Zsh installer strategy.
func NewOhMyZshInstallerStrategy(deps strategy.Dependencies, platform *entity.Platform) (strategy.InstallerStrategy, error) {
	return NewOhMyZshStrategy(deps, platform), nil
}
