"""Main application class for Zsh Installer."""

import sys
from typing import Optional

from .utils.system import SystemInfo, OSType
from .utils.cli import CLI, Colors
from .utils.runner import CommandRunner
from .installers.base import InstallOptions, ZshTheme
from .installers.macos import MacOSInstaller
from .installers.ubuntu import UbuntuInstaller


class ZshInstallerApp:
    """Main application for installing Zsh, Oh My Zsh and autocompletion."""

    VERSION = "1.0.0"

    def __init__(self, dry_run: bool = False, no_interaction: bool = False):
        """Initialize the application."""
        self.cli = CLI(no_interaction=no_interaction)
        self.system_info = SystemInfo.detect()
        self.runner = CommandRunner(self.cli, dry_run=dry_run)
        self.installer = self._get_installer()
        self.dry_run = dry_run
        self.no_interaction = no_interaction

    def _get_installer(self):
        """Get the appropriate installer for the current platform."""
        if self.system_info.os_type == OSType.MACOS:
            return MacOSInstaller(self.system_info, self.cli, self.runner)
        elif self.system_info.os_type in (OSType.UBUNTU, OSType.DEBIAN):
            return UbuntuInstaller(self.system_info, self.cli, self.runner)
        return None

    def print_banner(self):
        """Print the application banner."""
        self.cli.print_header("Zsh Installer v" + self.VERSION)

    def print_system_info(self):
        """Print detected system information."""
        self.cli.show_summary("Informations systeme", {
            "OS": str(self.system_info),
            "Architecture": self.system_info.architecture.value,
            "Repertoire home": self.system_info.home_dir,
            "Fichier .zshrc": self.system_info.get_zshrc_path(),
            "Privileges root": "Oui" if self.system_info.is_root else "Non",
        })

    def check_prerequisites(self) -> bool:
        """Check if the system meets prerequisites."""
        self.cli.print_section("Verification des prerequis")

        # Check if system is supported
        if not self.system_info.is_supported():
            self.cli.print_error(f"Systeme non supporte: {self.system_info.os_type.value}")
            self.cli.print_info("Systemes supportes: macOS, Ubuntu, Debian")
            return False

        self.cli.print_success(f"Systeme supporte: {self.system_info.os_name}")

        # Check for existing Zsh installation
        existing_zsh = self.runner.get_command_path("zsh")
        if existing_zsh:
            version = self.runner.get_command_version("zsh")
            self.cli.print_info(f"Zsh existant detecte: {version}")
            self.cli.print_info(f"Chemin: {existing_zsh}")

        # Check for existing Oh My Zsh installation
        import os
        if os.path.exists(self.system_info.get_oh_my_zsh_dir()):
            self.cli.print_info("Oh My Zsh deja installe")

        return True

    def show_dependency_status(self):
        """Show the status of all dependencies."""
        self.cli.print_section("Etat des dependances")

        deps = self.installer.check_all_dependencies()
        for name, installed in deps.items():
            if installed:
                self.cli.print_success(f"{name}")
            else:
                self.cli.print_warning(f"{name} - non installe")

    def get_install_options(self) -> Optional[InstallOptions]:
        """Interactively get installation options from user."""
        self.cli.print_section("Configuration de l'installation")

        # Ask about Oh My Zsh
        install_oh_my_zsh = self.cli.ask_yes_no(
            "Installer Oh My Zsh (framework de configuration)?",
            default=True
        )

        # Ask about plugins (only if Oh My Zsh is being installed)
        install_autosuggestions = False
        install_syntax_highlighting = False
        install_autocompletion = True
        theme = "robbyrussell"

        if install_oh_my_zsh:
            install_autosuggestions = self.cli.ask_yes_no(
                "Installer zsh-autosuggestions (suggestions basees sur l'historique)?",
                default=True
            )

            install_syntax_highlighting = self.cli.ask_yes_no(
                "Installer zsh-syntax-highlighting (coloration syntaxique)?",
                default=True
            )

            install_autocompletion = self.cli.ask_yes_no(
                "Installer zsh-completions (completions additionnelles)?",
                default=True
            )

            # Theme selection
            theme_choice = self.cli.ask_choice(
                "Quel theme utiliser?",
                [
                    "robbyrussell (defaut, simple)",
                    "agnoster (powerline, necessite polices speciales)",
                    "Garder le theme par defaut"
                ],
                default=0
            )

            if theme_choice == 0:
                theme = "robbyrussell"
            elif theme_choice == 1:
                theme = "agnoster"
            else:
                theme = "robbyrussell"

        # Ask about default shell
        set_as_default = self.cli.ask_yes_no(
            "Definir Zsh comme shell par defaut?",
            default=True
        )

        return InstallOptions(
            install_oh_my_zsh=install_oh_my_zsh,
            install_autocompletion=install_autocompletion,
            install_syntax_highlighting=install_syntax_highlighting,
            install_autosuggestions=install_autosuggestions,
            set_as_default_shell=set_as_default,
            theme=theme,
            backup_existing=True
        )

    def run_installation(self, options: InstallOptions) -> bool:
        """Run the Zsh installation."""
        result = self.installer.install(options)

        if result.success:
            self.cli.print_section("Installation terminee")
            self.cli.print_success(result.message)

            if result.zsh_path:
                self.cli.print_info(f"Chemin: {result.zsh_path}")
            if result.zsh_version:
                self.cli.print_info(f"Version: {result.zsh_version}")

            if result.warnings:
                self.cli.print()
                self.cli.print_warning("Avertissements:")
                for warning in result.warnings:
                    self.cli.print(f"  - {warning}")

            return True
        else:
            self.cli.print_section("Echec de l'installation")
            self.cli.print_error(result.message)

            if result.errors:
                for error in result.errors:
                    self.cli.print_error(f"  - {error}")

            return False

    def show_final_instructions(self):
        """Show final instructions after installation."""
        self.cli.print_section("Prochaines etapes")

        self.cli.print()
        self.cli.print(f"{Colors.BOLD}1. Redemarrer le terminal{Colors.RESET}")
        self.cli.print("   Pour que les changements de shell prennent effet")
        self.cli.print()

        self.cli.print(f"{Colors.BOLD}2. Ou executer directement:{Colors.RESET}")
        self.cli.print("   $ zsh")
        self.cli.print()

        self.cli.print(f"{Colors.BOLD}3. Si le shell n'a pas change:{Colors.RESET}")
        self.cli.print("   $ chsh -s $(which zsh)")
        self.cli.print("   Puis se deconnecter/reconnecter")
        self.cli.print()

        self.cli.print(f"{Colors.BOLD}Fonctionnalites installees:{Colors.RESET}")
        self.cli.print("   - Oh My Zsh: framework de configuration")
        self.cli.print("   - Autosuggestions: suggestions basees sur l'historique")
        self.cli.print("   - Syntax highlighting: coloration des commandes")
        self.cli.print("   - Completions: completions avancees")
        self.cli.print("   - bash-completion: completions pour Bash")
        self.cli.print()

        self.cli.print(f"{Colors.BOLD}Raccourcis utiles (Zsh):{Colors.RESET}")
        self.cli.print("   Tab         - Autocompletion")
        self.cli.print("   Ctrl+R      - Recherche dans l'historique")
        self.cli.print("   ->          - Accepter la suggestion")
        self.cli.print("   Alt+->      - Accepter un mot de la suggestion")
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

            # Show dependency status
            if self.installer:
                self.show_dependency_status()

            # Confirm to proceed
            if not self.cli.ask_yes_no("Proceder a l'installation?", default=True):
                self.cli.print_info("Installation annulee")
                return 0

            # Get install options
            install_options = self.get_install_options()
            if not install_options:
                return 1

            # Summary
            self.cli.show_summary("Resume de l'installation", {
                "Oh My Zsh": "Oui" if install_options.install_oh_my_zsh else "Non",
                "Autosuggestions": "Oui" if install_options.install_autosuggestions else "Non",
                "Syntax highlighting": "Oui" if install_options.install_syntax_highlighting else "Non",
                "Completions": "Oui" if install_options.install_autocompletion else "Non",
                "Theme": install_options.theme,
                "Shell par defaut": "Oui" if install_options.set_as_default_shell else "Non",
            })

            if not self.cli.ask_yes_no("Confirmer et lancer l'installation?", default=True):
                self.cli.print_info("Installation annulee")
                return 0

            # Run installation
            if not self.run_installation(install_options):
                return 1

            # Show final instructions
            self.show_final_instructions()

            self.cli.print_success("Installation terminee avec succes!")
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
        description="Zsh Installer - Installation automatique de Zsh, Oh My Zsh et autocompletion"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simuler l'installation sans effectuer de changements"
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"Zsh Installer {ZshInstallerApp.VERSION}"
    )

    args = parser.parse_args()

    app = ZshInstallerApp(dry_run=args.dry_run)
    sys.exit(app.run())


if __name__ == "__main__":
    main()
