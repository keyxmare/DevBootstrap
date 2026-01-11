package tui

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
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

// sudoKeepAliveStop is used to stop the sudo keep-alive goroutine
var sudoKeepAliveStop chan struct{}

// preCacheSudo asks for sudo password upfront and starts keep-alive
func (a *App) preCacheSudo() {
	// Skip if already root or in dry-run mode
	if a.sysInfo.IsRoot || a.dryRun {
		return
	}

	// Check if sudo is available
	if _, err := exec.LookPath("sudo"); err != nil {
		return
	}

	// Check if sudo credentials are already cached
	checkCmd := exec.Command("sudo", "-n", "true")
	if checkCmd.Run() == nil {
		// Already authenticated, start keep-alive
		a.startSudoKeepAlive()
		return
	}

	// Ask for password
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4"))
	fmt.Println()
	fmt.Println(infoStyle.Render("🔐 Certaines installations nécessitent les droits administrateur."))
	fmt.Println()

	// Run sudo -v to cache credentials (will prompt for password)
	sudoCmd := exec.Command("sudo", "-v")
	sudoCmd.Stdin = os.Stdin
	sudoCmd.Stdout = os.Stdout
	sudoCmd.Stderr = os.Stderr

	if err := sudoCmd.Run(); err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
		fmt.Println(errorStyle.Render("⚠ Impossible d'obtenir les droits sudo. Certaines installations peuvent échouer."))
	} else {
		// Start keep-alive to maintain sudo credentials during installations
		a.startSudoKeepAlive()
	}
	fmt.Println()
}

// startSudoKeepAlive starts a goroutine that refreshes sudo every 30 seconds
func (a *App) startSudoKeepAlive() {
	sudoKeepAliveStop = make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Refresh sudo timestamp silently
				exec.Command("sudo", "-v").Run()
			case <-sudoKeepAliveStop:
				return
			}
		}
	}()
}

// stopSudoKeepAlive stops the sudo keep-alive goroutine
func (a *App) stopSudoKeepAlive() {
	if sudoKeepAliveStop != nil {
		close(sudoKeepAliveStop)
		sudoKeepAliveStop = nil
	}
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

	// Stop sudo keep-alive
	a.stopSudoKeepAlive()

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
