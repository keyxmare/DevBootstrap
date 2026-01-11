// Package cli provides the CLI primary adapter.
package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/keyxmare/DevBootstrap/internal/application/usecase"
	"github.com/keyxmare/DevBootstrap/internal/config"
	"github.com/keyxmare/DevBootstrap/internal/domain/entity"
	"github.com/keyxmare/DevBootstrap/internal/domain/result"
	"github.com/keyxmare/DevBootstrap/internal/domain/valueobject"
	"github.com/keyxmare/DevBootstrap/internal/port/primary"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
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
	a.reporter.Header(fmt.Sprintf("DevBootstrap v%s (Clean Architecture)", version))
}

// printSystemInfo prints detected system information.
func (a *App) printSystemInfo() {
	platform := a.container.Platform
	rootStr := "Non"
	if platform.IsRoot() {
		rootStr = "Oui"
	}

	a.reporter.Summary("Informations systeme", map[string]string{
		"OS":           platform.OSName(),
		"Version":      platform.OSVersion(),
		"Architecture": platform.Arch().String(),
		"Home":         platform.HomeDir(),
		"Root":         rootStr,
	})
}

// displayAvailableApps displays all available applications with their status.
func (a *App) displayAvailableApps(apps []*entity.Application) {
	a.reporter.Section("Applications disponibles")
	fmt.Println()

	for _, app := range apps {
		statusIcon := "○"
		statusText := "non installe"

		if app.IsInstalled() {
			statusIcon = "✓"
			statusText = "installe"
			if !app.Version().IsEmpty() {
				statusText += fmt.Sprintf(" (%s)", app.Version())
			}
		}

		fmt.Printf("  %s %s\n", statusIcon, app.Name())
		fmt.Printf("      %s\n", app.Description())
		fmt.Printf("      Status: %s\n", statusText)
		fmt.Println()
	}
}

// askMode asks the user to select install or uninstall mode.
func (a *App) askMode() bool {
	a.reporter.Section("Mode d'operation")
	fmt.Println()
	fmt.Printf("  [1] Installer - Installer de nouvelles applications\n")
	fmt.Printf("  [2] Desinstaller - Supprimer des applications installees\n")
	fmt.Println()

	choice := a.prompter.Select("Votre choix", []string{"Installer", "Desinstaller"}, 0)
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
		a.reporter.Success("Toutes les applications sont deja installees!")
		if !a.prompter.Confirm("Voulez-vous quand meme reinstaller certaines applications?", false) {
			return 0
		}
	}

	// Ask user to select apps
	selectedApps := a.askMultiSelectInstall(apps)
	if len(selectedApps) == 0 {
		a.reporter.Info("Aucune application selectionnee")
		return 0
	}

	// Show summary and confirm
	if !a.showInstallSummary(selectedApps) {
		a.reporter.Info("Installation annulee")
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
		a.reporter.Warning("Aucune application installee a desinstaller")
		return 0
	}

	// Ask user to select apps
	selectedApps := a.askMultiSelectUninstall(installed)
	if len(selectedApps) == 0 {
		a.reporter.Info("Aucune application selectionnee")
		return 0
	}

	// Show summary and confirm
	if !a.showUninstallSummary(selectedApps) {
		a.reporter.Info("Desinstallation annulee")
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
		a.reporter.Section("Mode non-interactif")
		a.reporter.Info(fmt.Sprintf("Installation automatique de %d application(s)", len(result)))
		return result
	}

	a.reporter.Section("Selection des installations")
	fmt.Println()

	options := make([]string, 0, len(apps))
	defaultIndices := make([]int, 0)

	for i, app := range apps {
		status := ""
		if app.IsInstalled() {
			status = " (deja installe)"
		} else {
			defaultIndices = append(defaultIndices, i)
		}
		options = append(options, fmt.Sprintf("%s%s", app.Name(), status))
	}

	selected := a.prompter.MultiSelect("Quelles applications souhaitez-vous installer?", options, defaultIndices)

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
		a.reporter.Warning("Desinstallation automatique non supportee en mode non-interactif")
		return nil
	}

	a.reporter.Section("Selection des desinstallations")
	fmt.Println()

	options := make([]string, 0, len(installed))
	for _, app := range installed {
		options = append(options, app.Name())
	}

	selected := a.prompter.MultiSelect("Quelles applications souhaitez-vous desinstaller?", options, []int{})

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
	a.reporter.Section("Resume de l'installation")
	fmt.Println()
	fmt.Println("Applications a installer:")
	for _, app := range apps {
		fmt.Printf("  • %s - %s\n", app.Name(), app.Description())
	}
	fmt.Println()

	return a.prompter.Confirm("Proceder a l'installation?", true)
}

// showUninstallSummary shows the uninstallation summary and asks for confirmation.
func (a *App) showUninstallSummary(apps []*entity.Application) bool {
	a.reporter.Section("Resume de la desinstallation")
	fmt.Println()
	fmt.Println("Applications a desinstaller:")
	for _, app := range apps {
		fmt.Printf("  • %s\n", app.Name())
	}
	fmt.Println()
	a.reporter.Warning("Cette action supprimera les applications et leurs configurations!")

	return a.prompter.Confirm("Etes-vous sur de vouloir continuer?", false)
}

// showFinalSummary shows the final summary of install operations.
func (a *App) showFinalSummary(results map[string]*result.InstallResult, apps []*entity.Application) int {
	fmt.Println()
	a.reporter.Section("Resume final")
	fmt.Println()

	successCount := 0
	for _, app := range apps {
		r, ok := results[app.ID().String()]
		if !ok {
			continue
		}

		if r.Success() {
			successCount++
			a.reporter.Success(fmt.Sprintf("%s installe avec succes", app.Name()))
		} else {
			a.reporter.Error(fmt.Sprintf("%s - echec", app.Name()))
		}
	}

	fmt.Println()
	failureCount := len(apps) - successCount
	if failureCount == 0 {
		a.reporter.Success(fmt.Sprintf("Toutes les operations terminees avec succes! (%d/%d)", successCount, len(apps)))
		return 0
	}

	a.reporter.Warning(fmt.Sprintf("Operations terminees: %d succes, %d echec(s)", successCount, failureCount))
	return 1
}

// showUninstallFinalSummary shows the final summary of uninstall operations.
func (a *App) showUninstallFinalSummary(results map[string]*result.UninstallResult, apps []*entity.Application) int {
	fmt.Println()
	a.reporter.Section("Resume final")
	fmt.Println()

	successCount := 0
	for _, app := range apps {
		r, ok := results[app.ID().String()]
		if !ok {
			continue
		}

		if r.Success() {
			successCount++
			a.reporter.Success(fmt.Sprintf("%s desinstalle avec succes", app.Name()))
		} else {
			a.reporter.Error(fmt.Sprintf("%s - echec", app.Name()))
		}
	}

	fmt.Println()
	failureCount := len(apps) - successCount
	if failureCount == 0 {
		a.reporter.Success(fmt.Sprintf("Toutes les operations terminees avec succes! (%d/%d)", successCount, len(apps)))
		return 0
	}

	a.reporter.Warning(fmt.Sprintf("Operations terminees: %d succes, %d echec(s)", successCount, failureCount))
	return 1
}

// formatTags formats tags for display.
func formatTags(tags []valueobject.AppTag) string {
	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		parts = append(parts, fmt.Sprintf("[%s]", tag))
	}
	return strings.Join(parts, " ")
}
