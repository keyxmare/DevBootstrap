"""Main application class for Font Installer."""

import sys
from typing import Optional

from .utils.system import SystemInfo, OSType
from .utils.cli import CLI, Colors
from .utils.runner import CommandRunner
from .installers.base import InstallOptions, AVAILABLE_FONTS, FontInfo
from .installers.macos import MacOSInstaller
from .installers.ubuntu import UbuntuInstaller


class FontInstallerApp:
    """Main application for installing Nerd Fonts."""

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
        self.cli.print_header("Nerd Font Installer v" + self.VERSION)

    def print_system_info(self):
        """Print detected system information."""
        self.cli.show_summary("Informations systeme", {
            "OS": str(self.system_info),
            "Architecture": self.system_info.architecture.value,
            "Repertoire polices": self.system_info.get_fonts_dir(),
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

        # Check installer is available
        if not self.installer:
            self.cli.print_error("Aucun installateur disponible pour ce systeme")
            return False

        return True

    def show_available_fonts(self) -> list[tuple[FontInfo, bool]]:
        """Show available fonts and their installation status."""
        self.cli.print_section("Polices Nerd Font disponibles")
        self.cli.print()

        fonts_status = []
        for font in AVAILABLE_FONTS:
            is_installed = self.installer.check_font_installed(font)
            fonts_status.append((font, is_installed))

            if is_installed:
                status = f"{Colors.GREEN}[installee]{Colors.RESET}"
            else:
                status = f"{Colors.YELLOW}[non installee]{Colors.RESET}"

            self.cli.print(f"  {Colors.BOLD}{font.name}{Colors.RESET} {status}")
            self.cli.print(f"      {Colors.DIM}{font.description}{Colors.RESET}")
            self.cli.print()

        return fonts_status

    def get_install_options(self, fonts_status: list[tuple[FontInfo, bool]]) -> Optional[InstallOptions]:
        """Interactively get installation options from user."""
        self.cli.print_section("Selection des polices")

        # In no-interaction mode, install MesloLG (default for agnoster)
        if self.no_interaction:
            not_installed = [font for font, installed in fonts_status if not installed]
            if not not_installed:
                self.cli.print_info("Toutes les polices sont deja installees")
                return InstallOptions(fonts=[])

            # Install MesloLG by default (first in list, recommended for agnoster)
            meslo = next((f for f in not_installed if f.id == "meslo"), not_installed[0])
            self.cli.print_info(f"Installation automatique de {meslo.name}")
            return InstallOptions(fonts=[meslo])

        self.cli.print()
        self.cli.print("Entrez les numeros des polices a installer (separes par des espaces)")
        self.cli.print(f"{Colors.DIM}Appuyez sur Entree pour installer MesloLG (recommandee pour agnoster){Colors.RESET}")
        self.cli.print()

        for i, (font, is_installed) in enumerate(fonts_status, 1):
            if is_installed:
                status = f"{Colors.GREEN}[installee]{Colors.RESET}"
            else:
                status = ""
            self.cli.print(f"  [{i}] {font.name} {status}")

        self.cli.print()
        prompt = "  Votre choix [1]: "

        try:
            response = input(prompt).strip().lower()

            if not response:
                # Default: install MesloLG
                return InstallOptions(fonts=[AVAILABLE_FONTS[0]])

            if response in ("tous", "all", "*"):
                return InstallOptions(fonts=AVAILABLE_FONTS)

            if response in ("aucun", "none", "0"):
                return InstallOptions(fonts=[])

            # Parse space-separated numbers
            try:
                indices = [int(x) for x in response.split()]
                selected = []
                for idx in indices:
                    if 1 <= idx <= len(AVAILABLE_FONTS):
                        selected.append(AVAILABLE_FONTS[idx - 1])
                    else:
                        self.cli.print_warning(f"Index {idx} invalide, ignore")
                return InstallOptions(fonts=selected)
            except ValueError:
                self.cli.print_warning("Entree invalide, installation de MesloLG par defaut")
                return InstallOptions(fonts=[AVAILABLE_FONTS[0]])

        except (EOFError, KeyboardInterrupt):
            return None

    def run_installation(self, options: InstallOptions) -> bool:
        """Run the font installation."""
        result = self.installer.install(options)

        if result.success:
            self.cli.print_section("Installation terminee")
            self.cli.print_success(result.message)

            if result.installed_fonts:
                self.cli.print()
                self.cli.print("Polices installees:")
                for font_name in result.installed_fonts:
                    self.cli.print(f"  {Colors.GREEN}*{Colors.RESET} {font_name}")

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
        self.cli.print(f"{Colors.BOLD}1. Configurer le terminal{Colors.RESET}")
        self.cli.print("   Ouvrez les preferences de votre terminal et selectionnez")
        self.cli.print("   une police Nerd Font (ex: MesloLGS NF)")
        self.cli.print()

        if self.system_info.os_type == OSType.MACOS:
            self.cli.print(f"{Colors.BOLD}2. Terminal.app (macOS){Colors.RESET}")
            self.cli.print("   Preferences > Profils > Police > Changer...")
            self.cli.print("   Rechercher 'MesloLGS NF'")
            self.cli.print()

            self.cli.print(f"{Colors.BOLD}3. iTerm2{Colors.RESET}")
            self.cli.print("   Preferences > Profiles > Text > Font")
            self.cli.print("   Selectionner 'MesloLGS NF'")
            self.cli.print()

        self.cli.print(f"{Colors.BOLD}4. VS Code (terminal integre){Colors.RESET}")
        self.cli.print("   Settings > Terminal > Integrated: Font Family")
        self.cli.print("   Entrer: 'MesloLGS NF'")
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

            # Show available fonts
            fonts_status = self.show_available_fonts()

            # Check if all fonts are installed
            all_installed = all(installed for _, installed in fonts_status)
            if all_installed:
                self.cli.print_success("Toutes les polices Nerd Font sont deja installees!")
                if not self.cli.ask_yes_no("Voulez-vous quand meme reinstaller?", default=False):
                    return 0

            # Confirm to proceed
            if not self.cli.ask_yes_no("Proceder a l'installation des polices?", default=True):
                self.cli.print_info("Installation annulee")
                return 0

            # Get install options
            install_options = self.get_install_options(fonts_status)
            if install_options is None:
                self.cli.print_info("Installation annulee")
                return 0

            if not install_options.fonts:
                self.cli.print_info("Aucune police selectionnee")
                return 0

            # Summary
            self.cli.show_summary("Resume de l'installation", {
                "Polices": ", ".join(f.name for f in install_options.fonts),
                "Nombre": str(len(install_options.fonts)),
            })

            if not self.cli.ask_yes_no("Confirmer et lancer l'installation?", default=True):
                self.cli.print_info("Installation annulee")
                return 0

            # Run installation
            if not self.run_installation(install_options):
                return 1

            # Show final instructions
            self.show_final_instructions()

            self.cli.print_success("Installation des polices terminee avec succes!")
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
        description="Font Installer - Installation automatique des Nerd Fonts"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simuler l'installation sans effectuer de changements"
    )
    parser.add_argument(
        "-n", "--no-interaction",
        action="store_true",
        help="Mode non-interactif (installe MesloLG par defaut)"
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"Font Installer {FontInstallerApp.VERSION}"
    )

    args = parser.parse_args()

    app = FontInstallerApp(dry_run=args.dry_run, no_interaction=args.no_interaction)
    sys.exit(app.run())


if __name__ == "__main__":
    main()
