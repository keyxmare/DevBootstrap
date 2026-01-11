"""Main application class for Neovim Installer."""

import sys
from typing import Optional

from .utils.system import SystemInfo, OSType
from .utils.cli import CLI, Colors
from .utils.runner import CommandRunner
from .installers.base import InstallOptions
from .installers.macos import MacOSInstaller
from .installers.ubuntu import UbuntuInstaller
from .config_manager import ConfigManager, ConfigOptions, ConfigPreset


class NeovimInstallerApp:
    """Main application for installing Neovim."""

    VERSION = "1.0.0"

    # Mode: "install" for Neovim only, "config" for config/plugins only, "full" for both
    def __init__(self, dry_run: bool = False, no_interaction: bool = False, mode: str = "full"):
        """Initialize the application."""
        self.cli = CLI(no_interaction=no_interaction)
        self.system_info = SystemInfo.detect()
        self.runner = CommandRunner(self.cli, dry_run=dry_run)
        self.installer = self._get_installer()
        self.config_manager = ConfigManager(self.system_info, self.cli, self.runner)
        self.dry_run = dry_run
        self.no_interaction = no_interaction
        self.mode = mode  # "install", "config", or "full"

    def _get_installer(self):
        """Get the appropriate installer for the current platform."""
        if self.system_info.os_type == OSType.MACOS:
            return MacOSInstaller(self.system_info, self.cli, self.runner)
        elif self.system_info.os_type in (OSType.UBUNTU, OSType.DEBIAN):
            return UbuntuInstaller(self.system_info, self.cli, self.runner)
        return None

    def print_banner(self):
        """Print the application banner."""
        if self.mode == "install":
            self.cli.print_header("Neovim Installer v" + self.VERSION)
        elif self.mode == "config":
            self.cli.print_header("Neovim Config Installer v" + self.VERSION)
        else:
            self.cli.print_header("Neovim + Config Installer v" + self.VERSION)

    def print_system_info(self):
        """Print detected system information."""
        self.cli.show_summary("Informations système", {
            "OS": str(self.system_info),
            "Architecture": self.system_info.architecture.value,
            "Répertoire home": self.system_info.home_dir,
            "Config Neovim": self.system_info.get_config_dir(),
            "Privilèges root": "Oui" if self.system_info.is_root else "Non",
        })

    def check_prerequisites(self) -> bool:
        """Check if the system meets prerequisites."""
        self.cli.print_section("Vérification des prérequis")

        # Check if system is supported
        if not self.system_info.is_supported():
            self.cli.print_error(f"Système non supporté: {self.system_info.os_type.value}")
            self.cli.print_info("Systèmes supportés: macOS, Ubuntu, Debian")
            return False

        self.cli.print_success(f"Système supporté: {self.system_info.os_name}")

        # Check for existing Neovim installation
        existing_nvim = self.runner.get_command_path("nvim")
        if existing_nvim:
            version = self.runner.get_command_version("nvim")
            self.cli.print_info(f"Neovim existant détecté: {version}")
            self.cli.print_info(f"Chemin: {existing_nvim}")
        elif self.mode == "config":
            # Config mode requires Neovim to be installed
            self.cli.print_error("Neovim n'est pas installé. Installez d'abord Neovim.")
            return False

        return True

    def show_dependency_status(self):
        """Show the status of all dependencies."""
        self.cli.print_section("État des dépendances")

        deps = self.installer.check_all_dependencies()
        for name, installed in deps.items():
            if installed:
                self.cli.print_success(f"{name}")
            else:
                self.cli.print_warning(f"{name} - non installé")

    def get_install_options(self) -> Optional[InstallOptions]:
        """Interactively get installation options from user."""
        self.cli.print_section("Configuration de l'installation")

        # Default values
        neovim_version = "stable"
        install_deps = True
        install_config = False
        default_config_dir = self.system_info.get_config_dir()

        if self.mode == "install":
            # Neovim only mode - just install Neovim
            version_choice = self.cli.ask_choice(
                "Quelle version de Neovim installer?",
                ["Stable (recommandé)", "Nightly (dernières fonctionnalités)"],
                default=0
            )
            neovim_version = "stable" if version_choice == 0 else "nightly"

            install_deps = self.cli.ask_yes_no(
                "Installer les dépendances recommandées (ripgrep, fzf, etc.)?",
                default=True
            )

            # No config in install-only mode
            install_config = False

        elif self.mode == "config":
            # Config only mode - skip Neovim installation
            install_config = True

        else:
            # Full mode - ask everything
            version_choice = self.cli.ask_choice(
                "Quelle version de Neovim installer?",
                ["Stable (recommandé)", "Nightly (dernières fonctionnalités)"],
                default=0
            )
            neovim_version = "stable" if version_choice == 0 else "nightly"

            install_deps = self.cli.ask_yes_no(
                "Installer les dépendances recommandées (ripgrep, fzf, etc.)?",
                default=True
            )

            install_config = self.cli.ask_yes_no(
                "Installer une configuration Neovim?",
                default=True
            )

        # Get config directory if config will be installed
        config_dir = default_config_dir
        if install_config or self.mode == "config":
            config_dir = self.cli.ask_path(
                "Répertoire de configuration Neovim",
                default=default_config_dir
            )

        return InstallOptions(
            config_dir=config_dir,
            install_dependencies=install_deps,
            install_config=install_config,
            backup_existing=True,
            neovim_version=neovim_version
        )

    def get_config_options(self) -> Optional[ConfigOptions]:
        """Interactively get configuration options from user."""
        self.cli.print_section("Configuration de Neovim")

        # Ask for preset
        preset_choice = self.cli.ask_choice(
            "Quel preset de configuration utiliser?",
            [
                "Minimal - Options de base uniquement (sans plugins)",
                "Complet - Configuration complète avec tous les plugins",
                "Personnalisé - Importer depuis un chemin/URL"
            ],
            default=1
        )

        preset_map = {
            0: ConfigPreset.MINIMAL,
            1: ConfigPreset.FULL,
            2: ConfigPreset.CUSTOM
        }

        preset = preset_map[preset_choice]

        custom_path = None
        if preset == ConfigPreset.CUSTOM:
            custom_path = self.cli.ask_path(
                "Chemin ou URL Git de la configuration",
                default=""
            )
            if not custom_path:
                self.cli.print_warning("Aucun chemin fourni, utilisation du preset Complet")
                preset = ConfigPreset.FULL

        # Ask about plugin sync
        install_plugins = self.cli.ask_yes_no(
            "Synchroniser les plugins après l'installation?",
            default=True
        )

        return ConfigOptions(
            preset=preset,
            config_dir=self.system_info.get_config_dir(),
            backup_existing=True,
            install_plugins=install_plugins,
            custom_config_path=custom_path
        )

    def run_installation(self, options: InstallOptions) -> bool:
        """Run the Neovim installation."""
        result = self.installer.install(options)

        if result.success:
            self.cli.print_section("Installation terminée")
            self.cli.print_success(result.message)

            if result.neovim_path:
                self.cli.print_info(f"Chemin: {result.neovim_path}")
            if result.neovim_version:
                self.cli.print_info(f"Version: {result.neovim_version}")

            if result.warnings:
                self.cli.print()
                self.cli.print_warning("Avertissements:")
                for warning in result.warnings:
                    self.cli.print(f"  - {warning}")

            return True
        else:
            self.cli.print_section("Échec de l'installation")
            self.cli.print_error(result.message)

            if result.errors:
                for error in result.errors:
                    self.cli.print_error(f"  - {error}")

            return False

    def run_config_setup(self, options: ConfigOptions) -> bool:
        """Run the configuration setup."""
        return self.config_manager.setup(options)

    def show_final_instructions(self):
        """Show final instructions after installation."""
        self.cli.print_section("Prochaines étapes")

        self.cli.print()

        if self.mode == "install":
            # Neovim only mode
            self.cli.print(f"{Colors.BOLD}1. Redémarrer le terminal{Colors.RESET}")
            self.cli.print("   Pour que les changements de PATH prennent effet")
            self.cli.print()

            self.cli.print(f"{Colors.BOLD}2. Lancer Neovim{Colors.RESET}")
            self.cli.print("   $ nvim")
            self.cli.print()

            self.cli.print(f"{Colors.BOLD}Prochaine étape recommandée:{Colors.RESET}")
            self.cli.print("   Installer la configuration Neovim pour une meilleure experience:")
            self.cli.print("   $ devbootstrap  # puis sélectionner 'Neovim Config'")
            self.cli.print()

        elif self.mode == "config":
            # Config only mode
            self.cli.print(f"{Colors.BOLD}1. Lancer Neovim{Colors.RESET}")
            self.cli.print("   $ nvim")
            self.cli.print()

            self.cli.print(f"{Colors.BOLD}2. Premier lancement{Colors.RESET}")
            self.cli.print("   Les plugins seront automatiquement installés")
            self.cli.print("   Attendre que Lazy.nvim termine l'installation")
            self.cli.print()

            self.cli.print(f"{Colors.BOLD}Raccourcis utiles:{Colors.RESET}")
            self.cli.print("   <Space>ff  - Rechercher des fichiers")
            self.cli.print("   <Space>fg  - Rechercher du texte")
            self.cli.print("   <Space>e   - Explorateur de fichiers")
            self.cli.print("   <Space>gg  - LazyGit (si installé)")
            self.cli.print()

        else:
            # Full mode
            self.cli.print(f"{Colors.BOLD}1. Redémarrer le terminal{Colors.RESET}")
            self.cli.print("   Pour que les changements de PATH prennent effet")
            self.cli.print()

            self.cli.print(f"{Colors.BOLD}2. Lancer Neovim{Colors.RESET}")
            self.cli.print("   $ nvim")
            self.cli.print()

            self.cli.print(f"{Colors.BOLD}3. Premier lancement{Colors.RESET}")
            self.cli.print("   Les plugins seront automatiquement installés")
            self.cli.print("   Attendre que Lazy.nvim termine l'installation")
            self.cli.print()

            self.cli.print(f"{Colors.BOLD}Raccourcis utiles:{Colors.RESET}")
            self.cli.print("   <Space>ff  - Rechercher des fichiers")
            self.cli.print("   <Space>fg  - Rechercher du texte")
            self.cli.print("   <Space>e   - Explorateur de fichiers")
            self.cli.print("   <Space>gg  - LazyGit (si installé)")
            self.cli.print()

    def run(self) -> int:
        """Run the complete installation process."""
        try:
            # Print banner
            self.print_banner()

            # Print system info
            self.print_system_info()

            # Check prerequisites
            if not self.check_prerequisites():
                return 1

            # Show dependency status (only if installing Neovim)
            if self.installer and self.mode != "config":
                self.show_dependency_status()

            # Confirm to proceed
            if not self.cli.ask_yes_no("Procéder à l'installation?", default=True):
                self.cli.print_info("Installation annulée")
                return 0

            # Config-only mode: skip Neovim installation
            if self.mode == "config":
                # Get config options directly
                config_options = self.get_config_options()
                if config_options:
                    # Summary for config mode
                    self.cli.show_summary("Résumé de l'installation", {
                        "Preset": config_options.preset.value,
                        "Répertoire config": config_options.config_dir,
                        "Plugins": "Oui" if config_options.install_plugins else "Non",
                    })

                    if not self.cli.ask_yes_no("Confirmer et lancer l'installation?", default=True):
                        self.cli.print_info("Installation annulée")
                        return 0

                    if not self.run_config_setup(config_options):
                        self.cli.print_warning("Configuration partiellement installée")

                # Show final instructions
                self.show_final_instructions()
                self.cli.print_success("Installation terminée avec succès!")
                return 0

            # Install mode or full mode: install Neovim
            install_options = self.get_install_options()
            if not install_options:
                return 1

            # Summary
            summary = {
                "Version Neovim": install_options.neovim_version,
                "Dépendances": "Oui" if install_options.install_dependencies else "Non",
            }
            if self.mode == "full":
                summary["Configuration"] = "Oui" if install_options.install_config else "Non"
                summary["Répertoire config"] = install_options.config_dir

            self.cli.show_summary("Résumé de l'installation", summary)

            if not self.cli.ask_yes_no("Confirmer et lancer l'installation?", default=True):
                self.cli.print_info("Installation annulée")
                return 0

            # Run installation
            if not self.run_installation(install_options):
                return 1

            # Configuration setup if requested (only in full mode)
            if self.mode == "full" and install_options.install_config:
                config_options = self.get_config_options()
                if config_options:
                    if not self.run_config_setup(config_options):
                        self.cli.print_warning("Configuration partiellement installée")

            # Show final instructions
            self.show_final_instructions()

            self.cli.print_success("Installation terminée avec succès!")
            return 0

        except KeyboardInterrupt:
            self.cli.print()
            self.cli.print_warning("Installation interrompue par l'utilisateur")
            return 130

        except Exception as e:
            self.cli.print_error(f"Erreur inattendue: {e}")
            if self.dry_run:
                import traceback
                traceback.print_exc()
            return 1


def main():
    """Main entry point."""
    import argparse

    parser = argparse.ArgumentParser(
        description="Neovim Installer - Installation automatique de Neovim"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simuler l'installation sans effectuer de changements"
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"Neovim Installer {NeovimInstallerApp.VERSION}"
    )

    args = parser.parse_args()

    app = NeovimInstallerApp(dry_run=args.dry_run)
    sys.exit(app.run())


if __name__ == "__main__":
    main()
