package app

import (
	"fmt"
	"strings"

	"github.com/keyxmare/DevBootstrap/internal/cli"
	"github.com/keyxmare/DevBootstrap/internal/installers"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
)

const version = "2.0.0"

// App is the main DevBootstrap application.
type App struct {
	CLI           *cli.CLI
	Runner        *runner.Runner
	SystemInfo    *system.SystemInfo
	Registry      *Registry
	DryRun        bool
	NoInteraction bool
	UninstallMode bool
}

// New creates a new App instance.
func New(dryRun, noInteraction, uninstallMode bool) *App {
	c := cli.New(noInteraction, dryRun)
	r := runner.New(c, dryRun)
	sysInfo := system.Detect()

	return &App{
		CLI:           c,
		Runner:        r,
		SystemInfo:    sysInfo,
		Registry:      NewRegistry(c, r, sysInfo),
		DryRun:        dryRun,
		NoInteraction: noInteraction,
		UninstallMode: uninstallMode,
	}
}

// Run executes the main application flow.
func (a *App) Run() int {
	// Print banner
	a.printBanner()

	// Print system info
	a.printSystemInfo()

	// Display available apps
	appsStatus := a.displayAvailableApps()

	// Ask for mode if not specified
	if !a.NoInteraction && !a.UninstallMode {
		a.UninstallMode = a.askMode()
	}

	if a.UninstallMode {
		return a.runUninstallMode(appsStatus)
	}
	return a.runInstallMode(appsStatus)
}

// printBanner prints the application banner.
func (a *App) printBanner() {
	a.CLI.PrintHeader(fmt.Sprintf("DevBootstrap v%s", version))
	cli.Dim.Println("Suite d'installation automatique pour votre environnement de developpement")
	fmt.Println()
}

// printSystemInfo prints detected system information.
func (a *App) printSystemInfo() {
	a.CLI.PrintSummary("Informations systeme", map[string]string{
		"OS":           a.SystemInfo.OSName,
		"Version":      a.SystemInfo.OSVersion,
		"Architecture": a.SystemInfo.Arch.String(),
		"Home":         a.SystemInfo.HomeDir,
		"Root":         boolToFrench(a.SystemInfo.IsRoot),
	})
}

// displayAvailableApps displays all available applications with their status.
func (a *App) displayAvailableApps() []appStatusEntry {
	a.CLI.PrintSection("Applications disponibles")
	fmt.Println()

	entries := make([]appStatusEntry, 0, len(a.Registry.Apps))

	for _, app := range a.Registry.Apps {
		status, version := app.Installer.CheckExisting()
		entries = append(entries, appStatusEntry{
			App:     app,
			Status:  status,
			Version: version,
		})

		// Status indicator
		var statusIcon, statusText string
		if status == installers.StatusInstalled {
			statusIcon = cli.Green.Sprint(cli.IconSuccess)
			statusText = cli.Green.Sprint("installe")
			if version != "" {
				statusText += cli.Dim.Sprintf(" (%s)", version)
			}
		} else {
			statusIcon = cli.Yellow.Sprint(cli.IconPending)
			statusText = cli.Yellow.Sprint("non installe")
		}

		// Format tags
		tagsStr := formatTags(app.Tags)

		fmt.Printf("  %s %s %s\n", statusIcon, cli.Bold.Sprint(app.Name), tagsStr)
		fmt.Printf("      %s\n", cli.Dim.Sprint(app.Description))
		fmt.Printf("      Status: %s\n", statusText)
		fmt.Println()
	}

	return entries
}

// askMode asks the user to select install or uninstall mode.
func (a *App) askMode() bool {
	a.CLI.PrintSection("Mode d'operation")
	fmt.Println()
	fmt.Printf("  [1] %s - Installer de nouvelles applications\n", cli.Green.Sprint("Installer"))
	fmt.Printf("  [2] %s - Supprimer des applications installees\n", cli.Red.Sprint("Desinstaller"))
	fmt.Println()

	choice := a.CLI.AskChoice("Votre choix", []string{"Installer", "Desinstaller"}, 0)
	return choice == 1
}

// runInstallMode runs the installation mode.
func (a *App) runInstallMode(appsStatus []appStatusEntry) int {
	// Check if all apps are installed
	notInstalled := make([]*AppEntry, 0)
	for _, entry := range appsStatus {
		if entry.Status == installers.StatusNotInstalled {
			notInstalled = append(notInstalled, entry.App)
		}
	}

	if len(notInstalled) == 0 {
		a.CLI.PrintSuccess("Toutes les applications sont deja installees!")
		if !a.CLI.AskYesNo("Voulez-vous quand meme reinstaller certaines applications?", false) {
			return 0
		}
	}

	// Ask user to select apps
	selectedApps := a.askMultiSelectInstall(appsStatus)
	if len(selectedApps) == 0 {
		a.CLI.PrintInfo("Aucune application selectionnee")
		return 0
	}

	// Show summary and confirm
	if !a.showInstallSummary(selectedApps) {
		a.CLI.PrintInfo("Installation annulee")
		return 0
	}

	// Run installations
	results := make([]installResult, 0, len(selectedApps))
	for _, app := range selectedApps {
		result := a.runInstaller(app)
		results = append(results, result)
	}

	// Show final summary
	return a.showFinalSummary(results)
}

// runUninstallMode runs the uninstallation mode.
func (a *App) runUninstallMode(appsStatus []appStatusEntry) int {
	// Get installed apps
	installed := make([]*AppEntry, 0)
	for _, entry := range appsStatus {
		if entry.Status == installers.StatusInstalled {
			installed = append(installed, entry.App)
		}
	}

	if len(installed) == 0 {
		a.CLI.PrintWarning("Aucune application installee a desinstaller")
		return 0
	}

	// Ask user to select apps
	selectedApps := a.askMultiSelectUninstall(installed)
	if len(selectedApps) == 0 {
		a.CLI.PrintInfo("Aucune application selectionnee")
		return 0
	}

	// Show summary and confirm
	if !a.showUninstallSummary(selectedApps) {
		a.CLI.PrintInfo("Desinstallation annulee")
		return 0
	}

	// Run uninstallations
	results := make([]installResult, 0, len(selectedApps))
	for _, app := range selectedApps {
		result := a.runUninstaller(app)
		results = append(results, result)
	}

	// Show final summary
	return a.showFinalSummary(results)
}

// askMultiSelectInstall asks the user to select applications to install.
func (a *App) askMultiSelectInstall(appsStatus []appStatusEntry) []*AppEntry {
	if a.NoInteraction {
		// In non-interactive mode, install all apps that are not installed
		result := make([]*AppEntry, 0)
		for _, entry := range appsStatus {
			if entry.Status == installers.StatusNotInstalled {
				result = append(result, entry.App)
			}
		}
		a.CLI.PrintSection("Mode non-interactif")
		a.CLI.PrintInfo(fmt.Sprintf("Installation automatique de %d application(s)", len(result)))
		return result
	}

	a.CLI.PrintSection("Selection des installations")
	fmt.Println()

	options := make([]string, 0, len(appsStatus))
	defaultIndices := make([]int, 0)

	for i, entry := range appsStatus {
		status := ""
		if entry.Status == installers.StatusInstalled {
			status = " (deja installe)"
		} else {
			defaultIndices = append(defaultIndices, i)
		}
		options = append(options, fmt.Sprintf("%s%s", entry.App.Name, status))
	}

	selected := a.CLI.AskMultiSelect("Quelles applications souhaitez-vous installer?", options, defaultIndices)

	result := make([]*AppEntry, 0, len(selected))
	for _, idx := range selected {
		if idx < len(appsStatus) {
			result = append(result, appsStatus[idx].App)
		}
	}

	return result
}

// askMultiSelectUninstall asks the user to select applications to uninstall.
func (a *App) askMultiSelectUninstall(installed []*AppEntry) []*AppEntry {
	if a.NoInteraction {
		a.CLI.PrintWarning("Desinstallation automatique non supportee en mode non-interactif")
		return nil
	}

	a.CLI.PrintSection("Selection des desinstallations")
	fmt.Println()

	options := make([]string, 0, len(installed))
	for _, app := range installed {
		options = append(options, app.Name)
	}

	selected := a.CLI.AskMultiSelect("Quelles applications souhaitez-vous desinstaller?", options, []int{})

	result := make([]*AppEntry, 0, len(selected))
	for _, idx := range selected {
		if idx < len(installed) {
			result = append(result, installed[idx])
		}
	}

	return result
}

// showInstallSummary shows the installation summary and asks for confirmation.
func (a *App) showInstallSummary(apps []*AppEntry) bool {
	a.CLI.PrintSection("Resume de l'installation")
	fmt.Println()
	cli.Bold.Println("Applications a installer:")
	for _, app := range apps {
		fmt.Printf("  %s %s - %s\n", cli.Cyan.Sprint("•"), app.Name, app.Description)
	}
	fmt.Println()

	return a.CLI.AskYesNo("Proceder a l'installation?", true)
}

// showUninstallSummary shows the uninstallation summary and asks for confirmation.
func (a *App) showUninstallSummary(apps []*AppEntry) bool {
	a.CLI.PrintSection("Resume de la desinstallation")
	fmt.Println()
	cli.BoldRed.Println("Applications a desinstaller:")
	for _, app := range apps {
		fmt.Printf("  %s %s\n", cli.Red.Sprint("•"), app.Name)
	}
	fmt.Println()
	a.CLI.PrintWarning("Cette action supprimera les applications et leurs configurations!")

	return a.CLI.AskYesNo("Etes-vous sur de vouloir continuer?", false)
}

// runInstaller runs the installer for an application.
func (a *App) runInstaller(app *AppEntry) installResult {
	fmt.Println()
	cli.Magenta.Println(strings.Repeat("═", 60))
	cli.BoldCyan.Printf("Installation de %s\n", app.Name)
	cli.Magenta.Println(strings.Repeat("═", 60))
	fmt.Println()

	opts := &installers.InstallOptions{
		DryRun:        a.DryRun,
		NoInteraction: a.NoInteraction,
	}

	result := app.Installer.Install(opts)

	return installResult{
		App:     app,
		Success: result.Success,
		Message: result.Message,
	}
}

// runUninstaller runs the uninstaller for an application.
func (a *App) runUninstaller(app *AppEntry) installResult {
	fmt.Println()
	cli.Red.Println(strings.Repeat("═", 60))
	cli.BoldRed.Printf("Desinstallation de %s\n", app.Name)
	cli.Red.Println(strings.Repeat("═", 60))
	fmt.Println()

	opts := &installers.UninstallOptions{
		DryRun:        a.DryRun,
		NoInteraction: a.NoInteraction,
		RemoveConfig:  true,
		RemoveCache:   true,
		RemoveData:    true,
	}

	result := app.Uninstaller.Uninstall(opts)

	return installResult{
		App:     app,
		Success: result.Success,
		Message: result.Message,
	}
}

// showFinalSummary shows the final summary of operations.
func (a *App) showFinalSummary(results []installResult) int {
	fmt.Println()
	a.CLI.PrintSection("Resume final")
	fmt.Println()

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
			a.CLI.PrintSuccess(fmt.Sprintf("%s installe avec succes", result.App.Name))
		} else {
			a.CLI.PrintError(fmt.Sprintf("%s - echec", result.App.Name))
		}
	}

	fmt.Println()
	failureCount := len(results) - successCount
	if failureCount == 0 {
		a.CLI.PrintSuccess(fmt.Sprintf("Toutes les operations terminees avec succes! (%d/%d)", successCount, len(results)))
		return 0
	}

	a.CLI.PrintWarning(fmt.Sprintf("Operations terminees: %d succes, %d echec(s)", successCount, failureCount))
	return 1
}

// Helper types and functions

type appStatusEntry struct {
	App     *AppEntry
	Status  installers.AppStatus
	Version string
}

type installResult struct {
	App     *AppEntry
	Success bool
	Message string
}

func boolToFrench(b bool) string {
	if b {
		return "Oui"
	}
	return "Non"
}

func formatTags(tags []installers.AppTag) string {
	if len(tags) == 0 {
		return ""
	}

	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		var tagStr string
		switch tag {
		case installers.TagApp:
			tagStr = cli.Blue.Sprintf("[%s]", tag)
		case installers.TagConfig:
			tagStr = cli.Magenta.Sprintf("[%s]", tag)
		case installers.TagAlias:
			tagStr = cli.Cyan.Sprintf("[%s]", tag)
		case installers.TagEditor:
			tagStr = cli.Green.Sprintf("[%s]", tag)
		case installers.TagShell:
			tagStr = cli.Yellow.Sprintf("[%s]", tag)
		case installers.TagContainer:
			tagStr = cli.Red.Sprintf("[%s]", tag)
		case installers.TagFont:
			tagStr = cli.White.Sprintf("[%s]", tag)
		default:
			tagStr = cli.Dim.Sprintf("[%s]", tag)
		}
		parts = append(parts, tagStr)
	}

	return strings.Join(parts, " ")
}
