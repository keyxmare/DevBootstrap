"""Main application class for VS Code Installer."""

import sys
from typing import Optional

from .utils.system import SystemInfo, OSType
from .utils.cli import CLI, Colors
from .utils.runner import CommandRunner
from .installers.base import InstallOptions, InstallResult
from .installers.macos import MacOSInstaller
from .installers.ubuntu import UbuntuInstaller


# Default extensions to suggest
DEFAULT_EXTENSIONS = [
    ("ms-python.python", "Python"),
    ("esbenp.prettier-vscode", "Prettier - Code formatter"),
    ("dbaeumer.vscode-eslint", "ESLint"),
    ("ms-vscode.vscode-typescript-next", "TypeScript"),
    ("bradlc.vscode-tailwindcss", "Tailwind CSS IntelliSense"),
    ("eamodio.gitlens", "GitLens"),
    ("pkief.material-icon-theme", "Material Icon Theme"),
]


class VSCodeInstallerApp:
    """Main application for installing VS Code."""

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
        self.cli.print_header("VS Code Installer v" + self.VERSION)

    def print_system_info(self):
        """Print detected system information."""
        self.cli.show_summary("Informations systeme", {
            "OS": str(self.system_info),
            "Architecture": self.system_info.architecture.value,
            "Repertoire home": self.system_info.home_dir,
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

        # Check for existing VS Code installation
        existing_vscode = self.runner.get_command_path("code")
        if existing_vscode:
            version = self.runner.get_command_version("code")
            self.cli.print_info(f"VS Code existant detecte: {version}")
            self.cli.print_info(f"Chemin: {existing_vscode}")

        return True

    def get_install_options(self) -> Optional[InstallOptions]:
        """Interactively get installation options from user."""
        self.cli.print_section("Configuration de l'installation")

        # Ask about extensions
        install_extensions = self.cli.ask_yes_no(
            "Installer des extensions recommandees?",
            default=True
        )

        selected_extensions = []
        if install_extensions:
            self.cli.print()
            self.cli.print("Extensions disponibles:")
            for i, (ext_id, ext_name) in enumerate(DEFAULT_EXTENSIONS, 1):
                self.cli.print(f"  [{i}] {ext_name} ({ext_id})")

            self.cli.print()
            self.cli.print_info("Entrez les numeros separes par des espaces, ou 'tous' pour tout installer")
            prompt = "  Votre choix [tous]: "

            try:
                response = input(prompt).strip().lower()

                if not response or response in ("tous", "all", "*"):
                    selected_extensions = [ext_id for ext_id, _ in DEFAULT_EXTENSIONS]
                elif response in ("aucun", "none", "0"):
                    selected_extensions = []
                else:
                    try:
                        indices = [int(x) for x in response.split()]
                        for idx in indices:
                            if 1 <= idx <= len(DEFAULT_EXTENSIONS):
                                selected_extensions.append(DEFAULT_EXTENSIONS[idx - 1][0])
                    except ValueError:
                        self.cli.print_warning("Entree invalide, installation de toutes les extensions")
                        selected_extensions = [ext_id for ext_id, _ in DEFAULT_EXTENSIONS]

            except (EOFError, KeyboardInterrupt):
                selected_extensions = []

        return InstallOptions(
            install_extensions=install_extensions,
            default_extensions=selected_extensions
        )

    def run_installation(self, options: InstallOptions) -> bool:
        """Run the VS Code installation."""
        result = self.installer.install(options)

        if result.success:
            self.cli.print_section("Installation terminee")
            self.cli.print_success(result.message)

            if result.vscode_path:
                self.cli.print_info(f"Chemin: {result.vscode_path}")
            if result.vscode_version:
                self.cli.print_info(f"Version: {result.vscode_version}")

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
        self.cli.print(f"{Colors.BOLD}1. Lancer VS Code{Colors.RESET}")
        self.cli.print("   $ code .")
        self.cli.print()

        self.cli.print(f"{Colors.BOLD}2. Raccourcis utiles{Colors.RESET}")
        self.cli.print("   Cmd/Ctrl+Shift+P  - Palette de commandes")
        self.cli.print("   Cmd/Ctrl+P        - Recherche rapide de fichiers")
        self.cli.print("   Cmd/Ctrl+,        - Parametres")
        self.cli.print()

        self.cli.print(f"{Colors.BOLD}3. Installer plus d'extensions{Colors.RESET}")
        self.cli.print("   code --install-extension <extension-id>")
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

            # Check if installer is available
            if not self.installer:
                self.cli.print_error("Aucun installateur disponible pour ce systeme")
                return 1

            # Confirm to proceed
            if not self.cli.ask_yes_no("Proceder a l'installation de VS Code?", default=True):
                self.cli.print_info("Installation annulee")
                return 0

            # Get install options
            install_options = self.get_install_options()
            if not install_options:
                return 1

            # Summary
            ext_count = len(install_options.default_extensions) if install_options.install_extensions else 0
            self.cli.show_summary("Resume de l'installation", {
                "Extensions": f"{ext_count} selectionnee(s)" if ext_count else "Aucune",
            })

            if not self.cli.ask_yes_no("Confirmer et lancer l'installation?", default=True):
                self.cli.print_info("Installation annulee")
                return 0

            # Run installation
            if not self.run_installation(install_options):
                return 1

            # Show final instructions
            self.show_final_instructions()

            self.cli.print_success("Installation de VS Code terminee avec succes!")
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
        description="VS Code Installer - Installation automatique de Visual Studio Code"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simuler l'installation sans effectuer de changements"
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"VS Code Installer {VSCodeInstallerApp.VERSION}"
    )

    args = parser.parse_args()

    app = VSCodeInstallerApp(dry_run=args.dry_run)
    sys.exit(app.run())


if __name__ == "__main__":
    main()
