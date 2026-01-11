// Package cmd provides the CLI commands for DevBootstrap.
package cmd

import (
	"fmt"
	"os"

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

func runTUI(sysInfo *system.SystemInfo) {
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
