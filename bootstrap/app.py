"""Main Bootstrap application with unified menu."""

import sys
import subprocess
import shutil
from typing import Optional

from .apps import AVAILABLE_APPS, AppInfo, AppStatus


class Colors:
    """ANSI color codes for terminal output."""
    RESET = "\033[0m"
    BOLD = "\033[1m"
    DIM = "\033[2m"

    RED = "\033[31m"
    GREEN = "\033[32m"
    YELLOW = "\033[33m"
    BLUE = "\033[34m"
    MAGENTA = "\033[35m"
    CYAN = "\033[36m"
    WHITE = "\033[37m"

    @classmethod
    def disable(cls):
        """Disable all colors (for non-TTY output)."""
        for attr in dir(cls):
            if not attr.startswith('_') and attr.isupper():
                setattr(cls, attr, "")


class BootstrapApp:
    """Main application for DevBootstrap installer suite."""

    VERSION = "1.0.0"

    def __init__(self, dry_run: bool = False, no_interaction: bool = False):
        """Initialize the application."""
        self.dry_run = dry_run
        self.no_interaction = no_interaction
        self.use_colors = sys.stdout.isatty()
        if not self.use_colors:
            Colors.disable()

    def print(self, message: str = "", end: str = "\n"):
        """Print a message."""
        print(message, end=end, flush=True)

    def print_banner(self):
        """Print the application banner."""
        title = f"DevBootstrap v{self.VERSION}"
        width = max(60, len(title) + 4)
        border = "═" * width

        self.print()
        self.print(f"{Colors.CYAN}{Colors.BOLD}╔{border}╗{Colors.RESET}")
        self.print(f"{Colors.CYAN}{Colors.BOLD}║{Colors.RESET} {title.center(width - 2)} {Colors.CYAN}{Colors.BOLD}║{Colors.RESET}")
        self.print(f"{Colors.CYAN}{Colors.BOLD}╚{border}╝{Colors.RESET}")
        self.print()
        self.print(f"{Colors.DIM}Suite d'installation automatique pour votre environnement de développement{Colors.RESET}")
        self.print()

    def print_section(self, title: str):
        """Print a section header."""
        self.print()
        self.print(f"{Colors.BLUE}{Colors.BOLD}▶ {title}{Colors.RESET}")
        self.print(f"{Colors.DIM}{'─' * 50}{Colors.RESET}")

    def print_success(self, message: str):
        """Print a success message."""
        self.print(f"{Colors.GREEN}✓{Colors.RESET} {message}")

    def print_error(self, message: str):
        """Print an error message."""
        self.print(f"{Colors.RED}✗{Colors.RESET} {message}")

    def print_warning(self, message: str):
        """Print a warning message."""
        self.print(f"{Colors.YELLOW}⚠{Colors.RESET} {message}")

    def print_info(self, message: str):
        """Print an info message."""
        self.print(f"{Colors.CYAN}ℹ{Colors.RESET} {message}")

    def check_app_status(self, app: AppInfo) -> tuple[AppStatus, Optional[str]]:
        """Check if an application is installed and get its version."""
        import os

        # Check if command exists in PATH
        command_exists = shutil.which(app.check_command) is not None

        # Check macOS app paths if command not in PATH
        app_exists = False
        if not command_exists and app.macos_app_paths:
            for app_path in app.macos_app_paths:
                if os.path.exists(app_path):
                    app_exists = True
                    break

        if not command_exists and not app_exists:
            return AppStatus.NOT_INSTALLED, None

        version = None
        if app.version_command and command_exists:
            try:
                result = subprocess.run(
                    app.version_command,
                    shell=True,
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                if result.returncode == 0:
                    version = result.stdout.strip().split('\n')[0]
            except (subprocess.TimeoutExpired, Exception):
                pass

        # If app exists but no version (command not in PATH)
        if app_exists and not version:
            version = "(commande 'code' non configurée)"

        return AppStatus.INSTALLED, version

    def display_available_apps(self) -> list[tuple[AppInfo, AppStatus, Optional[str]]]:
        """Display all available applications with their status."""
        self.print_section("Applications disponibles")
        self.print()

        apps_status = []
        for app in AVAILABLE_APPS:
            status, version = self.check_app_status(app)
            apps_status.append((app, status, version))

            # Status indicator
            if status == AppStatus.INSTALLED:
                status_icon = f"{Colors.GREEN}✓{Colors.RESET}"
                status_text = f"{Colors.GREEN}installé{Colors.RESET}"
                if version:
                    status_text += f" {Colors.DIM}({version}){Colors.RESET}"
            else:
                status_icon = f"{Colors.YELLOW}○{Colors.RESET}"
                status_text = f"{Colors.YELLOW}non installé{Colors.RESET}"

            self.print(f"  {status_icon} {Colors.BOLD}{app.name}{Colors.RESET}")
            self.print(f"      {Colors.DIM}{app.description}{Colors.RESET}")
            self.print(f"      Status: {status_text}")
            self.print()

        return apps_status

    def ask_yes_no(self, question: str, default: bool = True) -> bool:
        """Ask a yes/no question and return the answer."""
        default_str = "O/n" if default else "o/N"
        prompt = f"{Colors.YELLOW}?{Colors.RESET} {question} [{default_str}]: "

        while True:
            try:
                response = input(prompt).strip().lower()

                if not response:
                    return default

                if response in ("o", "oui", "y", "yes"):
                    return True
                elif response in ("n", "non", "no"):
                    return False
                else:
                    self.print_warning("Répondez par 'o' (oui) ou 'n' (non)")

            except EOFError:
                return default
            except KeyboardInterrupt:
                self.print()
                raise

    def ask_multi_select(
        self,
        question: str,
        apps_status: list[tuple[AppInfo, AppStatus, Optional[str]]]
    ) -> list[AppInfo]:
        """Ask user to select multiple applications to install."""
        self.print_section("Sélection des installations")
        self.print()
        self.print(f"{Colors.YELLOW}?{Colors.RESET} {question}")
        self.print(f"  {Colors.DIM}(Entrez les numéros séparés par des espaces, ou 'tous' pour tout sélectionner){Colors.RESET}")
        self.print()

        # List only apps that can be installed (not already installed or user wants to reinstall)
        selectable_apps = []
        for i, (app, status, version) in enumerate(apps_status, 1):
            if status == AppStatus.NOT_INSTALLED:
                marker = f"{Colors.YELLOW}○{Colors.RESET}"
                status_hint = ""
            else:
                marker = f"{Colors.GREEN}✓{Colors.RESET}"
                status_hint = f" {Colors.DIM}(déjà installé){Colors.RESET}"

            self.print(f"  [{i}] {marker} {app.name}{status_hint}")
            selectable_apps.append(app)

        self.print()
        prompt = f"  Votre choix: "

        while True:
            try:
                response = input(prompt).strip().lower()

                if not response:
                    # Default: install apps that are not installed
                    return [app for app, status, _ in apps_status if status == AppStatus.NOT_INSTALLED]

                if response in ("tous", "all", "*"):
                    return [app for app, _, _ in apps_status]

                if response in ("aucun", "none", "0"):
                    return []

                # Parse space-separated numbers
                try:
                    indices = [int(x) for x in response.split()]
                    selected = []
                    for idx in indices:
                        if 1 <= idx <= len(apps_status):
                            selected.append(apps_status[idx - 1][0])
                        else:
                            raise ValueError(f"Index {idx} invalide")
                    return selected
                except ValueError as e:
                    self.print_warning(f"Entrée invalide: {e}")
                    self.print_info(f"Entrez des numéros entre 1 et {len(apps_status)}, séparés par des espaces")

            except EOFError:
                return []
            except KeyboardInterrupt:
                self.print()
                raise

    def show_installation_summary(self, selected_apps: list[AppInfo]) -> bool:
        """Show a summary of what will be installed and ask for confirmation."""
        self.print_section("Résumé de l'installation")
        self.print()

        if not selected_apps:
            self.print_info("Aucune application sélectionnée pour l'installation.")
            return False

        self.print(f"{Colors.BOLD}Applications à installer:{Colors.RESET}")
        for app in selected_apps:
            self.print(f"  {Colors.CYAN}•{Colors.RESET} {app.name} - {app.description}")

        self.print()
        return self.ask_yes_no("Procéder à l'installation?", default=True)

    def run_installer(self, app: AppInfo) -> bool:
        """Run the installer for a specific application."""
        self.print()
        self.print(f"{Colors.MAGENTA}{'═' * 60}{Colors.RESET}")
        self.print(f"{Colors.MAGENTA}{Colors.BOLD}Installation de {app.name}{Colors.RESET}")
        self.print(f"{Colors.MAGENTA}{'═' * 60}{Colors.RESET}")
        self.print()

        if not app.module:
            self.print_error(f"Aucun module d'installation défini pour {app.name}")
            return False

        try:
            # Import and run the installer module
            if app.id == "neovim":
                from nvim_installer.app import NeovimInstallerApp
                installer = NeovimInstallerApp(dry_run=self.dry_run)
                return installer.run() == 0
            elif app.id == "docker":
                from docker_installer.app import DockerInstallerApp
                installer = DockerInstallerApp(dry_run=self.dry_run)
                return installer.run() == 0
            elif app.id == "vscode":
                from vscode_installer.app import VSCodeInstallerApp
                installer = VSCodeInstallerApp(dry_run=self.dry_run)
                return installer.run() == 0
            elif app.id == "zsh":
                from zsh_installer.app import ZshInstallerApp
                installer = ZshInstallerApp(dry_run=self.dry_run)
                return installer.run() == 0
            elif app.id == "alias":
                from alias_installer.app import AliasInstallerApp
                installer = AliasInstallerApp(dry_run=self.dry_run)
                return installer.run() == 0
            else:
                self.print_error(f"Installateur inconnu pour {app.name}")
                return False

        except ImportError as e:
            self.print_error(f"Impossible de charger l'installateur: {e}")
            return False
        except Exception as e:
            self.print_error(f"Erreur lors de l'installation: {e}")
            return False

    def run(self) -> int:
        """Run the complete bootstrap process."""
        try:
            # Print banner
            self.print_banner()

            # Display available apps with status
            apps_status = self.display_available_apps()

            # Check if all apps are already installed
            not_installed = [app for app, status, _ in apps_status if status == AppStatus.NOT_INSTALLED]

            if not not_installed:
                self.print_success("Toutes les applications sont déjà installées!")
                if not self.ask_yes_no("Voulez-vous quand même réinstaller certaines applications?", default=False):
                    return 0

            # Ask user to select apps to install
            selected_apps = self.ask_multi_select(
                "Quelles applications souhaitez-vous installer?",
                apps_status
            )

            # Show summary and confirm
            if not self.show_installation_summary(selected_apps):
                self.print_info("Installation annulée")
                return 0

            # Run installations
            results = []
            for app in selected_apps:
                success = self.run_installer(app)
                results.append((app, success))

            # Show final summary
            self.print()
            self.print_section("Résumé final")
            self.print()

            success_count = sum(1 for _, success in results if success)
            failure_count = len(results) - success_count

            for app, success in results:
                if success:
                    self.print_success(f"{app.name} installé avec succès")
                else:
                    self.print_error(f"{app.name} - échec de l'installation")

            self.print()
            if failure_count == 0:
                self.print_success(f"Toutes les installations terminées avec succès! ({success_count}/{len(results)})")
            else:
                self.print_warning(f"Installations terminées: {success_count} succès, {failure_count} échec(s)")

            return 0 if failure_count == 0 else 1

        except KeyboardInterrupt:
            self.print()
            self.print_warning("Installation interrompue par l'utilisateur")
            return 130

        except Exception as e:
            self.print_error(f"Erreur inattendue: {e}")
            if self.dry_run:
                import traceback
                traceback.print_exc()
            return 1


def main():
    """Main entry point."""
    import argparse

    parser = argparse.ArgumentParser(
        description="DevBootstrap - Suite d'installation pour environnement de développement"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simuler l'installation sans effectuer de changements"
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"DevBootstrap {BootstrapApp.VERSION}"
    )

    args = parser.parse_args()

    app = BootstrapApp(dry_run=args.dry_run)
    sys.exit(app.run())


if __name__ == "__main__":
    main()
