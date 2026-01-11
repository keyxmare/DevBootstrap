// Package alias provides devbootstrap command installation functionality.
package alias

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/keyxmare/DevBootstrap/internal/cli"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

const (
	commandName = "devbootstrap"
	githubRepo  = "keyxmare/DevBootstrap"
)

// Installer handles devbootstrap command installation.
type Installer struct {
	*installers.BaseInstaller
}

// NewInstaller creates a new alias installer.
func NewInstaller(c *cli.CLI, r *runner.Runner, sysInfo *system.SystemInfo) *Installer {
	return &Installer{
		BaseInstaller: installers.NewBaseInstaller(c, r, sysInfo),
	}
}

// Name returns the application name.
func (i *Installer) Name() string {
	return "Commande devbootstrap"
}

// ID returns the application ID.
func (i *Installer) ID() string {
	return "alias"
}

// Description returns the application description.
func (i *Installer) Description() string {
	return "Installe la commande 'devbootstrap'"
}

// Tags returns the application tags.
func (i *Installer) Tags() []installers.AppTag {
	return []installers.AppTag{installers.TagAlias}
}

// CheckExisting checks if the devbootstrap command is already installed.
func (i *Installer) CheckExisting() (installers.AppStatus, string) {
	// Check in PATH
	if i.Runner.CommandExists(commandName) {
		return installers.StatusInstalled, ""
	}

	// Check in ~/.local/bin
	scriptPath := filepath.Join(i.SystemInfo.HomeDir, ".local/bin", commandName)
	if _, err := os.Stat(scriptPath); err == nil {
		return installers.StatusInstalled, ""
	}

	return installers.StatusNotInstalled, ""
}

// Install installs the devbootstrap command.
func (i *Installer) Install(opts *installers.InstallOptions) *installers.InstallResult {
	// Check if already installed
	status, _ := i.CheckExisting()
	if status == installers.StatusInstalled {
		i.CLI.PrintInfo("La commande devbootstrap est deja installee")
		if !i.CLI.AskYesNo("Voulez-vous reinstaller?", false) {
			return installers.NewSuccessResult("Commande deja installee")
		}
	}

	i.CLI.PrintSection("Installation de la commande devbootstrap")

	// Create ~/.local/bin directory
	binDir := filepath.Join(i.SystemInfo.HomeDir, ".local/bin")
	if err := i.Runner.EnsureDirectory(binDir, 0755); err != nil {
		return installers.NewFailureResult("Impossible de creer le repertoire", err.Error())
	}

	// Create the script
	scriptPath := filepath.Join(binDir, commandName)
	scriptContent := i.generateScript()

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return installers.NewFailureResult("Impossible de creer le script", err.Error())
	}
	i.CLI.PrintSuccess("Script cree dans ~/.local/bin/" + commandName)

	// Ensure ~/.local/bin is in PATH
	i.ensurePathConfigured()

	i.showFinalInstructions()

	return installers.NewSuccessResult("Commande devbootstrap installee avec succes")
}

// generateScript generates the devbootstrap shell script.
func (i *Installer) generateScript() string {
	return fmt.Sprintf(`#!/bin/bash
# DevBootstrap - Installation automatique de l'environnement de developpement
# https://github.com/%s

# Refuse to run as root/sudo on macOS (causes issues with Homebrew)
if [[ "$OSTYPE" == "darwin"* ]] && [[ $EUID -eq 0 ]]; then
    echo "Erreur: Ne pas executer devbootstrap avec sudo sur macOS."
    echo "Homebrew ne fonctionne pas correctement en root."
    echo ""
    echo "Utilisez simplement: devbootstrap"
    exit 1
fi

INSTALL_DIR="${HOME}/.devbootstrap"
REPO_URL="https://github.com/%s"

# Sync repository silently (ignore errors, continue with local version)
sync_repo() {
    if command -v git &> /dev/null; then
        if [ -d "$INSTALL_DIR/.git" ]; then
            # Update existing repo
            cd "$INSTALL_DIR" && git fetch origin main --quiet 2>/dev/null && git reset --hard origin/main --quiet 2>/dev/null
        else
            # Clone if not exists
            rm -rf "$INSTALL_DIR" 2>/dev/null
            git clone --depth=1 --quiet "$REPO_URL" "$INSTALL_DIR" 2>/dev/null
        fi
    fi
}

# Sync in background (silent)
sync_repo &>/dev/null

# Wait for sync to complete (max 5 seconds)
wait

# Check if we have a local installation with Go binary
if [ -f "$INSTALL_DIR/devbootstrap" ]; then
    "$INSTALL_DIR/devbootstrap" "$@"
    exit $?
fi

# Fallback: check for Python version
if [ -d "$INSTALL_DIR" ] && [ -f "$INSTALL_DIR/bootstrap/__main__.py" ]; then
    cd "$INSTALL_DIR"
    python3 -m bootstrap "$@"
    exit $?
fi

# Not installed
echo "DevBootstrap n'est pas installe. Telechargement..."
if command -v git &> /dev/null; then
    git clone --depth=1 "$REPO_URL" "$INSTALL_DIR"
    if [ -f "$INSTALL_DIR/devbootstrap" ]; then
        "$INSTALL_DIR/devbootstrap" "$@"
    elif [ -f "$INSTALL_DIR/bootstrap/__main__.py" ]; then
        cd "$INSTALL_DIR" && python3 -m bootstrap "$@"
    fi
else
    echo "Erreur: git est requis pour installer DevBootstrap"
    exit 1
fi
`, githubRepo, githubRepo)
}

// ensurePathConfigured ensures ~/.local/bin is in PATH.
func (i *Installer) ensurePathConfigured() {
	pathLine := "\n# Add ~/.local/bin to PATH\nexport PATH=\"$HOME/.local/bin:$PATH\"\n"

	rcFiles := i.getShellRCFiles()
	for _, rcFile := range rcFiles {
		// Check if already configured
		if content, err := os.ReadFile(rcFile); err == nil {
			if contains(string(content), ".local/bin") {
				continue
			}
		}

		// Add to RC file
		f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			continue
		}
		f.WriteString(pathLine)
		f.Close()
		i.CLI.PrintSuccess("PATH configure dans " + rcFile)
	}
}

// getShellRCFiles returns the list of shell RC files to update.
func (i *Installer) getShellRCFiles() []string {
	var rcFiles []string

	zshrc := filepath.Join(i.SystemInfo.HomeDir, ".zshrc")
	bashrc := filepath.Join(i.SystemInfo.HomeDir, ".bashrc")
	bashProfile := filepath.Join(i.SystemInfo.HomeDir, ".bash_profile")

	if _, err := os.Stat(zshrc); err == nil {
		rcFiles = append(rcFiles, zshrc)
	}

	if _, err := os.Stat(bashrc); err == nil {
		rcFiles = append(rcFiles, bashrc)
	} else if _, err := os.Stat(bashProfile); err == nil {
		rcFiles = append(rcFiles, bashProfile)
	}

	// If no RC file exists, create based on current shell
	if len(rcFiles) == 0 {
		currentShell := os.Getenv("SHELL")
		if contains(currentShell, "zsh") {
			rcFiles = append(rcFiles, zshrc)
		} else {
			rcFiles = append(rcFiles, bashrc)
		}
	}

	return rcFiles
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}

// showFinalInstructions shows final instructions.
func (i *Installer) showFinalInstructions() {
	i.CLI.PrintSection("Prochaines etapes")
	i.CLI.Println("")
	i.CLI.Println("1. Redemarrer le terminal ou executer:")
	i.CLI.Println("   $ source ~/.bashrc  # ou ~/.zshrc")
	i.CLI.Println("")
	i.CLI.Println("2. Utiliser la commande:")
	i.CLI.Println("   $ devbootstrap")
	i.CLI.Println("")
	i.CLI.PrintInfo("Comportement:")
	i.CLI.Println("   - Synchronise automatiquement avec GitHub (silencieux)")
	i.CLI.Println("   - En cas d'echec, utilise la version locale")
	i.CLI.Println("   - Lance le menu DevBootstrap")
}

// Verify verifies the installation.
func (i *Installer) Verify() bool {
	status, _ := i.CheckExisting()
	return status == installers.StatusInstalled
}

// Uninstall removes the devbootstrap command.
func (i *Installer) Uninstall(opts *installers.UninstallOptions) *installers.UninstallResult {
	i.CLI.PrintSection("Desinstallation de la commande devbootstrap")

	// Remove script
	scriptPath := filepath.Join(i.SystemInfo.HomeDir, ".local/bin", commandName)
	os.Remove(scriptPath)

	// Remove installation directory
	installDir := filepath.Join(i.SystemInfo.HomeDir, ".devbootstrap")
	i.Runner.RemoveAll(installDir)

	i.CLI.PrintSuccess("Commande devbootstrap desinstallee")
	return installers.NewUninstallSuccessResult("Commande devbootstrap desinstallee avec succes")
}
