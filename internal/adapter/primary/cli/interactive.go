// Package cli provides the CLI primary adapter.
package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/keyxmare/DevBootstrap/internal/application/usecase"
	"github.com/keyxmare/DevBootstrap/internal/config"
	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// Soft minimal colors
var (
	accent        = color.New(color.FgCyan)
	textPrimary   = color.New(color.FgHiWhite)
	textSecondary = color.New(color.FgWhite)
	textMuted     = color.New(color.Faint)
	textDim       = color.New(color.FgHiBlack)
	statusSuccess = color.New(color.FgGreen)
	statusError   = color.New(color.FgRed)
	labelColor    = color.New(color.FgCyan, color.Faint)
)

const version = "3.0.0"

// App is the CLI application using hexagonal architecture.
type App struct {
	container        *config.Container
	installUseCase   *usecase.InstallApplicationUseCase
	uninstallUseCase *usecase.UninstallApplicationUseCase
	listAppsUseCase  *usecase.ListApplicationsUseCase
	reporter         secondary.ProgressReporter
	prompter         secondary.UserPrompter
}

// NewApp creates a new CLI application.
func NewApp(container *config.Container) *App {
	return &App{
		container:        container,
		installUseCase:   container.InstallUseCase,
		uninstallUseCase: container.UninstallUseCase,
		listAppsUseCase:  container.ListAppsUseCase,
		reporter:         container.ProgressReporter,
		prompter:         container.UserPrompter,
	}
}

// Run executes the main application flow.
func (a *App) Run() int {
	ctx := context.Background()

	// Print banner
	a.printBanner()

	// Print system info
	a.printSystemInfo()

	// Get all applications with status
	apps, err := a.listAppsUseCase.Execute(ctx)
	if err != nil {
		a.reporter.Error(fmt.Sprintf("Erreur lors du chargement des applications: %v", err))
		return 1
	}

	// Display available apps
	a.displayAvailableApps(apps)

	// Ask for mode if not specified
	uninstallMode := a.container.UninstallMode
	if !a.container.NoInteraction && !uninstallMode {
		uninstallMode = a.askMode()
	}

	if uninstallMode {
		return a.runUninstallMode(ctx, apps)
	}
	return a.runInstallMode(ctx, apps)
}

// printBanner prints the application banner.
func (a *App) printBanner() {
	a.reporter.Header(fmt.Sprintf("DevBootstrap v%s", version))
}

// printSystemInfo prints detected system information.
func (a *App) printSystemInfo() {
	platform := a.container.Platform
	rootStr := "Non"
	if platform.IsRoot() {
		rootStr = "Oui"
	}

	a.reporter.Summary("Système", map[string]string{
		"OS":           platform.OSName(),
		"Version":      platform.OSVersion(),
		"Architecture": platform.Arch().String(),
		"Root":         rootStr,
	})
}

// displayAvailableApps displays all available applications with their status.
func (a *App) displayAvailableApps(apps []*entity.Application) {
	a.reporter.Section("Applications")
	fmt.Println()

	// Count installed
	installedCount := 0
	for _, app := range apps {
		if app.IsInstalled() {
			installedCount++
		}
	}

	// Get padding from reporter if available
	pad := "          " // default padding

	for _, app := range apps {
		fmt.Print(pad)

		if app.IsInstalled() {
			statusSuccess.Print("  ●  ")
			textPrimary.Print(app.Name())

			// Version
			if !app.Version().IsEmpty() {
				ver := cleanVersion(app.Version().String())
				if ver != "" {
					textMuted.Printf("  %s", ver)
				}
			}
		} else {
			textDim.Print("  ○  ")
			textSecondary.Print(app.Name())
		}
		fmt.Println()

		// Description
		fmt.Print(pad)
		textMuted.Printf("       %s\n", app.Description())
		fmt.Println()
	}

	// Summary line
	fmt.Print(pad)
	textMuted.Printf("  %d/%d installées\n", installedCount, len(apps))
}

// cleanVersion cleans and shortens version strings.
func cleanVersion(version string) string {
	// Remove common prefixes
	version = strings.TrimPrefix(version, "Docker version ")
	version = strings.TrimPrefix(version, "NVIM ")
	version = strings.TrimPrefix(version, "zsh ")

	// Handle VSCode special case
	if strings.Contains(version, "commande 'code' non") {
		return ""
	}

	// Simplify Nerd Font names
	if strings.Contains(version, "Nerd Font") {
		// Extract just the font family name
		parts := strings.Split(version, " ")
		if len(parts) > 0 {
			return parts[0]
		}
	}

	// Remove build info after comma
	if idx := strings.Index(version, ","); idx > 0 {
		version = version[:idx]
	}

	// Remove git hash
	if idx := strings.Index(version, "+g"); idx > 0 {
		version = version[:idx]
	}

	// Remove -dev suffix
	if idx := strings.Index(version, "-dev"); idx > 0 {
		version = version[:idx]
	}

	// Remove architecture in parentheses
	if idx := strings.Index(version, " ("); idx > 0 {
		version = version[:idx]
	}

	// Truncate if too long
	if len(version) > 20 {
		return version[:17] + "..."
	}

	return version
}

// askMode asks the user to select install or uninstall mode.
func (a *App) askMode() bool {
	a.reporter.Section("Action")
	fmt.Println()

	pad := "          "

	fmt.Print(pad)
	accent.Print("  [1]  ")
	textPrimary.Print("Installer")
	textMuted.Println("     Ajouter des applications")

	fmt.Print(pad)
	accent.Print("  [2]  ")
	textPrimary.Print("Désinstaller")
	textMuted.Println("  Supprimer des applications")

	fmt.Println()

	choice := a.prompter.Select("Choix", []string{"Installer", "Désinstaller"}, 0)
	return choice == 1
}

// runInstallMode runs the installation mode.
func (a *App) runInstallMode(ctx context.Context, apps []*entity.Application) int {
	// Get not-installed apps
	notInstalled := make([]*entity.Application, 0)
	for _, app := range apps {
		if !app.IsInstalled() {
			notInstalled = append(notInstalled, app)
		}
	}

	if len(notInstalled) == 0 {
		a.reporter.Success("Toutes les applications sont déjà installées")
		if !a.prompter.Confirm("Réinstaller certaines applications?", false) {
			return 0
		}
	}

	// Ask user to select apps
	selectedApps := a.askMultiSelectInstall(apps)
	if len(selectedApps) == 0 {
		a.reporter.Info("Aucune sélection")
		return 0
	}

	// Show summary and confirm
	if !a.showInstallSummary(selectedApps) {
		a.reporter.Info("Annulé")
		return 0
	}

	// Run installations
	appIDs := make([]valueobject.AppID, len(selectedApps))
	for i, app := range selectedApps {
		appIDs[i] = app.ID()
	}

	opts := primary.InstallOptions{
		DryRun:        a.container.DryRun,
		NoInteraction: a.container.NoInteraction,
	}

	results, err := a.installUseCase.ExecuteMultiple(ctx, appIDs, opts)
	if err != nil {
		a.reporter.Error(fmt.Sprintf("Erreur: %v", err))
		return 1
	}

	// Show final summary
	return a.showFinalSummary(results, selectedApps)
}

// runUninstallMode runs the uninstallation mode.
func (a *App) runUninstallMode(ctx context.Context, apps []*entity.Application) int {
	// Get installed apps
	installed := make([]*entity.Application, 0)
	for _, app := range apps {
		if app.IsInstalled() {
			installed = append(installed, app)
		}
	}

	if len(installed) == 0 {
		a.reporter.Warning("Aucune application installée")
		return 0
	}

	// Ask user to select apps
	selectedApps := a.askMultiSelectUninstall(installed)
	if len(selectedApps) == 0 {
		a.reporter.Info("Aucune sélection")
		return 0
	}

	// Show summary and confirm
	if !a.showUninstallSummary(selectedApps) {
		a.reporter.Info("Annulé")
		return 0
	}

	// Run uninstallations
	appIDs := make([]valueobject.AppID, len(selectedApps))
	for i, app := range selectedApps {
		appIDs[i] = app.ID()
	}

	opts := primary.UninstallOptions{
		DryRun:        a.container.DryRun,
		NoInteraction: a.container.NoInteraction,
		RemoveConfig:  true,
		RemoveCache:   true,
		RemoveData:    true,
	}

	results, err := a.uninstallUseCase.ExecuteMultiple(ctx, appIDs, opts)
	if err != nil {
		a.reporter.Error(fmt.Sprintf("Erreur: %v", err))
		return 1
	}

	// Show final summary
	return a.showUninstallFinalSummary(results, selectedApps)
}

// askMultiSelectInstall asks the user to select applications to install.
func (a *App) askMultiSelectInstall(apps []*entity.Application) []*entity.Application {
	if a.container.NoInteraction {
		result := make([]*entity.Application, 0)
		for _, app := range apps {
			if !app.IsInstalled() {
				result = append(result, app)
			}
		}
		a.reporter.Info(fmt.Sprintf("Mode auto: %d application(s)", len(result)))
		return result
	}

	a.reporter.Section("Sélection")
	fmt.Println()

	options := make([]string, 0, len(apps))
	defaultIndices := make([]int, 0)

	for i, app := range apps {
		status := ""
		if app.IsInstalled() {
			status = " (installé)"
		} else {
			defaultIndices = append(defaultIndices, i)
		}
		options = append(options, fmt.Sprintf("%s%s", app.Name(), status))
	}

	selected := a.prompter.MultiSelect("Applications à installer", options, defaultIndices)

	result := make([]*entity.Application, 0, len(selected))
	for _, idx := range selected {
		if idx < len(apps) {
			result = append(result, apps[idx])
		}
	}

	return result
}

// askMultiSelectUninstall asks the user to select applications to uninstall.
func (a *App) askMultiSelectUninstall(installed []*entity.Application) []*entity.Application {
	if a.container.NoInteraction {
		a.reporter.Warning("Désinstallation auto non supportée")
		return nil
	}

	a.reporter.Section("Sélection")
	fmt.Println()

	options := make([]string, 0, len(installed))
	for _, app := range installed {
		options = append(options, app.Name())
	}

	selected := a.prompter.MultiSelect("Applications à désinstaller", options, []int{})

	result := make([]*entity.Application, 0, len(selected))
	for _, idx := range selected {
		if idx < len(installed) {
			result = append(result, installed[idx])
		}
	}

	return result
}

// showInstallSummary shows the installation summary and asks for confirmation.
func (a *App) showInstallSummary(apps []*entity.Application) bool {
	a.reporter.Section("Récapitulatif")
	fmt.Println()

	pad := "          "

	fmt.Print(pad)
	textMuted.Println("  Applications sélectionnées:")
	fmt.Println()

	for _, app := range apps {
		fmt.Print(pad)
		accent.Print("    →  ")
		textPrimary.Println(app.Name())
	}

	fmt.Println()
	return a.prompter.Confirm("Continuer?", true)
}

// showUninstallSummary shows the uninstallation summary and asks for confirmation.
func (a *App) showUninstallSummary(apps []*entity.Application) bool {
	a.reporter.Section("Récapitulatif")
	fmt.Println()

	pad := "          "

	fmt.Print(pad)
	textMuted.Println("  Applications à supprimer:")
	fmt.Println()

	for _, app := range apps {
		fmt.Print(pad)
		statusError.Print("    ✕  ")
		textPrimary.Println(app.Name())
	}

	fmt.Println()
	a.reporter.Warning("Les configurations seront supprimées")
	fmt.Println()

	return a.prompter.Confirm("Confirmer la suppression?", false)
}

// showFinalSummary shows the final summary of install operations.
func (a *App) showFinalSummary(results map[string]*result.InstallResult, apps []*entity.Application) int {
	fmt.Println()
	a.reporter.Section("Résultat")
	fmt.Println()

	pad := "          "
	successCount := 0

	for _, app := range apps {
		r, ok := results[app.ID().String()]
		if !ok {
			continue
		}

		fmt.Print(pad)
		if r.Success() {
			successCount++
			statusSuccess.Print("  ✓  ")
			textPrimary.Println(app.Name())
		} else {
			statusError.Print("  ✕  ")
			textSecondary.Println(app.Name())
		}
	}

	fmt.Println()
	failureCount := len(apps) - successCount
	if failureCount == 0 {
		a.reporter.Success(fmt.Sprintf("Terminé avec succès (%d/%d)", successCount, len(apps)))
		return 0
	}

	a.reporter.Warning(fmt.Sprintf("Terminé: %d succès, %d échec(s)", successCount, failureCount))
	return 1
}

// showUninstallFinalSummary shows the final summary of uninstall operations.
func (a *App) showUninstallFinalSummary(results map[string]*result.UninstallResult, apps []*entity.Application) int {
	fmt.Println()
	a.reporter.Section("Résultat")
	fmt.Println()

	pad := "          "
	successCount := 0

	for _, app := range apps {
		r, ok := results[app.ID().String()]
		if !ok {
			continue
		}

		fmt.Print(pad)
		if r.Success() {
			successCount++
			statusSuccess.Print("  ✓  ")
			textPrimary.Println(app.Name())
		} else {
			statusError.Print("  ✕  ")
			textSecondary.Println(app.Name())
		}
	}

	fmt.Println()
	failureCount := len(apps) - successCount
	if failureCount == 0 {
		a.reporter.Success(fmt.Sprintf("Terminé avec succès (%d/%d)", successCount, len(apps)))
		return 0
	}

	a.reporter.Warning(fmt.Sprintf("Terminé: %d succès, %d échec(s)", successCount, failureCount))
	return 1
}
