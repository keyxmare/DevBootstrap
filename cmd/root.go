// Package cmd provides the CLI commands for DevBootstrap.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/keyxmare/DevBootstrap/internal/adapter/primary/cli"
	"github.com/keyxmare/DevBootstrap/internal/adapter/primary/tui"
	"github.com/keyxmare/DevBootstrap/internal/config"
)

const version = "3.0.0"

var (
	dryRun        bool
	noInteraction bool
	uninstall     bool
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
		// Create DI container
		container := config.NewContainer(dryRun, noInteraction, uninstall)

		// Check for root on macOS
		if container.Platform.IsMacOS() && container.Platform.IsRoot() {
			fmt.Println("Erreur: Ne pas executer devbootstrap avec sudo sur macOS.")
			fmt.Println("Homebrew ne fonctionne pas correctement en root.")
			fmt.Println()
			fmt.Println("Utilisez simplement: devbootstrap")
			os.Exit(1)
		}

		// Use TUI for interactive mode, CLI for non-interactive
		if noInteraction {
			app := cli.NewApp(container)
			os.Exit(app.Run())
		} else {
			app := tui.NewApp(container)
			os.Exit(app.Run())
		}
	},
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
	rootCmd.Version = version
}
