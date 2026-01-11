"""macOS-specific Docker installer using Docker Desktop."""

import os
import time
from typing import Optional
from .base import BaseInstaller, InstallOptions, InstallResult
from ..utils.system import SystemInfo, Architecture
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class MacOSInstaller(BaseInstaller):
    """Installer for Docker on macOS using Docker Desktop."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the macOS Docker installer."""
        super().__init__(system_info, cli, runner)
        self._homebrew_path: Optional[str] = None

    def _get_homebrew_path(self) -> Optional[str]:
        """Get the Homebrew installation path."""
        if self._homebrew_path:
            return self._homebrew_path

        # Check common Homebrew locations
        possible_paths = [
            "/opt/homebrew/bin/brew",  # Apple Silicon
            "/usr/local/bin/brew",      # Intel
        ]

        for path in possible_paths:
            if os.path.exists(path):
                self._homebrew_path = path
                return path

        # Check if brew is in PATH
        brew_path = self.runner.get_command_path("brew")
        if brew_path:
            self._homebrew_path = brew_path
            return brew_path

        return None

    def _ensure_homebrew_in_path(self) -> bool:
        """Ensure Homebrew is in the PATH for this session."""
        brew_path = self._get_homebrew_path()
        if not brew_path:
            return False

        brew_dir = os.path.dirname(brew_path)
        current_path = os.environ.get("PATH", "")

        if brew_dir not in current_path:
            os.environ["PATH"] = f"{brew_dir}:{current_path}"

        return True

    def _install_homebrew(self) -> bool:
        """Install Homebrew if not present."""
        if self._get_homebrew_path():
            self._ensure_homebrew_in_path()
            self.cli.print_success("Homebrew est deja installe")
            return True

        self.cli.print_info("Installation de Homebrew...")

        # Homebrew installation script
        install_script = '/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'

        result = self.runner.run_interactive(
            ["/bin/bash", "-c", install_script],
            description="Installation de Homebrew",
            sudo=False
        )

        if not result:
            return False

        self._ensure_homebrew_in_path()
        return self._get_homebrew_path() is not None

    def check_existing_installation(self) -> bool:
        """Check if Docker is already installed."""
        # Check for docker command
        if self.runner.check_command_exists("docker"):
            # Verify docker daemon is accessible
            result = self.runner.run(
                ["docker", "info"],
                sudo=False,
                timeout=10
            )
            if result.success:
                return True

        # Check if Docker Desktop app exists
        docker_app_paths = [
            "/Applications/Docker.app",
            os.path.expanduser("~/Applications/Docker.app")
        ]

        for path in docker_app_paths:
            if os.path.exists(path):
                return True

        return False

    def install_docker(self) -> bool:
        """Install Docker Desktop using Homebrew Cask."""
        # Ensure Homebrew is available
        if not self._install_homebrew():
            self.cli.print_error("Homebrew est requis pour installer Docker Desktop")
            return False

        brew_path = self._get_homebrew_path()

        # Check if Docker Desktop is already installed via brew
        result = self.runner.run(
            [brew_path, "list", "--cask", "docker"],
            sudo=False
        )

        if result.success:
            self.cli.print_info("Docker Desktop est deja installe via Homebrew")
            # Try to upgrade
            result = self.runner.run(
                [brew_path, "upgrade", "--cask", "docker"],
                description="Mise a jour de Docker Desktop",
                sudo=False,
                timeout=600
            )
            return True

        # Install Docker Desktop
        self.cli.print_info("Installation de Docker Desktop via Homebrew...")
        result = self.runner.run(
            [brew_path, "install", "--cask", "docker"],
            description="Installation de Docker Desktop",
            sudo=False,
            timeout=900  # Docker Desktop is large
        )

        if not result.success:
            self.cli.print_error("Echec de l'installation de Docker Desktop")
            self.cli.print_info("Vous pouvez telecharger Docker Desktop manuellement:")
            self.cli.print_info("https://www.docker.com/products/docker-desktop/")
            return False

        self.cli.print_success("Docker Desktop installe")
        return True

    def install_docker_compose(self) -> bool:
        """Docker Compose is included with Docker Desktop."""
        self.cli.print_info("Docker Compose est inclus avec Docker Desktop")
        return True

    def configure_docker(self) -> bool:
        """Configure Docker on macOS."""
        # Docker Desktop handles most configuration automatically
        self.cli.print_info("Docker Desktop gere la configuration automatiquement")

        # Add docker CLI to PATH if needed
        docker_cli_path = "/usr/local/bin/docker"
        if not os.path.exists(docker_cli_path):
            # Docker Desktop symlinks are usually created when the app starts
            self.cli.print_info("Les liens symboliques seront crees au demarrage de Docker Desktop")

        return True

    def start_docker(self) -> bool:
        """Start Docker Desktop application."""
        # Check if Docker is already running
        result = self.runner.run(
            ["docker", "info"],
            sudo=False,
            timeout=10
        )

        if result.success:
            self.cli.print_success("Docker est deja en cours d'execution")
            return True

        # Start Docker Desktop
        self.cli.print_info("Demarrage de Docker Desktop...")

        result = self.runner.run(
            ["open", "-a", "Docker"],
            description="Demarrage de Docker Desktop",
            sudo=False
        )

        if not result.success:
            self.cli.print_warning("Impossible de demarrer Docker Desktop automatiquement")
            self.cli.print_info("Veuillez demarrer Docker Desktop manuellement depuis Applications")
            return False

        # Wait for Docker to be ready
        self.cli.print_info("Attente du demarrage de Docker...")
        max_attempts = 30
        for i in range(max_attempts):
            time.sleep(2)
            result = self.runner.run(
                ["docker", "info"],
                sudo=False,
                timeout=5
            )
            if result.success:
                self.cli.print_success("Docker est pret!")
                return True

            if i < max_attempts - 1:
                self.cli.print_progress(f"Attente... ({i + 1}/{max_attempts})")

        self.cli.print_warning("Docker prend du temps a demarrer")
        self.cli.print_info("Veuillez attendre que l'icone Docker dans la barre de menu indique 'Running'")
        return False

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete Docker installation for macOS."""
        result = super().install(options)

        if result.success:
            # Additional macOS-specific info
            self.cli.print_section("Informations supplementaires")
            self.cli.print_info("Docker Desktop inclut:")
            self.cli.print("  - Docker Engine")
            self.cli.print("  - Docker CLI")
            self.cli.print("  - Docker Compose")
            self.cli.print("  - Docker BuildKit")
            self.cli.print("  - Kubernetes (optionnel)")
            self.cli.print()
            self.cli.print_info("Pour activer Kubernetes:")
            self.cli.print("  Docker Desktop > Preferences > Kubernetes > Enable Kubernetes")

        return result
