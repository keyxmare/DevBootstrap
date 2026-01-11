"""Main Bootstrap application with unified menu."""

import sys
import subprocess
import shutil
from typing import Optional

from .apps import AVAILABLE_APPS, AppInfo, AppStatus, AppTag


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

    VERSION = "1.1.0"

    def __init__(self, dry_run: bool = False, no_interaction: bool = False, mode: str = "install"):
        """Initialize the application.

        Args:
            dry_run: If True, simulate without making changes
            no_interaction: If True, use defaults without prompting
            mode: "install" for installation, "uninstall" for uninstallation
        """
        self.dry_run = dry_run
        self.no_interaction = no_interaction
        self.mode = mode
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

    def format_tags(self, tags: list[AppTag]) -> str:
        """Format tags for display with colors."""
        if not tags:
            return ""

        # Color mapping for different tag types
        tag_colors = {
            AppTag.APP: Colors.BLUE,
            AppTag.CONFIG: Colors.MAGENTA,
            AppTag.ALIAS: Colors.CYAN,
            AppTag.EDITOR: Colors.GREEN,
            AppTag.SHELL: Colors.YELLOW,
            AppTag.CONTAINER: Colors.RED,
            AppTag.FONT: Colors.WHITE,
        }

        formatted_tags = []
        for tag in tags:
            color = tag_colors.get(tag, Colors.DIM)
            formatted_tags.append(f"{color}[{tag.value}]{Colors.RESET}")

        return " ".join(formatted_tags)

    def check_app_status(self, app: AppInfo) -> tuple[AppStatus, Optional[str]]:
        """Check if an application is installed and get its version."""
        import os

        # Handle custom checks
        if app.custom_check == "font":
            return self._check_font_installed()

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

    def _check_font_installed(self) -> tuple[AppStatus, Optional[str]]:
        """Check if any Nerd Font is installed."""
        import os
        import platform

        system = platform.system().lower()

        if system == "darwin":
            # macOS: Check via Homebrew cask
            brew_path = shutil.which("brew")
            if brew_path:
                try:
                    result = subprocess.run(
                        [brew_path, "list", "--cask"],
                        capture_output=True,
                        text=True,
                        timeout=10
                    )
                    if result.returncode == 0:
                        casks = result.stdout.lower()
                        if "nerd-font" in casks:
                            # Find which font is installed
                            for line in result.stdout.split('\n'):
                                if "nerd-font" in line.lower():
                                    return AppStatus.INSTALLED, line.strip()
                except (subprocess.TimeoutExpired, Exception):
                    pass

            # Also check in ~/Library/Fonts
            fonts_dir = os.path.expanduser("~/Library/Fonts")
            if os.path.exists(fonts_dir):
                for filename in os.listdir(fonts_dir):
                    if "nerd" in filename.lower() or "meslo" in filename.lower():
                        return AppStatus.INSTALLED, "MesloLG Nerd Font"

        else:
            # Linux: Check in ~/.local/share/fonts
            fonts_dir = os.path.expanduser("~/.local/share/fonts")
            if os.path.exists(fonts_dir):
                for filename in os.listdir(fonts_dir):
                    if "nerd" in filename.lower() or "meslo" in filename.lower():
                        return AppStatus.INSTALLED, "MesloLG Nerd Font"

        return AppStatus.NOT_INSTALLED, None

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

            # Format tags
            tags_str = self.format_tags(app.tags)

            self.print(f"  {status_icon} {Colors.BOLD}{app.name}{Colors.RESET} {tags_str}")
            self.print(f"      {Colors.DIM}{app.description}{Colors.RESET}")
            self.print(f"      Status: {status_text}")
            self.print()

        return apps_status

    def ask_yes_no(self, question: str, default: bool = True) -> bool:
        """Ask a yes/no question and return the answer."""
        # In no-interaction mode, always return default
        if self.no_interaction:
            self.print_info(f"{question} → {'oui' if default else 'non'} (auto)")
            return default

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
        # In no-interaction mode, install all apps that are not installed
        if self.no_interaction:
            not_installed = [app for app, status, _ in apps_status if status == AppStatus.NOT_INSTALLED]
            self.print_section("Mode non-interactif")
            self.print_info(f"Installation automatique de {len(not_installed)} application(s) non installée(s)")
            for app in not_installed:
                self.print(f"  {Colors.CYAN}•{Colors.RESET} {app.name}")
            return not_installed

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

            tags_str = self.format_tags(app.tags)
            self.print(f"  [{i}] {marker} {app.name} {tags_str}{status_hint}")
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
                installer = NeovimInstallerApp(dry_run=self.dry_run, no_interaction=self.no_interaction, mode="install")
                return installer.run() == 0
            elif app.id == "neovim-config":
                from nvim_installer.app import NeovimInstallerApp
                installer = NeovimInstallerApp(dry_run=self.dry_run, no_interaction=self.no_interaction, mode="config")
                return installer.run() == 0
            elif app.id == "docker":
                from docker_installer.app import DockerInstallerApp
                installer = DockerInstallerApp(dry_run=self.dry_run, no_interaction=self.no_interaction)
                return installer.run() == 0
            elif app.id == "vscode":
                from vscode_installer.app import VSCodeInstallerApp
                installer = VSCodeInstallerApp(dry_run=self.dry_run, no_interaction=self.no_interaction)
                return installer.run() == 0
            elif app.id == "zsh":
                from zsh_installer.app import ZshInstallerApp
                installer = ZshInstallerApp(dry_run=self.dry_run, no_interaction=self.no_interaction, mode="zsh")
                return installer.run() == 0
            elif app.id == "oh-my-zsh":
                from zsh_installer.app import ZshInstallerApp
                installer = ZshInstallerApp(dry_run=self.dry_run, no_interaction=self.no_interaction, mode="oh-my-zsh")
                return installer.run() == 0
            elif app.id == "alias":
                from alias_installer.app import AliasInstallerApp
                installer = AliasInstallerApp(dry_run=self.dry_run, no_interaction=self.no_interaction)
                return installer.run() == 0
            elif app.id == "nerd-font":
                from font_installer.app import FontInstallerApp
                installer = FontInstallerApp(dry_run=self.dry_run, no_interaction=self.no_interaction)
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

    def run_uninstaller(self, app: AppInfo) -> bool:
        """Run the uninstaller for a specific application."""
        self.print()
        self.print(f"{Colors.RED}{'═' * 60}{Colors.RESET}")
        self.print(f"{Colors.RED}{Colors.BOLD}Désinstallation de {app.name}{Colors.RESET}")
        self.print(f"{Colors.RED}{'═' * 60}{Colors.RESET}")
        self.print()

        try:
            if app.id == "neovim" or app.id == "neovim-config":
                from nvim_installer.utils.system import SystemInfo, OSType
                from nvim_installer.utils.cli import CLI
                from nvim_installer.utils.runner import CommandRunner
                from nvim_installer.uninstallers.base import UninstallOptions

                system_info = SystemInfo.detect()
                cli = CLI(no_interaction=self.no_interaction)
                runner = CommandRunner(cli, dry_run=self.dry_run)

                if system_info.os_type == OSType.MACOS:
                    from nvim_installer.uninstallers.macos import MacOSUninstaller
                    uninstaller = MacOSUninstaller(system_info, cli, runner)
                else:
                    from nvim_installer.uninstallers.ubuntu import UbuntuUninstaller
                    uninstaller = UbuntuUninstaller(system_info, cli, runner)

                options = UninstallOptions(
                    remove_config=(app.id == "neovim-config" or app.id == "neovim"),
                    remove_cache=True,
                    remove_data=True
                )
                result = uninstaller.uninstall(options)
                return result.success

            elif app.id == "docker":
                from docker_installer.utils.system import SystemInfo, OSType
                from docker_installer.utils.cli import CLI
                from docker_installer.utils.runner import CommandRunner
                from docker_installer.uninstallers.base import UninstallOptions

                system_info = SystemInfo.detect()
                cli = CLI(no_interaction=self.no_interaction)
                runner = CommandRunner(cli, dry_run=self.dry_run)

                if system_info.os_type == OSType.MACOS:
                    from docker_installer.uninstallers.macos import MacOSUninstaller
                    uninstaller = MacOSUninstaller(system_info, cli, runner)
                else:
                    from docker_installer.uninstallers.ubuntu import UbuntuUninstaller
                    uninstaller = UbuntuUninstaller(system_info, cli, runner)

                options = UninstallOptions()
                result = uninstaller.uninstall(options)
                return result.success

            elif app.id == "vscode":
                from vscode_installer.utils.system import SystemInfo, OSType
                from vscode_installer.utils.cli import CLI
                from vscode_installer.utils.runner import CommandRunner
                from vscode_installer.uninstallers.base import UninstallOptions

                system_info = SystemInfo.detect()
                cli = CLI(no_interaction=self.no_interaction)
                runner = CommandRunner(cli, dry_run=self.dry_run)

                if system_info.os_type == OSType.MACOS:
                    from vscode_installer.uninstallers.macos import MacOSUninstaller
                    uninstaller = MacOSUninstaller(system_info, cli, runner)
                else:
                    from vscode_installer.uninstallers.ubuntu import UbuntuUninstaller
                    uninstaller = UbuntuUninstaller(system_info, cli, runner)

                options = UninstallOptions()
                result = uninstaller.uninstall(options)
                return result.success

            elif app.id == "zsh" or app.id == "oh-my-zsh":
                from zsh_installer.utils.system import SystemInfo, OSType
                from zsh_installer.utils.cli import CLI
                from zsh_installer.utils.runner import CommandRunner
                from zsh_installer.uninstallers.base import UninstallOptions

                system_info = SystemInfo.detect()
                cli = CLI(no_interaction=self.no_interaction)
                runner = CommandRunner(cli, dry_run=self.dry_run)

                if system_info.os_type == OSType.MACOS:
                    from zsh_installer.uninstallers.macos import MacOSUninstaller
                    uninstaller = MacOSUninstaller(system_info, cli, runner)
                else:
                    from zsh_installer.uninstallers.ubuntu import UbuntuUninstaller
                    uninstaller = UbuntuUninstaller(system_info, cli, runner)

                options = UninstallOptions(
                    remove_oh_my_zsh=True,
                    remove_plugins=True,
                    remove_zshrc=(app.id == "oh-my-zsh")
                )
                result = uninstaller.uninstall(options)
                return result.success

            elif app.id == "alias":
                from alias_installer.uninstaller import AliasUninstallerApp
                uninstaller = AliasUninstallerApp(dry_run=self.dry_run, no_interaction=self.no_interaction)
                return uninstaller.run() == 0

            elif app.id == "nerd-font":
                from font_installer.utils.system import SystemInfo, OSType
                from font_installer.utils.cli import CLI
                from font_installer.utils.runner import CommandRunner
                from font_installer.uninstallers.base import UninstallOptions

                system_info = SystemInfo.detect()
                cli = CLI(no_interaction=self.no_interaction)
                runner = CommandRunner(cli, dry_run=self.dry_run)

                if system_info.os_type == OSType.MACOS:
                    from font_installer.uninstallers.macos import MacOSUninstaller
                    uninstaller = MacOSUninstaller(system_info, cli, runner)
                else:
                    from font_installer.uninstallers.ubuntu import UbuntuUninstaller
                    uninstaller = UbuntuUninstaller(system_info, cli, runner)

                options = UninstallOptions(remove_all=True)
                result = uninstaller.uninstall(options)
                return result.success

            else:
                self.print_error(f"Désinstallateur inconnu pour {app.name}")
                return False

        except ImportError as e:
            self.print_error(f"Impossible de charger le désinstallateur: {e}")
            return False
        except Exception as e:
            self.print_error(f"Erreur lors de la désinstallation: {e}")
            import traceback
            traceback.print_exc()
            return False

    def ask_mode(self) -> str:
        """Ask user to select mode: install or uninstall."""
        if self.no_interaction:
            return self.mode

        self.print_section("Mode d'opération")
        self.print()
        self.print(f"  [1] {Colors.GREEN}Installer{Colors.RESET} - Installer de nouvelles applications")
        self.print(f"  [2] {Colors.RED}Désinstaller{Colors.RESET} - Supprimer des applications installées")
        self.print()

        while True:
            try:
                response = input(f"{Colors.YELLOW}?{Colors.RESET} Votre choix [1]: ").strip()

                if not response or response == "1":
                    return "install"
                elif response == "2":
                    return "uninstall"
                else:
                    self.print_warning("Entrez 1 pour installer ou 2 pour désinstaller")

            except (EOFError, KeyboardInterrupt):
                self.print()
                raise

    def ask_multi_select_uninstall(
        self,
        question: str,
        apps_status: list[tuple[AppInfo, AppStatus, Optional[str]]]
    ) -> list[AppInfo]:
        """Ask user to select multiple applications to uninstall."""
        # Filter only installed apps
        installed_apps = [(app, status, version) for app, status, version in apps_status
                         if status == AppStatus.INSTALLED]

        if not installed_apps:
            self.print_warning("Aucune application installée à désinstaller")
            return []

        if self.no_interaction:
            self.print_section("Mode non-interactif")
            self.print_warning("Désinstallation automatique non supportée en mode non-interactif")
            return []

        self.print_section("Sélection des désinstallations")
        self.print()
        self.print(f"{Colors.YELLOW}?{Colors.RESET} {question}")
        self.print(f"  {Colors.DIM}(Entrez les numéros séparés par des espaces){Colors.RESET}")
        self.print()

        for i, (app, status, version) in enumerate(installed_apps, 1):
            tags_str = self.format_tags(app.tags)
            version_str = f" {Colors.DIM}({version}){Colors.RESET}" if version else ""
            self.print(f"  [{i}] {Colors.GREEN}✓{Colors.RESET} {app.name} {tags_str}{version_str}")

        self.print()
        prompt = f"  Votre choix: "

        while True:
            try:
                response = input(prompt).strip().lower()

                if not response:
                    return []

                if response in ("tous", "all", "*"):
                    return [app for app, _, _ in installed_apps]

                if response in ("aucun", "none", "0"):
                    return []

                try:
                    indices = [int(x) for x in response.split()]
                    selected = []
                    for idx in indices:
                        if 1 <= idx <= len(installed_apps):
                            selected.append(installed_apps[idx - 1][0])
                        else:
                            raise ValueError(f"Index {idx} invalide")
                    return selected
                except ValueError as e:
                    self.print_warning(f"Entrée invalide: {e}")

            except (EOFError, KeyboardInterrupt):
                self.print()
                raise

    def show_uninstall_summary(self, selected_apps: list[AppInfo]) -> bool:
        """Show a summary of what will be uninstalled and ask for confirmation."""
        self.print_section("Résumé de la désinstallation")
        self.print()

        if not selected_apps:
            self.print_info("Aucune application sélectionnée pour la désinstallation.")
            return False

        self.print(f"{Colors.BOLD}{Colors.RED}Applications à désinstaller:{Colors.RESET}")
        for app in selected_apps:
            self.print(f"  {Colors.RED}•{Colors.RESET} {app.name}")

        self.print()
        self.print_warning("Cette action supprimera les applications et leurs configurations!")
        return self.ask_yes_no("Êtes-vous sûr de vouloir continuer?", default=False)

    def run(self) -> int:
        """Run the complete bootstrap process."""
        try:
            # Print banner
            self.print_banner()

            # Display available apps with status
            apps_status = self.display_available_apps()

            # Ask for mode (install/uninstall)
            mode = self.ask_mode()

            if mode == "uninstall":
                return self._run_uninstall_mode(apps_status)
            else:
                return self._run_install_mode(apps_status)

        except KeyboardInterrupt:
            self.print()
            self.print_warning("Opération interrompue par l'utilisateur")
            return 130

        except Exception as e:
            self.print_error(f"Erreur inattendue: {e}")
            if self.dry_run:
                import traceback
                traceback.print_exc()
            return 1

    def _run_install_mode(self, apps_status: list) -> int:
        """Run the installation mode."""
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

    def _run_uninstall_mode(self, apps_status: list) -> int:
        """Run the uninstallation mode."""
        # Ask user to select apps to uninstall
        selected_apps = self.ask_multi_select_uninstall(
            "Quelles applications souhaitez-vous désinstaller?",
            apps_status
        )

        # Show summary and confirm
        if not self.show_uninstall_summary(selected_apps):
            self.print_info("Désinstallation annulée")
            return 0

        # Run uninstallations
        results = []
        for app in selected_apps:
            success = self.run_uninstaller(app)
            results.append((app, success))

        # Show final summary
        self.print()
        self.print_section("Résumé final")
        self.print()

        success_count = sum(1 for _, success in results if success)
        failure_count = len(results) - success_count

        for app, success in results:
            if success:
                self.print_success(f"{app.name} désinstallé avec succès")
            else:
                self.print_error(f"{app.name} - échec de la désinstallation")

        self.print()
        if failure_count == 0:
            self.print_success(f"Toutes les désinstallations terminées avec succès! ({success_count}/{len(results)})")
        else:
            self.print_warning(f"Désinstallations terminées: {success_count} succès, {failure_count} échec(s)")

        return 0 if failure_count == 0 else 1


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
        "-n", "--no-interaction",
        action="store_true",
        help="Mode non-interactif (installe tout sans confirmation)"
    )
    parser.add_argument(
        "--uninstall", "-u",
        action="store_true",
        help="Mode désinstallation (supprimer les applications)"
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"DevBootstrap {BootstrapApp.VERSION}"
    )

    args = parser.parse_args()

    mode = "uninstall" if args.uninstall else "install"
    app = BootstrapApp(dry_run=args.dry_run, no_interaction=args.no_interaction, mode=mode)
    sys.exit(app.run())


if __name__ == "__main__":
    main()
