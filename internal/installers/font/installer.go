// Package font provides Nerd Font installation functionality.
package font

import (
	"os"
	"path/filepath"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/cli"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

// FontInfo represents a Nerd Font.
type FontInfo struct {
	ID           string
	Name         string
	HomebrewCask string
}

// Available fonts
var availableFonts = []FontInfo{
	{ID: "meslo", Name: "MesloLG Nerd Font", HomebrewCask: "font-meslo-lg-nerd-font"},
	{ID: "firacode", Name: "FiraCode Nerd Font", HomebrewCask: "font-fira-code-nerd-font"},
	{ID: "jetbrains", Name: "JetBrainsMono Nerd Font", HomebrewCask: "font-jetbrains-mono-nerd-font"},
	{ID: "hack", Name: "Hack Nerd Font", HomebrewCask: "font-hack-nerd-font"},
}

// Installer handles Nerd Font installation.
type Installer struct {
	*installers.BaseInstaller
}

// NewInstaller creates a new Nerd Font installer.
func NewInstaller(c *cli.CLI, r *runner.Runner, sysInfo *system.SystemInfo) *Installer {
	return &Installer{
		BaseInstaller: installers.NewBaseInstaller(c, r, sysInfo),
	}
}

// Name returns the application name.
func (i *Installer) Name() string {
	return "Nerd Font"
}

// ID returns the application ID.
func (i *Installer) ID() string {
	return "nerd-font"
}

// Description returns the application description.
func (i *Installer) Description() string {
	return "Polices avec icones pour terminal"
}

// Tags returns the application tags.
func (i *Installer) Tags() []installers.AppTag {
	return []installers.AppTag{installers.TagFont}
}

// CheckExisting checks if any Nerd Font is already installed.
func (i *Installer) CheckExisting() (installers.AppStatus, string) {
	if i.SystemInfo.IsMacOS() {
		// Check via Homebrew
		brewPath := i.getHomebrewPath()
		if brewPath != "" {
			result := i.Runner.Run([]string{brewPath, "list", "--cask"})
			if result.Success {
				for _, font := range availableFonts {
					if contains(result.Stdout, font.HomebrewCask) {
						return installers.StatusInstalled, font.Name
					}
				}
			}
		}

		// Check in ~/Library/Fonts
		fontsDir := filepath.Join(i.SystemInfo.HomeDir, "Library/Fonts")
		if hasNerdFont(fontsDir) {
			return installers.StatusInstalled, "MesloLG Nerd Font"
		}
	} else {
		// Linux: Check in ~/.local/share/fonts
		fontsDir := filepath.Join(i.SystemInfo.HomeDir, ".local/share/fonts")
		if hasNerdFont(fontsDir) {
			return installers.StatusInstalled, "MesloLG Nerd Font"
		}
	}

	return installers.StatusNotInstalled, ""
}

// hasNerdFont checks if a directory contains Nerd Font files.
func hasNerdFont(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		name := entry.Name()
		if contains(name, "Nerd") || contains(name, "Meslo") {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}

// Install installs Nerd Fonts.
func (i *Installer) Install(opts *installers.InstallOptions) *installers.InstallResult {
	// Check if already installed
	status, version := i.CheckExisting()
	if status == installers.StatusInstalled {
		i.CLI.PrintInfo("Une Nerd Font est deja installee: " + version)
		if !i.CLI.AskYesNo("Voulez-vous installer d'autres polices?", false) {
			return installers.NewSuccessResult("Nerd Font deja installee")
		}
	}

	// Select font to install
	fontIndex := 0 // Default to MesloLG
	if !opts.NoInteraction {
		options := make([]string, len(availableFonts))
		for idx, font := range availableFonts {
			options[idx] = font.Name
		}
		fontIndex = i.CLI.AskChoice("Quelle police souhaitez-vous installer?", options, 0)
	}

	selectedFont := availableFonts[fontIndex]

	if i.SystemInfo.IsMacOS() {
		return i.installMacOS(selectedFont)
	} else if i.SystemInfo.IsDebian() {
		return i.installLinux(selectedFont)
	}

	return installers.NewFailureResult("Systeme non supporte")
}

// installMacOS installs a Nerd Font on macOS using Homebrew.
func (i *Installer) installMacOS(font FontInfo) *installers.InstallResult {
	brewPath := i.getHomebrewPath()
	if brewPath == "" {
		return installers.NewFailureResult("Homebrew est requis pour installer les polices sur macOS")
	}

	i.CLI.PrintSection("Installation de " + font.Name)

	// Add homebrew/cask-fonts tap
	i.CLI.PrintInfo("Ajout du depot homebrew/cask-fonts...")
	i.Runner.Run([]string{brewPath, "tap", "homebrew/cask-fonts"})

	// Install font
	result := i.Runner.Run(
		[]string{brewPath, "install", "--cask", font.HomebrewCask},
		runner.WithDescription("Installation de " + font.Name),
		runner.WithTimeout(5*time.Minute),
	)

	if !result.Success {
		return installers.NewFailureResult("Echec de l'installation de " + font.Name, result.Stderr)
	}

	i.CLI.PrintSuccess(font.Name + " installee")
	i.showFontInstructions(font)

	return installers.NewSuccessResult(font.Name + " installee avec succes")
}

// installLinux installs a Nerd Font on Linux.
func (i *Installer) installLinux(font FontInfo) *installers.InstallResult {
	i.CLI.PrintSection("Installation de " + font.Name)

	// Create fonts directory
	fontsDir := filepath.Join(i.SystemInfo.HomeDir, ".local/share/fonts")
	i.Runner.EnsureDirectory(fontsDir, 0755)

	// Download font from GitHub releases
	// Using MesloLG as default
	fontURL := "https://github.com/ryanoasis/nerd-fonts/releases/latest/download/Meslo.zip"
	zipPath := filepath.Join(os.TempDir(), "nerd-font.zip")

	i.CLI.PrintInfo("Telechargement de " + font.Name + "...")
	if !i.Runner.DownloadFile(fontURL, zipPath, runner.WithDescription("Telechargement")) {
		return installers.NewFailureResult("Echec du telechargement de la police")
	}

	// Extract to fonts directory
	i.CLI.PrintInfo("Extraction de la police...")
	result := i.Runner.Run(
		[]string{"unzip", "-o", zipPath, "-d", fontsDir},
		runner.WithTimeout(2*time.Minute),
	)

	if !result.Success {
		return installers.NewFailureResult("Echec de l'extraction de la police", result.Stderr)
	}

	// Update font cache
	i.CLI.PrintInfo("Mise a jour du cache des polices...")
	i.Runner.Run([]string{"fc-cache", "-fv"})

	// Cleanup
	os.Remove(zipPath)

	i.CLI.PrintSuccess(font.Name + " installee")
	i.showFontInstructions(font)

	return installers.NewSuccessResult(font.Name + " installee avec succes")
}

// showFontInstructions shows instructions for using the font.
func (i *Installer) showFontInstructions(font FontInfo) {
	i.CLI.PrintSection("Configuration")
	i.CLI.PrintInfo("Pour utiliser cette police:")
	i.CLI.Println("  1. Ouvrez les preferences de votre terminal")
	i.CLI.Println("  2. Selectionnez '" + font.Name + "' comme police")
	i.CLI.Println("  3. Redemarrez le terminal")
	i.CLI.Println("")
	i.CLI.PrintInfo("Cette police est recommandee pour le theme 'agnoster' de Oh My Zsh")
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

// Verify verifies the Nerd Font installation.
func (i *Installer) Verify() bool {
	status, _ := i.CheckExisting()
	return status == installers.StatusInstalled
}

// Uninstall removes Nerd Fonts.
func (i *Installer) Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult {
	i.CLI.PrintSection("Desinstallation des Nerd Fonts")

	if i.SystemInfo.IsMacOS() {
		brewPath := i.getHomebrewPath()
		if brewPath != "" {
			for _, font := range availableFonts {
				i.Runner.Run([]string{brewPath, "uninstall", "--cask", font.HomebrewCask})
			}
		}
	} else {
		// Linux: Remove font files
		fontsDir := filepath.Join(i.SystemInfo.HomeDir, ".local/share/fonts")
		entries, _ := os.ReadDir(fontsDir)
		for _, entry := range entries {
			name := entry.Name()
			if contains(name, "Nerd") || contains(name, "Meslo") {
				os.Remove(filepath.Join(fontsDir, name))
			}
		}
		// Update font cache
		i.Runner.Run([]string{"fc-cache", "-fv"})
	}

	i.CLI.PrintSuccess("Nerd Fonts desinstallees")
	return installers.NewUninstallSuccessResult("Nerd Fonts desinstallees avec succes")
}
