package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
	"golang.org/x/term"
)

// App is the main TUI application
type App struct {
	model   *Model
	runner  *runner.Runner
	sysInfo *system.SystemInfo
	dryRun  bool

	// Selected items after TUI exits
	selectedItems []AppItem
	confirmed     bool
}

// NewApp creates a new TUI application
func NewApp(sysInfo *system.SystemInfo, r *runner.Runner, uninstall, noInteract, dryRun bool) *App {
	model := NewModel(sysInfo, uninstall, noInteract, dryRun)

	return &App{
		model:   model,
		runner:  r,
		sysInfo: sysInfo,
		dryRun:  dryRun,
	}
}

// AddApp adds an application to the list
func (a *App) AddApp(id, name, description string, tags []installers.AppTag, status installers.AppStatus, version string, inst installers.Installer, uninst installers.Uninstaller) {
	item := AppItem{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		Status:      status,
		Version:     version,
		Selected:    false,
		Installer:   inst,
		Uninstaller: uninst,
	}
	a.model.items = append(a.model.items, item)
}

// Run starts the TUI application
func (a *App) Run() error {
	// Set callback to capture selection
	a.model.SetConfirmCallback(func(items []AppItem) {
		a.selectedItems = items
		a.confirmed = true
	})

	// Run the TUI for selection only
	p := tea.NewProgram(a.model, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		return err
	}

	// If user confirmed selection, run installations outside TUI
	if a.confirmed && len(a.selectedItems) > 0 {
		return a.runInstallations()
	}

	return nil
}

// sudoAskpassScript stores the path to the temporary askpass script
var sudoAskpassScript string

// preCacheSudo asks for sudo password upfront and sets up SUDO_ASKPASS
func (a *App) preCacheSudo() {
	// Skip if already root or in dry-run mode
	if a.sysInfo.IsRoot || a.dryRun {
		return
	}

	// Check if sudo is available
	if _, err := exec.LookPath("sudo"); err != nil {
		return
	}

	// Always ask for password to set up SUDO_ASKPASS
	// (the TTY-based sudo cache doesn't work reliably across subprocesses)
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4"))
	fmt.Println()
	fmt.Println(infoStyle.Render("🔐 Certaines installations nécessitent les droits administrateur."))
	fmt.Print("Mot de passe: ")

	// Read password without echo
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input

	if err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
		fmt.Println(errorStyle.Render("⚠ Impossible de lire le mot de passe."))
		fmt.Println()
		return
	}

	password := strings.TrimSpace(string(passwordBytes))
	if password == "" {
		fmt.Println()
		return
	}

	// Verify password is correct
	verifyCmd := exec.Command("sudo", "-S", "-v")
	verifyCmd.Stdin = strings.NewReader(password + "\n")
	if err := verifyCmd.Run(); err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
		fmt.Println(errorStyle.Render("⚠ Mot de passe incorrect."))
		fmt.Println()
		return
	}

	// Create temporary askpass script
	tmpFile, err := os.CreateTemp("", "askpass-*.sh")
	if err != nil {
		return
	}

	// Write script that echoes the password
	// Using printf to avoid issues with special characters
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' '%s'\n", escapeForShell(password))
	tmpFile.WriteString(script)
	tmpFile.Close()
	os.Chmod(tmpFile.Name(), 0700)

	sudoAskpassScript = tmpFile.Name()

	// Set environment variable for the runner to use
	a.runner.SetSudoAskpass(sudoAskpassScript)

	fmt.Println()
}

// escapeForShell escapes a string for safe use in single quotes
func escapeForShell(s string) string {
	// In single quotes, we only need to escape single quotes
	// We do this by ending the quote, adding escaped quote, starting new quote
	return strings.ReplaceAll(s, "'", "'\"'\"'")
}

// cleanupSudo removes the temporary askpass script
func (a *App) cleanupSudo() {
	if sudoAskpassScript != "" {
		os.Remove(sudoAskpassScript)
		sudoAskpassScript = ""
	}
	a.runner.SetSudoAskpass("")
}

func (a *App) runInstallations() error {
	// Ask for sudo password now, right before installations
	a.preCacheSudo()

	// Styles for output
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4"))
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)

	action := "Installation"
	if a.model.uninstall {
		action = "Désinstallation"
	}

	fmt.Println(headerStyle.Render(fmt.Sprintf("━━━ %s ━━━", action)))
	fmt.Println()

	opts := &installers.InstallOptions{
		DryRun:        a.dryRun,
		NoInteraction: true,
	}

	uninstallOpts := &installers.UninstallOptions{
		DryRun:        a.dryRun,
		NoInteraction: true,
		RemoveConfig:  true,
		RemoveCache:   true,
		RemoveData:    true,
	}

	var hasError bool

	for _, item := range a.selectedItems {
		fmt.Println(infoStyle.Render(fmt.Sprintf("→ %s %s...", action, item.Name)))

		if a.model.uninstall {
			if item.Uninstaller == nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("  ✗ Pas de désinstallateur pour %s", item.Name)))
				continue
			}
			result := item.Uninstaller.Uninstall(uninstallOpts)
			if result.Success {
				fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s désinstallé", item.Name)))
			} else {
				fmt.Println(errorStyle.Render(fmt.Sprintf("  ✗ Échec: %s", result.Message)))
				hasError = true
			}
		} else {
			if item.Installer == nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("  ✗ Pas d'installateur pour %s", item.Name)))
				continue
			}
			result := item.Installer.Install(opts)
			if result.Success {
				fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s installé", item.Name)))
			} else {
				fmt.Println(errorStyle.Render(fmt.Sprintf("  ✗ Échec: %s", result.Message)))
				hasError = true
			}
		}
		fmt.Println()
	}

	// Cleanup sudo askpass script
	a.cleanupSudo()

	fmt.Println(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━"))
	if hasError {
		fmt.Println(errorStyle.Render("Certaines installations ont échoué."))
		os.Exit(1)
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s terminée avec succès!", action)))
	}
	fmt.Println()

	return nil
}
