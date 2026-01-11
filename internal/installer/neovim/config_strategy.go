package neovim

import (
	"context"
	"path/filepath"

	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/installer/strategy"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
)

// ConfigStrategy implements Neovim configuration installation.
type ConfigStrategy struct {
	strategy.BaseStrategy
}

// NewConfigStrategy creates a new Neovim config installer strategy.
func NewConfigStrategy(deps strategy.Dependencies, platform *entity.Platform) *ConfigStrategy {
	return &ConfigStrategy{
		BaseStrategy: strategy.NewBaseStrategy(deps, platform),
	}
}

// CheckStatus checks if Neovim config is already installed.
func (s *ConfigStrategy) CheckStatus(ctx context.Context) (valueobject.AppStatus, string, error) {
	configPath := filepath.Join(s.Platform.HomeDir(), ".config/nvim/init.lua")
	if s.Deps.FileSystem.Exists(configPath) {
		return valueobject.StatusInstalled, "", nil
	}
	return valueobject.StatusNotInstalled, "", nil
}

// Install installs Neovim configuration.
func (s *ConfigStrategy) Install(ctx context.Context, opts primary.InstallOptions) (*result.InstallResult, error) {
	if !s.CommandExists("nvim") {
		return result.NewFailure("Neovim doit etre installe avant la configuration"), nil
	}

	configDir := filepath.Join(s.Platform.HomeDir(), ".config/nvim")

	// Backup existing config
	if s.Deps.FileSystem.Exists(configDir) {
		s.Info("Sauvegarde de la configuration existante...")
		backupDir := configDir + ".backup"
		s.Deps.FileSystem.Rename(configDir, backupDir)
	}

	s.Section("Installation de la configuration Neovim")

	if err := s.Deps.FileSystem.MkdirAll(configDir, 0755); err != nil {
		return result.NewFailure("Impossible de creer le repertoire de config", err.Error()), nil
	}

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
	if err := s.Deps.FileSystem.WriteFile(initPath, []byte(minimalConfig), 0644); err != nil {
		return result.NewFailure("Echec de l'ecriture de la configuration", err.Error()), nil
	}

	s.Success("Configuration Neovim installee")
	return result.NewSuccess("Configuration installee avec succes"), nil
}

// Verify verifies the Neovim config installation.
func (s *ConfigStrategy) Verify(ctx context.Context) bool {
	configPath := filepath.Join(s.Platform.HomeDir(), ".config/nvim/init.lua")
	return s.Deps.FileSystem.Exists(configPath)
}

// Uninstall removes Neovim configuration.
func (s *ConfigStrategy) Uninstall(ctx context.Context, opts primary.UninstallOptions) (*result.UninstallResult, error) {
	configDir := filepath.Join(s.Platform.HomeDir(), ".config/nvim")
	s.Deps.FileSystem.RemoveAll(configDir)
	s.Success("Configuration Neovim supprimee")
	return result.NewUninstallSuccess("Configuration supprimee avec succes"), nil
}

// NewNeovimConfigStrategy creates the Neovim config installer strategy.
func NewNeovimConfigStrategy(deps strategy.Dependencies, platform *entity.Platform) (strategy.InstallerStrategy, error) {
	return NewConfigStrategy(deps, platform), nil
}
