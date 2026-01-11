// Package neovim provides Neovim installation functionality.
package neovim

import (
	"os"
	"path/filepath"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/cli"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// Installer handles Neovim installation.
type Installer struct {
	*installers.BaseInstaller
}

// NewInstaller creates a new Neovim installer.
func NewInstaller(c *cli.CLI, r *runner.Runner, sysInfo *system.SystemInfo) *Installer {
	return &Installer{
		BaseInstaller: installers.NewBaseInstaller(c, r, sysInfo),
	}
}

// Name returns the application name.
func (i *Installer) Name() string {
	return "Neovim"
}

// ID returns the application ID.
func (i *Installer) ID() string {
	return "neovim"
}

// Description returns the application description.
func (i *Installer) Description() string {
	return "Editeur de texte moderne"
}

// Tags returns the application tags.
func (i *Installer) Tags() []installers.AppTag {
	return []installers.AppTag{installers.TagApp, installers.TagEditor}
}

// CheckExisting checks if Neovim is already installed.
func (i *Installer) CheckExisting() (installers.AppStatus, string) {
	if i.Runner.CommandExists("nvim") {
		version := i.Runner.GetCommandVersion("nvim", "--version")
		return installers.StatusInstalled, version
	}
	return installers.StatusNotInstalled, ""
}

// Install installs Neovim.
func (i *Installer) Install(opts *installers.InstallOptions) *installers.InstallResult {
	// Check if already installed
	status, version := i.CheckExisting()
	if status == installers.StatusInstalled {
		i.CLI.PrintInfo("Neovim est deja installe: " + version)
		if !i.CLI.AskYesNo("Voulez-vous reinstaller?", false) {
			return installers.NewSuccessResult("Neovim deja installe")
		}
	}

	if i.SystemInfo.IsMacOS() {
		return i.installMacOS(opts)
	} else if i.SystemInfo.IsDebian() {
		return i.installUbuntu(opts)
	}

	return installers.NewFailureResult("Systeme non supporte")
}

// installMacOS installs Neovim on macOS using Homebrew.
func (i *Installer) installMacOS(opts *installers.InstallOptions) *installers.InstallResult {
	brewPath := i.getHomebrewPath()
	if brewPath == "" {
		return installers.NewFailureResult("Homebrew est requis pour installer Neovim sur macOS")
	}

	i.CLI.PrintSection("Installation de Neovim")
	i.CLI.PrintInfo("Installation via Homebrew...")

	result := i.Runner.Run(
		[]string{brewPath, "install", "neovim"},
		runner.WithDescription("Installation de Neovim"),
		runner.WithTimeout(10*time.Minute),
	)

	if !result.Success {
		return installers.NewFailureResult("Echec de l'installation de Neovim", result.Stderr)
	}

	// Install dependencies
	i.installDependencies()

	i.CLI.PrintSuccess("Neovim installe")
	return installers.NewSuccessResult("Neovim installe avec succes")
}

// installUbuntu installs Neovim on Ubuntu/Debian.
func (i *Installer) installUbuntu(opts *installers.InstallOptions) *installers.InstallResult {
	i.CLI.PrintSection("Installation de Neovim")

	// Try to use PPA for latest version
	i.CLI.PrintInfo("Ajout du PPA Neovim...")
	i.Runner.Run([]string{"add-apt-repository", "-y", "ppa:neovim-ppa/unstable"}, runner.WithSudo())
	i.Runner.Run([]string{"apt-get", "update"}, runner.WithSudo())

	result := i.Runner.Run(
		[]string{"apt-get", "install", "-y", "neovim"},
		runner.WithSudo(),
		runner.WithDescription("Installation de Neovim"),
		runner.WithTimeout(10*time.Minute),
	)

	if !result.Success {
		return installers.NewFailureResult("Echec de l'installation de Neovim", result.Stderr)
	}

	// Install dependencies
	i.installDependencies()

	i.CLI.PrintSuccess("Neovim installe")
	return installers.NewSuccessResult("Neovim installe avec succes")
}

// installDependencies installs common dependencies for Neovim.
func (i *Installer) installDependencies() {
	i.CLI.PrintSection("Installation des dependances")

	deps := []string{"ripgrep", "fd"}
	if i.SystemInfo.IsMacOS() {
		deps = append(deps, "fzf", "lazygit")
	}

	for _, dep := range deps {
		i.CLI.PrintInfo("Installation de " + dep + "...")
		if i.SystemInfo.IsMacOS() {
			brewPath := i.getHomebrewPath()
			i.Runner.Run([]string{brewPath, "install", dep})
		} else {
			// fd is called fd-find on Ubuntu
			pkgName := dep
			if dep == "fd" {
				pkgName = "fd-find"
			}
			i.Runner.Run([]string{"apt-get", "install", "-y", pkgName}, runner.WithSudo())
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

// Verify verifies the Neovim installation.
func (i *Installer) Verify() bool {
	return i.Runner.CommandExists("nvim")
}

// Uninstall removes Neovim.
func (i *Installer) Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult {
	i.CLI.PrintSection("Desinstallation de Neovim")

	if i.SystemInfo.IsMacOS() {
		brewPath := i.getHomebrewPath()
		if brewPath != "" {
			i.Runner.Run([]string{brewPath, "uninstall", "neovim"})
		}
	} else if i.SystemInfo.IsDebian() {
		i.Runner.Run([]string{"apt-get", "purge", "-y", "neovim"}, runner.WithSudo())
	}

	if opts.RemoveConfig {
		i.CLI.PrintInfo("Suppression de la configuration...")
		i.Runner.RemoveAll(filepath.Join(i.SystemInfo.HomeDir, ".config/nvim"))
	}

	if opts.RemoveData {
		i.CLI.PrintInfo("Suppression des donnees...")
		i.Runner.RemoveAll(filepath.Join(i.SystemInfo.HomeDir, ".local/share/nvim"))
		i.Runner.RemoveAll(filepath.Join(i.SystemInfo.HomeDir, ".local/state/nvim"))
	}

	if opts.RemoveCache {
		i.CLI.PrintInfo("Suppression du cache...")
		i.Runner.RemoveAll(filepath.Join(i.SystemInfo.HomeDir, ".cache/nvim"))
	}

	i.CLI.PrintSuccess("Neovim desinstalle")
	return installers.NewUninstallSuccessResult("Neovim desinstalle avec succes")
}

// ConfigInstaller handles Neovim configuration installation.
type ConfigInstaller struct {
	*installers.BaseInstaller
}

// NewConfigInstaller creates a new Neovim config installer.
func NewConfigInstaller(c *cli.CLI, r *runner.Runner, sysInfo *system.SystemInfo) *ConfigInstaller {
	return &ConfigInstaller{
		BaseInstaller: installers.NewBaseInstaller(c, r, sysInfo),
	}
}

// Name returns the application name.
func (i *ConfigInstaller) Name() string {
	return "Neovim Config"
}

// ID returns the application ID.
func (i *ConfigInstaller) ID() string {
	return "neovim-config"
}

// Description returns the application description.
func (i *ConfigInstaller) Description() string {
	return "Configuration et plugins pour Neovim"
}

// Tags returns the application tags.
func (i *ConfigInstaller) Tags() []installers.AppTag {
	return []installers.AppTag{installers.TagConfig}
}

// CheckExisting checks if Neovim config is already installed.
func (i *ConfigInstaller) CheckExisting() (installers.AppStatus, string) {
	configPath := filepath.Join(i.SystemInfo.HomeDir, ".config/nvim/init.lua")
	if _, err := os.Stat(configPath); err == nil {
		return installers.StatusInstalled, ""
	}
	return installers.StatusNotInstalled, ""
}

// Install installs Neovim configuration.
func (i *ConfigInstaller) Install(opts *installers.InstallOptions) *installers.InstallResult {
	// Check if Neovim is installed
	if !i.Runner.CommandExists("nvim") {
		return installers.NewFailureResult("Neovim doit etre installe avant la configuration")
	}

	configDir := filepath.Join(i.SystemInfo.HomeDir, ".config/nvim")

	// Backup existing config
	if _, err := os.Stat(configDir); err == nil {
		i.CLI.PrintInfo("Sauvegarde de la configuration existante...")
		backupDir := configDir + ".backup"
		os.Rename(configDir, backupDir)
	}

	i.CLI.PrintSection("Installation de la configuration Neovim")

	// Create minimal config
	i.Runner.EnsureDirectory(configDir, 0755)

	minimalConfig := `-- Neovim configuration
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.tabstop = 4
vim.opt.shiftwidth = 4
vim.opt.expandtab = true
vim.opt.smartindent = true
vim.opt.wrap = false
vim.opt.termguicolors = true
vim.opt.scrolloff = 8
vim.opt.signcolumn = "yes"
vim.opt.updatetime = 50

-- Leader key
vim.g.mapleader = " "

-- Basic keymaps
vim.keymap.set("n", "<leader>w", ":w<CR>")
vim.keymap.set("n", "<leader>q", ":q<CR>")
vim.keymap.set("n", "<Esc>", ":noh<CR>")
`

	initPath := filepath.Join(configDir, "init.lua")
	if err := os.WriteFile(initPath, []byte(minimalConfig), 0644); err != nil {
		return installers.NewFailureResult("Echec de l'ecriture de la configuration", err.Error())
	}

	i.CLI.PrintSuccess("Configuration Neovim installee")
	return installers.NewSuccessResult("Configuration installee avec succes")
}

// Verify verifies the Neovim config installation.
func (i *ConfigInstaller) Verify() bool {
	configPath := filepath.Join(i.SystemInfo.HomeDir, ".config/nvim/init.lua")
	_, err := os.Stat(configPath)
	return err == nil
}

// Uninstall removes Neovim configuration.
func (i *ConfigInstaller) Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult {
	configDir := filepath.Join(i.SystemInfo.HomeDir, ".config/nvim")
	i.Runner.RemoveAll(configDir)
	i.CLI.PrintSuccess("Configuration Neovim supprimee")
	return installers.NewUninstallSuccessResult("Configuration supprimee avec succes")
}
