// Package cmd provides the CLI commands for DevBootstrap.
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/keyxmare/DevBootstrap/internal/app"
	"github.com/keyxmare/DevBootstrap/internal/cli"
	"github.com/keyxmare/DevBootstrap/internal/runner"
	"github.com/keyxmare/DevBootstrap/internal/system"
	"github.com/keyxmare/DevBootstrap/internal/tui"
)

const version = "2.0.0"

var (
	dryRun        bool
	noInteraction bool
	uninstall     bool
	legacyMode    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "devbootstrap",
	Short: "Suite d'installation automatique pour environnement de developpement",
	Long: `DevBootstrap est une suite d'installation automatique pour configurer
votre environnement de developpement sur macOS, Ubuntu et Debian.

Applications disponibles:
  - Docker (conteneurisation)
  - Visual Studio Code (editeur)
  - Neovim (editeur terminal)
  - Zsh & Oh My Zsh (shell)
  - Nerd Fonts (polices terminal)`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for root on macOS
		sysInfo := system.Detect()
		if sysInfo.IsMacOS() && sysInfo.IsRoot {
			fmt.Println("Erreur: Ne pas executer devbootstrap avec sudo sur macOS.")
			fmt.Println("Homebrew ne fonctionne pas correctement en root.")
			fmt.Println()
			fmt.Println("Utilisez simplement: devbootstrap")
			os.Exit(1)
		}

		// Use legacy mode if requested or in non-interactive mode
		if legacyMode || noInteraction {
			application := app.New(dryRun, noInteraction, uninstall)
			os.Exit(application.Run())
			return
		}

		// Run the new TUI
		runTUI(sysInfo)
	},
}

// preCacheSudo asks for sudo password upfront
func preCacheSudo(sysInfo *system.SystemInfo) {
	// Skip if already root
	if sysInfo.IsRoot {
		return
	}

	// Skip in dry-run mode
	if dryRun {
		return
	}

	// Check if sudo is available
	if _, err := exec.LookPath("sudo"); err != nil {
		return
	}

	// Check if sudo credentials are already cached
	checkCmd := exec.Command("sudo", "-n", "true")
	if checkCmd.Run() == nil {
		// Already authenticated
		return
	}

	// Ask for password
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4"))
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
		fmt.Println()
	} else {
		fmt.Println()
	}
}

func runTUI(sysInfo *system.SystemInfo) {
	// Pre-cache sudo credentials on Linux
	preCacheSudo(sysInfo)

	// Create components
	cliUtil := cli.New(noInteraction, dryRun)
	r := runner.New(cliUtil, dryRun)

	// Create registry to get apps
	registry := app.NewRegistry(cliUtil, r, sysInfo)

	// Create TUI app
	tuiApp := tui.NewApp(sysInfo, r, uninstall, noInteraction, dryRun)

	// Add apps from registry
	for _, entry := range registry.GetAll() {
		status, version := entry.Installer.CheckExisting()
		tuiApp.AddApp(
			entry.ID,
			entry.Name,
			entry.Description,
			entry.Tags,
			status,
			version,
			entry.Installer,
			entry.Uninstaller,
		)
	}

	// Run
	if err := tuiApp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simuler sans effectuer de changements")
	rootCmd.Flags().BoolVarP(&noInteraction, "no-interaction", "n", false, "Mode non-interactif (utilise les valeurs par defaut)")
	rootCmd.Flags().BoolVarP(&uninstall, "uninstall", "u", false, "Mode desinstallation")
	rootCmd.Flags().BoolVar(&legacyMode, "legacy", false, "Utiliser l'ancienne interface")
	rootCmd.Version = version
}
