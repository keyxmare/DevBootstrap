"""Main application class for Docker Installer."""

import sys
from typing import Optional

from .utils.system import SystemInfo, OSType
from .utils.cli import CLI, Colors
from .utils.runner import CommandRunner
from .installers.base import InstallOptions, InstallResult
from .installers.macos import MacOSInstaller
from .installers.ubuntu import UbuntuInstaller


class DockerInstallerApp:
    """Main application for installing Docker."""

    VERSION = "1.0.0"

    def __init__(self, dry_run: bool = False):
        """Initialize the application."""
        self.cli = CLI()
        self.system_info = SystemInfo.detect()
        self.runner = CommandRunner(self.cli, dry_run=dry_run)
        self.installer = self._get_installer()
        self.dry_run = dry_run

    def _get_installer(self):
        """Get the appropriate installer for the current platform."""
        if self.system_info.os_type == OSType.MACOS:
            return MacOSInstaller(self.system_info, self.cli, self.runner)
        elif self.system_info.os_type in (OSType.UBUNTU, OSType.DEBIAN):
            return UbuntuInstaller(self.system_info, self.cli, self.runner)
        return None

    def print_banner(self):
        """Print the application banner."""
        self.cli.print_header("Docker Installer v" + self.VERSION)

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

        # Check for existing Docker installation
        existing_docker = self.runner.get_command_path("docker")
        if existing_docker:
            version = self.runner.get_command_version("docker")
            self.cli.print_info(f"Docker existant detecte: {version}")
            self.cli.print_info(f"Chemin: {existing_docker}")

        # Check for sudo on Linux
        if self.system_info.os_type in (OSType.UBUNTU, OSType.DEBIAN):
            if not self.system_info.is_root:
                if not self.runner.check_command_exists("sudo"):
                    self.cli.print_error("sudo est requis pour installer Docker sur Linux")
                    return False
                self.cli.print_success("sudo est disponible")

        return True

    def get_install_options(self) -> Optional[InstallOptions]:
        """Interactively get installation options from user."""
        self.cli.print_section("Configuration de l'installation")

        # Ask about Docker Compose
        install_compose = self.cli.ask_yes_no(
            "Installer Docker Compose?",
            default=True
        )

        # Platform-specific options
        add_to_group = True
        start_on_boot = True

        if self.system_info.os_type in (OSType.UBUNTU, OSType.DEBIAN):
            add_to_group = self.cli.ask_yes_no(
                "Ajouter l'utilisateur au groupe docker (pour utiliser Docker sans sudo)?",
                default=True
            )

            start_on_boot = self.cli.ask_yes_no(
                "Demarrer Docker automatiquement au boot?",
                default=True
            )

        return InstallOptions(
            install_compose=install_compose,
            install_buildx=True,
            add_user_to_docker_group=add_to_group,
            start_on_boot=start_on_boot
        )

    def run_installation(self, options: InstallOptions) -> bool:
        """Run the Docker installation."""
        result = self.installer.install(options)

        if result.success:
            self.cli.print_section("Installation terminee")
            self.cli.print_success(result.message)

            if result.docker_path:
                self.cli.print_info(f"Chemin Docker: {result.docker_path}")
            if result.docker_version:
                self.cli.print_info(f"Version Docker: {result.docker_version}")
            if result.compose_version:
                self.cli.print_info(f"Version Compose: {result.compose_version}")

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

    def test_installation(self) -> bool:
        """Test Docker installation with hello-world."""
        self.cli.print_section("Test de l'installation")

        if self.installer.test_docker():
            self.cli.print_success("Docker fonctionne correctement!")
            return True
        else:
            self.cli.print_warning("Le test Docker a echoue")
            self.cli.print_info("Cela peut etre normal si vous devez vous reconnecter")
            return False

    def show_final_instructions(self):
        """Show final instructions after installation."""
        self.cli.print_section("Prochaines etapes")

        self.cli.print()
        self.cli.print(f"{Colors.BOLD}1. Verification{Colors.RESET}")
        self.cli.print("   $ docker --version")
        self.cli.print("   $ docker compose version")
        self.cli.print()

        self.cli.print(f"{Colors.BOLD}2. Test{Colors.RESET}")
        self.cli.print("   $ docker run hello-world")
        self.cli.print()

        if self.system_info.os_type in (OSType.UBUNTU, OSType.DEBIAN):
            self.cli.print(f"{Colors.BOLD}3. Important (Linux){Colors.RESET}")
            self.cli.print("   Deconnectez-vous et reconnectez-vous pour")
            self.cli.print("   utiliser Docker sans sudo")
            self.cli.print()
            self.cli.print("   Ou executez: newgrp docker")
            self.cli.print()

        self.cli.print(f"{Colors.BOLD}Commandes utiles:{Colors.RESET}")
        self.cli.print("   docker ps              - Lister les conteneurs")
        self.cli.print("   docker images          - Lister les images")
        self.cli.print("   docker compose up -d   - Demarrer un projet")
        self.cli.print("   docker compose down    - Arreter un projet")
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
            if not self.cli.ask_yes_no("Proceder a l'installation de Docker?", default=True):
                self.cli.print_info("Installation annulee")
                return 0

            # Get install options
            install_options = self.get_install_options()
            if not install_options:
                return 1

            # Summary
            self.cli.show_summary("Resume de l'installation", {
                "Docker Compose": "Oui" if install_options.install_compose else "Non",
                "Groupe docker": "Oui" if install_options.add_user_to_docker_group else "Non",
                "Demarrage auto": "Oui" if install_options.start_on_boot else "Non",
            })

            if not self.cli.ask_yes_no("Confirmer et lancer l'installation?", default=True):
                self.cli.print_info("Installation annulee")
                return 0

            # Run installation
            if not self.run_installation(install_options):
                return 1

            # Test installation (optional)
            if self.cli.ask_yes_no("Tester l'installation avec hello-world?", default=True):
                self.test_installation()

            # Show final instructions
            self.show_final_instructions()

            self.cli.print_success("Installation de Docker terminee avec succes!")
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
        description="Docker Installer - Installation automatique de Docker"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simuler l'installation sans effectuer de changements"
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"Docker Installer {DockerInstallerApp.VERSION}"
    )

    args = parser.parse_args()

    app = DockerInstallerApp(dry_run=args.dry_run)
    sys.exit(app.run())


if __name__ == "__main__":
    main()
