"""Ubuntu/Debian-specific VS Code installer using official Microsoft repository."""

import os
from typing import Optional
from .base import BaseInstaller, InstallOptions, InstallResult
from ..utils.system import SystemInfo, OSType, Architecture
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class UbuntuInstaller(BaseInstaller):
    """Installer for VS Code on Ubuntu/Debian using official Microsoft repository."""

    # Microsoft GPG key URL
    MS_GPG_URL = "https://packages.microsoft.com/keys/microsoft.asc"

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the Ubuntu VS Code installer."""
        super().__init__(system_info, cli, runner)

    def _get_arch_string(self) -> str:
        """Get the architecture string for the repository."""
        if self.system_info.architecture == Architecture.ARM64:
            return "arm64"
        return "amd64"

    def _install_prerequisites(self) -> bool:
        """Install prerequisites for VS Code installation."""
        self.cli.print_info("Installation des prerequis...")

        packages = [
            "wget",
            "gpg",
            "apt-transport-https"
        ]

        # Update apt
        result = self.runner.run(
            ["apt-get", "update"],
            description="Mise a jour des paquets",
            sudo=True,
            timeout=300
        )

        if not result.success:
            self.cli.print_warning("Impossible de mettre a jour apt")

        # Install packages
        result = self.runner.run(
            ["apt-get", "install", "-y"] + packages,
            description="Installation des prerequis",
            sudo=True,
            timeout=300
        )

        return result.success

    def _setup_microsoft_repository(self) -> bool:
        """Set up the official Microsoft repository."""
        self.cli.print_info("Configuration du repository Microsoft...")

        # Create keyrings directory
        keyrings_dir = "/etc/apt/keyrings"
        self.runner.run(
            ["mkdir", "-p", keyrings_dir],
            sudo=True
        )

        # Download and add Microsoft GPG key
        gpg_file = f"{keyrings_dir}/packages.microsoft.gpg"

        # Remove old key if exists
        if os.path.exists(gpg_file):
            self.runner.run(["rm", "-f", gpg_file], sudo=True)

        # Download GPG key
        result = self.runner.run(
            ["wget", "-qO-", self.MS_GPG_URL],
            description="Telechargement de la cle GPG Microsoft",
            sudo=False
        )

        if not result.success:
            self.cli.print_error("Impossible de telecharger la cle GPG Microsoft")
            return False

        # Dearmor and install GPG key
        result = self.runner.run(
            ["bash", "-c", f"wget -qO- {self.MS_GPG_URL} | gpg --dearmor > /tmp/packages.microsoft.gpg"],
            sudo=False
        )

        if result.success:
            self.runner.run(
                ["mv", "/tmp/packages.microsoft.gpg", gpg_file],
                sudo=True
            )

        # Set proper permissions
        self.runner.run(
            ["chmod", "a+r", gpg_file],
            sudo=True
        )

        # Get architecture
        arch = self._get_arch_string()

        # Add Microsoft repository
        repo_line = f"deb [arch={arch} signed-by={gpg_file}] https://packages.microsoft.com/repos/code stable main"

        result = self.runner.run(
            ["bash", "-c", f'echo "{repo_line}" > /etc/apt/sources.list.d/vscode.list'],
            sudo=True
        )

        if not result.success:
            self.cli.print_error("Impossible d'ajouter le repository Microsoft")
            return False

        # Update apt with new repository
        result = self.runner.run(
            ["apt-get", "update"],
            description="Mise a jour avec le nouveau repository",
            sudo=True,
            timeout=300
        )

        return result.success

    def check_existing_installation(self) -> bool:
        """Check if VS Code is already installed."""
        return self.runner.check_command_exists("code")

    def install_vscode(self) -> bool:
        """Install VS Code from official Microsoft repository."""
        # Install prerequisites
        if not self._install_prerequisites():
            self.cli.print_error("Impossible d'installer les prerequis")
            return False

        # Setup Microsoft repository
        if not self._setup_microsoft_repository():
            self.cli.print_error("Impossible de configurer le repository Microsoft")
            return False

        # Install VS Code
        self.cli.print_info("Installation de VS Code...")

        result = self.runner.run(
            ["apt-get", "install", "-y", "code"],
            description="Installation de VS Code",
            sudo=True,
            timeout=600
        )

        if not result.success:
            self.cli.print_error("Echec de l'installation de VS Code")
            return False

        self.cli.print_success("VS Code installe")
        return True

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete VS Code installation for Ubuntu/Debian."""
        result = super().install(options)

        if result.success:
            self.cli.print_section("Informations supplementaires")
            self.cli.print_info("Pour lancer VS Code depuis le terminal:")
            self.cli.print("  code .              # Ouvrir le dossier courant")
            self.cli.print("  code fichier.txt    # Ouvrir un fichier")
            self.cli.print()
            self.cli.print_info("Pour mettre a jour VS Code:")
            self.cli.print("  sudo apt update && sudo apt upgrade code")

        return result
