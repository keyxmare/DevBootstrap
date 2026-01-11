// Package zsh provides Zsh and Oh My Zsh installation functionality.
package zsh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/cli"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// Available themes
var availableThemes = []string{
	"robbyrussell",
	"agnoster",
	"powerlevel10k/powerlevel10k",
	"spaceship",
}

// Available plugins
var availablePlugins = []string{
	"zsh-autosuggestions",
	"zsh-syntax-highlighting",
	"zsh-completions",
}

// Installer handles Zsh installation.
type Installer struct {
	*installers.BaseInstaller
}

// NewInstaller creates a new Zsh installer.
func NewInstaller(c *cli.CLI, r *runner.Runner, sysInfo *system.SystemInfo) *Installer {
	return &Installer{
		BaseInstaller: installers.NewBaseInstaller(c, r, sysInfo),
	}
}

// Name returns the application name.
func (i *Installer) Name() string {
	return "Zsh"
}

// ID returns the application ID.
func (i *Installer) ID() string {
	return "zsh"
}

// Description returns the application description.
func (i *Installer) Description() string {
	return "Shell Z moderne"
}

// Tags returns the application tags.
func (i *Installer) Tags() []installers.AppTag {
	return []installers.AppTag{installers.TagApp, installers.TagShell}
}

// CheckExisting checks if Zsh is already installed.
func (i *Installer) CheckExisting() (installers.AppStatus, string) {
	if i.Runner.CommandExists("zsh") {
		version := i.Runner.GetCommandVersion("zsh", "--version")
		return installers.StatusInstalled, version
	}
	return installers.StatusNotInstalled, ""
}

// Install installs Zsh.
func (i *Installer) Install(opts *installers.InstallOptions) *installers.InstallResult {
	// Check if already installed
	status, version := i.CheckExisting()
	if status == installers.StatusInstalled {
		i.CLI.PrintInfo("Zsh est deja installe: " + version)
		if !i.CLI.AskYesNo("Voulez-vous reinstaller?", false) {
			return installers.NewSuccessResult("Zsh deja installe")
		}
	}

	if i.SystemInfo.IsMacOS() {
		return i.installMacOS(opts)
	} else if i.SystemInfo.IsDebian() {
		return i.installUbuntu(opts)
	}

	return installers.NewFailureResult("Systeme non supporte")
}

// installMacOS installs Zsh on macOS (usually pre-installed).
func (i *Installer) installMacOS(opts *installers.InstallOptions) *installers.InstallResult {
	// Zsh is usually pre-installed on macOS
	if i.Runner.CommandExists("zsh") {
		i.CLI.PrintSuccess("Zsh est deja installe sur macOS")
		return installers.NewSuccessResult("Zsh deja installe")
	}

	brewPath := i.getHomebrewPath()
	if brewPath == "" {
		return installers.NewFailureResult("Homebrew est requis pour installer Zsh sur macOS")
	}

	i.CLI.PrintSection("Installation de Zsh")
	result := i.Runner.Run(
		[]string{brewPath, "install", "zsh"},
		runner.WithDescription("Installation de Zsh"),
		runner.WithTimeout(5*time.Minute),
	)

	if !result.Success {
		return installers.NewFailureResult("Echec de l'installation de Zsh", result.Stderr)
	}

	i.CLI.PrintSuccess("Zsh installe")
	return installers.NewSuccessResult("Zsh installe avec succes")
}

// installUbuntu installs Zsh on Ubuntu/Debian.
func (i *Installer) installUbuntu(opts *installers.InstallOptions) *installers.InstallResult {
	i.CLI.PrintSection("Installation de Zsh")

	result := i.Runner.Run(
		[]string{"apt-get", "install", "-y", "zsh"},
		runner.WithSudo(),
		runner.WithDescription("Installation de Zsh"),
		runner.WithTimeout(5*time.Minute),
	)

	if !result.Success {
		return installers.NewFailureResult("Echec de l'installation de Zsh", result.Stderr)
	}

	i.CLI.PrintSuccess("Zsh installe")
	return installers.NewSuccessResult("Zsh installe avec succes")
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

// Verify verifies the Zsh installation.
func (i *Installer) Verify() bool {
	return i.Runner.CommandExists("zsh")
}

// Uninstall removes Zsh.
func (i *Installer) Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult {
	i.CLI.PrintSection("Desinstallation de Zsh")

	// Don't uninstall Zsh on macOS (system shell)
	if i.SystemInfo.IsMacOS() {
		i.CLI.PrintWarning("Zsh est un composant systeme sur macOS, desinstallation non recommandee")
		return installers.NewUninstallSuccessResult("Zsh conserve (composant systeme)")
	}

	if i.SystemInfo.IsDebian() {
		i.Runner.Run([]string{"apt-get", "purge", "-y", "zsh"}, runner.WithSudo())
	}

	if opts.RemoveZshrc {
		i.Runner.RemoveAll(filepath.Join(i.SystemInfo.HomeDir, ".zshrc"))
	}

	i.CLI.PrintSuccess("Zsh desinstalle")
	return installers.NewUninstallSuccessResult("Zsh desinstalle avec succes")
}

// OhMyZshInstaller handles Oh My Zsh installation.
type OhMyZshInstaller struct {
	*installers.BaseInstaller
}

// NewOhMyZshInstaller creates a new Oh My Zsh installer.
func NewOhMyZshInstaller(c *cli.CLI, r *runner.Runner, sysInfo *system.SystemInfo) *OhMyZshInstaller {
	return &OhMyZshInstaller{
		BaseInstaller: installers.NewBaseInstaller(c, r, sysInfo),
	}
}

// Name returns the application name.
func (i *OhMyZshInstaller) Name() string {
	return "Oh My Zsh"
}

// ID returns the application ID.
func (i *OhMyZshInstaller) ID() string {
	return "oh-my-zsh"
}

// Description returns the application description.
func (i *OhMyZshInstaller) Description() string {
	return "Framework de configuration pour Zsh avec plugins"
}

// Tags returns the application tags.
func (i *OhMyZshInstaller) Tags() []installers.AppTag {
	return []installers.AppTag{installers.TagConfig}
}

// CheckExisting checks if Oh My Zsh is already installed.
func (i *OhMyZshInstaller) CheckExisting() (installers.AppStatus, string) {
	ohmyzshDir := filepath.Join(i.SystemInfo.HomeDir, ".oh-my-zsh")
	if _, err := os.Stat(ohmyzshDir); err == nil {
		return installers.StatusInstalled, ""
	}
	return installers.StatusNotInstalled, ""
}

// Install installs Oh My Zsh.
func (i *OhMyZshInstaller) Install(opts *installers.InstallOptions) *installers.InstallResult {
	// Check if Zsh is installed
	if !i.Runner.CommandExists("zsh") {
		return installers.NewFailureResult("Zsh doit etre installe avant Oh My Zsh")
	}

	// Check if already installed
	status, _ := i.CheckExisting()
	if status == installers.StatusInstalled {
		i.CLI.PrintInfo("Oh My Zsh est deja installe")
		if !i.CLI.AskYesNo("Voulez-vous reinstaller?", false) {
			return installers.NewSuccessResult("Oh My Zsh deja installe")
		}
		// Remove existing installation
		ohmyzshDir := filepath.Join(i.SystemInfo.HomeDir, ".oh-my-zsh")
		i.Runner.RemoveAll(ohmyzshDir)
	}

	i.CLI.PrintSection("Installation de Oh My Zsh")

	// Clone Oh My Zsh
	ohmyzshDir := filepath.Join(i.SystemInfo.HomeDir, ".oh-my-zsh")
	result := i.Runner.Run(
		[]string{"git", "clone", "--depth=1", "https://github.com/ohmyzsh/ohmyzsh.git", ohmyzshDir},
		runner.WithDescription("Telechargement de Oh My Zsh"),
		runner.WithTimeout(5*time.Minute),
	)

	if !result.Success {
		return installers.NewFailureResult("Echec du telechargement de Oh My Zsh", result.Stderr)
	}

	// Create .zshrc from template
	i.createZshrc(opts)

	// Install plugins
	i.installPlugins(opts.ZshPlugins)

	// Set Zsh as default shell
	// In NoInteraction mode (TUI), set automatically. Otherwise, ask.
	if opts.NoInteraction {
		i.setDefaultShell()
	} else if i.CLI.AskYesNo("Definir Zsh comme shell par defaut?", true) {
		i.setDefaultShell()
	}

	i.CLI.PrintSuccess("Oh My Zsh installe")
	return installers.NewSuccessResult("Oh My Zsh installe avec succes")
}

// createZshrc creates a .zshrc file.
func (i *OhMyZshInstaller) createZshrc(opts *installers.InstallOptions) {
	theme := opts.ZshTheme
	if theme == "" {
		theme = "robbyrussell"
	}

	plugins := opts.ZshPlugins
	if len(plugins) == 0 {
		plugins = []string{"git", "zsh-autosuggestions", "zsh-syntax-highlighting"}
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

	zshrcPath := filepath.Join(i.SystemInfo.HomeDir, ".zshrc")

	// Backup existing .zshrc
	if _, err := os.Stat(zshrcPath); err == nil {
		os.Rename(zshrcPath, zshrcPath+".backup")
	}

	os.WriteFile(zshrcPath, []byte(zshrcContent), 0644)
	i.CLI.PrintSuccess(".zshrc cree")
}

// installPlugins installs Oh My Zsh plugins.
func (i *OhMyZshInstaller) installPlugins(plugins []string) {
	i.CLI.PrintSection("Installation des plugins")
	customDir := filepath.Join(i.SystemInfo.HomeDir, ".oh-my-zsh/custom/plugins")

	// Always install these essential plugins
	essentialPlugins := map[string]string{
		"zsh-autosuggestions":    "https://github.com/zsh-users/zsh-autosuggestions",
		"zsh-syntax-highlighting": "https://github.com/zsh-users/zsh-syntax-highlighting",
		"zsh-completions":        "https://github.com/zsh-users/zsh-completions",
	}

	for name, url := range essentialPlugins {
		pluginDir := filepath.Join(customDir, name)
		if _, err := os.Stat(pluginDir); err == nil {
			continue // Already installed
		}

		i.CLI.PrintInfo("Installation de " + name + "...")
		result := i.Runner.Run(
			[]string{"git", "clone", "--depth=1", url, pluginDir},
			runner.WithTimeout(2*time.Minute),
		)
		if result.Success {
			i.CLI.PrintSuccess(name + " installe")
		} else {
			i.CLI.PrintWarning("Echec: " + name)
		}
	}
}

// setDefaultShell sets Zsh as the default shell.
func (i *OhMyZshInstaller) setDefaultShell() {
	zshPath := i.Runner.GetCommandPath("zsh")
	if zshPath == "" {
		i.CLI.PrintWarning("Impossible de trouver le chemin de Zsh")
		return
	}

	username := i.SystemInfo.Username
	if username == "" {
		i.CLI.PrintWarning("Impossible de determiner le nom d'utilisateur")
		return
	}

	i.CLI.PrintInfo("Configuration de Zsh comme shell par defaut...")

	var result *runner.Result

	if i.SystemInfo.IsMacOS() {
		// On macOS, use dscl to change shell (chsh always prompts for password)
		// dscl . -change /Users/<username> UserShell <old_shell> <new_shell>
		// Or simply: dscl . -create /Users/<username> UserShell <new_shell>
		result = i.Runner.Run(
			[]string{"dscl", ".", "-create", "/Users/" + username, "UserShell", zshPath},
			runner.WithSudo(),
			runner.WithDescription("Configuration du shell par defaut"),
		)
	} else {
		// On Linux, use sudo chsh
		result = i.Runner.Run(
			[]string{"chsh", "-s", zshPath, username},
			runner.WithSudo(),
			runner.WithDescription("Configuration du shell par defaut"),
		)
	}

	if result.Success {
		i.CLI.PrintSuccess("Zsh defini comme shell par defaut")
	} else {
		i.CLI.PrintWarning("Impossible de definir Zsh comme shell par defaut")
		if i.SystemInfo.IsMacOS() {
			i.CLI.PrintInfo("Executez manuellement: sudo dscl . -create /Users/" + username + " UserShell " + zshPath)
		} else {
			i.CLI.PrintInfo("Executez manuellement: sudo chsh -s " + zshPath + " " + username)
		}
	}
}

// Verify verifies the Oh My Zsh installation.
func (i *OhMyZshInstaller) Verify() bool {
	ohmyzshDir := filepath.Join(i.SystemInfo.HomeDir, ".oh-my-zsh")
	_, err := os.Stat(ohmyzshDir)
	return err == nil
}

// Uninstall removes Oh My Zsh.
func (i *OhMyZshInstaller) Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult {
	i.CLI.PrintSection("Desinstallation de Oh My Zsh")

	if opts.RemoveOhMyZsh {
		ohmyzshDir := filepath.Join(i.SystemInfo.HomeDir, ".oh-my-zsh")
		i.Runner.RemoveAll(ohmyzshDir)
		i.CLI.PrintSuccess("Oh My Zsh supprime")
	}

	if opts.RemoveZshrc {
		zshrcPath := filepath.Join(i.SystemInfo.HomeDir, ".zshrc")
		i.Runner.RemoveAll(zshrcPath)

		// Restore backup if exists
		backupPath := zshrcPath + ".backup"
		if _, err := os.Stat(backupPath); err == nil {
			os.Rename(backupPath, zshrcPath)
			i.CLI.PrintInfo(".zshrc restaure depuis la sauvegarde")
		}
	}

	return installers.NewUninstallSuccessResult("Oh My Zsh desinstalle avec succes")
}
