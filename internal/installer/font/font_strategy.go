// Package font provides Nerd Font installation strategy.
package font

import (
	"context"
	"fmt"
	"os"
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

// FontInfo represents a Nerd Font.
type FontInfo struct {
	ID           string
	Name         string
	HomebrewCask string
}

// AvailableFonts lists the available Nerd Fonts.
var AvailableFonts = []FontInfo{
	{ID: "meslo", Name: "MesloLG Nerd Font", HomebrewCask: "font-meslo-lg-nerd-font"},
	{ID: "firacode", Name: "FiraCode Nerd Font", HomebrewCask: "font-fira-code-nerd-font"},
	{ID: "jetbrains", Name: "JetBrainsMono Nerd Font", HomebrewCask: "font-jetbrains-mono-nerd-font"},
	{ID: "hack", Name: "Hack Nerd Font", HomebrewCask: "font-hack-nerd-font"},
}

// Strategy implements Nerd Font installation.
type Strategy struct {
	strategy.BaseStrategy
}

// NewStrategy creates a new Nerd Font installer strategy.
func NewStrategy(deps strategy.Dependencies, platform *entity.Platform) *Strategy {
	return &Strategy{
		BaseStrategy: strategy.NewBaseStrategy(deps, platform),
	}
}

// CheckStatus checks if any Nerd Font is already installed.
func (s *Strategy) CheckStatus(ctx context.Context) (valueobject.AppStatus, string, error) {
	if s.Platform.IsMacOS() {
		// Check via Homebrew
		brewPath := s.getHomebrewPath()
		if brewPath != "" {
			res := s.Run(ctx, []string{brewPath, "list", "--cask"}, secondary.WithSkipDryRun())
			if res.Success {
				for _, font := range AvailableFonts {
					if strings.Contains(res.Stdout, font.HomebrewCask) {
						return valueobject.StatusInstalled, font.Name, nil
					}
				}
			}
		}

		// Check in ~/Library/Fonts
		fontsDir := filepath.Join(s.Platform.HomeDir(), "Library/Fonts")
		if s.hasNerdFont(fontsDir) {
			return valueobject.StatusInstalled, "MesloLG Nerd Font", nil
		}
	} else {
		// Linux: Check in ~/.local/share/fonts
		fontsDir := filepath.Join(s.Platform.HomeDir(), ".local/share/fonts")
		if s.hasNerdFont(fontsDir) {
			return valueobject.StatusInstalled, "MesloLG Nerd Font", nil
		}
	}

	return valueobject.StatusNotInstalled, "", nil
}

// hasNerdFont checks if a directory contains Nerd Font files.
func (s *Strategy) hasNerdFont(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, "Nerd") || strings.Contains(name, "Meslo") {
			return true
		}
	}
	return false
}

// Install installs Nerd Fonts.
func (s *Strategy) Install(ctx context.Context, opts primary.InstallOptions) (*result.InstallResult, error) {
	// Check if already installed
	status, version, _ := s.CheckStatus(ctx)
	if status.IsInstalled() {
		s.Info("Une Nerd Font est deja installee: " + version)
		if !s.Confirm("Voulez-vous installer d'autres polices?", false) {
			return result.NewSuccess("Nerd Font deja installee"), nil
		}
	}

	// Select font to install (default to MesloLG)
	selectedFont := AvailableFonts[0]

	if s.Platform.IsMacOS() {
		return s.installMacOS(ctx, selectedFont), nil
	} else if s.Platform.IsDebian() {
		return s.installLinux(ctx, selectedFont), nil
	}

	return result.NewFailure("Systeme non supporte"), nil
}

// installMacOS installs a Nerd Font on macOS using Homebrew.
func (s *Strategy) installMacOS(ctx context.Context, font FontInfo) *result.InstallResult {
	brewPath := s.getHomebrewPath()
	if brewPath == "" {
		return result.NewFailure("Homebrew est requis pour installer les polices sur macOS")
	}

	s.Section("Installation de " + font.Name)

	// Add homebrew/cask-fonts tap
	s.Info("Ajout du depot homebrew/cask-fonts...")
	s.Run(ctx, []string{brewPath, "tap", "homebrew/cask-fonts"})

	// Install font
	res := s.Run(ctx,
		[]string{brewPath, "install", "--cask", font.HomebrewCask},
		secondary.WithDescription("Installation de "+font.Name),
		secondary.WithTimeout(5*time.Minute),
	)

	if !res.Success {
		return result.NewFailure("Echec de l'installation de "+font.Name, res.Stderr)
	}

	s.Success(font.Name + " installee")
	s.showFontInstructions(font)

	return result.NewSuccess(font.Name + " installee avec succes")
}

// installLinux installs a Nerd Font on Linux.
func (s *Strategy) installLinux(ctx context.Context, font FontInfo) *result.InstallResult {
	s.Section("Installation de " + font.Name)

	// Create fonts directory
	fontsDir := filepath.Join(s.Platform.HomeDir(), ".local/share/fonts")
	s.Deps.FileSystem.MkdirAll(fontsDir, 0755)

	// Download font from GitHub releases
	fontURL := "https://github.com/ryanoasis/nerd-fonts/releases/latest/download/Meslo.zip"
	zipPath := filepath.Join(os.TempDir(), "nerd-font.zip")

	s.Info("Telechargement de " + font.Name + "...")
	err := s.Deps.HTTPClient.Download(ctx, fontURL, zipPath)
	if err != nil {
		return result.NewFailure("Echec du telechargement de la police", err.Error())
	}

	// Extract to fonts directory
	s.Info("Extraction de la police...")
	res := s.Run(ctx,
		[]string{"unzip", "-o", zipPath, "-d", fontsDir},
		secondary.WithTimeout(2*time.Minute),
	)

	if !res.Success {
		return result.NewFailure("Echec de l'extraction de la police", res.Stderr)
	}

	// Update font cache
	s.Info("Mise a jour du cache des polices...")
	s.Run(ctx, []string{"fc-cache", "-fv"})

	// Cleanup
	os.Remove(zipPath)

	s.Success(font.Name + " installee")
	s.showFontInstructions(font)

	return result.NewSuccess(font.Name + " installee avec succes")
}

// showFontInstructions shows instructions for using the font.
func (s *Strategy) showFontInstructions(font FontInfo) {
	s.Section("Configuration")
	s.Info("Pour utiliser cette police:")
	s.Info("  1. Ouvrez les preferences de votre terminal")
	s.Info("  2. Selectionnez '" + font.Name + "' comme police")
	s.Info("  3. Redemarrez le terminal")
	s.Info("Cette police est recommandee pour le theme 'agnoster' de Oh My Zsh")
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

// Verify verifies the Nerd Font installation.
func (s *Strategy) Verify(ctx context.Context) bool {
	status, _, _ := s.CheckStatus(context.Background())
	return status.IsInstalled()
}

// Uninstall removes Nerd Fonts.
func (s *Strategy) Uninstall(ctx context.Context, opts primary.UninstallOptions) (*result.UninstallResult, error) {
	s.Section("Desinstallation des Nerd Fonts")

	if s.Platform.IsMacOS() {
		brewPath := s.getHomebrewPath()
		if brewPath != "" {
			for _, font := range AvailableFonts {
				s.Run(ctx, []string{brewPath, "uninstall", "--cask", font.HomebrewCask})
			}
		}
	} else {
		// Linux: Remove font files
		fontsDir := filepath.Join(s.Platform.HomeDir(), ".local/share/fonts")
		entries, _ := os.ReadDir(fontsDir)
		for _, entry := range entries {
			name := entry.Name()
			if strings.Contains(name, "Nerd") || strings.Contains(name, "Meslo") {
				os.Remove(filepath.Join(fontsDir, name))
			}
		}
		// Update font cache
		s.Run(ctx, []string{"fc-cache", "-fv"})
	}

	s.Success("Nerd Fonts desinstallees")
	return result.NewUninstallSuccess("Nerd Fonts desinstallees avec succes"), nil
}

// NewFontStrategy creates the Nerd Font installer strategy.
func NewFontStrategy(deps strategy.Dependencies, platform *entity.Platform) (strategy.InstallerStrategy, error) {
	if !platform.IsMacOS() && !platform.IsDebian() {
		return nil, fmt.Errorf("unsupported platform for Nerd Font: %s", platform.OS())
	}
	return NewStrategy(deps, platform), nil
}
