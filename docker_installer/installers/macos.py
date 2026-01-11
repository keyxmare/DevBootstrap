"""macOS-specific Docker installer using Docker Desktop DMG."""

import os
import time
import tempfile
import subprocess
from typing import Optional
from .base import BaseInstaller, InstallOptions, InstallResult
from ..utils.system import SystemInfo, Architecture
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class MacOSInstaller(BaseInstaller):
    """Installer for Docker on macOS using Docker Desktop DMG."""

    # Docker Desktop download URLs
    DOCKER_DMG_URL_ARM64 = "https://desktop.docker.com/mac/main/arm64/Docker.dmg"
    DOCKER_DMG_URL_AMD64 = "https://desktop.docker.com/mac/main/amd64/Docker.dmg"

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the macOS Docker installer."""
        super().__init__(system_info, cli, runner)

    def _get_docker_dmg_url(self) -> str:
        """Get the appropriate Docker DMG URL for the current architecture."""
        if self.system_info.architecture == Architecture.ARM64:
            return self.DOCKER_DMG_URL_ARM64
        return self.DOCKER_DMG_URL_AMD64

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
        """Install Docker Desktop by downloading and installing the official DMG."""
        dmg_url = self._get_docker_dmg_url()
        arch_name = "Apple Silicon" if self.system_info.architecture == Architecture.ARM64 else "Intel"

        self.cli.print_info(f"Telechargement de Docker Desktop pour {arch_name}...")
        self.cli.print_info(f"URL: {dmg_url}")

        # Create temporary directory for download
        with tempfile.TemporaryDirectory() as tmp_dir:
            dmg_path = os.path.join(tmp_dir, "Docker.dmg")

            # Download the DMG
            self.cli.print_progress("Telechargement en cours...")
            result = self.runner.run(
                ["curl", "-fSL", "-o", dmg_path, dmg_url],
                description="Telechargement de Docker Desktop",
                sudo=False,
                timeout=600  # 10 minutes for download
            )

            if not result.success:
                self.cli.print_error("Echec du telechargement de Docker Desktop")
                self.cli.print_info("Vous pouvez telecharger manuellement depuis:")
                self.cli.print_info("https://www.docker.com/products/docker-desktop/")
                return False

            self.cli.print_success("Telechargement termine")

            # Mount the DMG
            self.cli.print_info("Montage du DMG...")
            mount_point = "/Volumes/Docker"

            # Unmount if already mounted
            if os.path.exists(mount_point):
                self.runner.run(["hdiutil", "detach", mount_point, "-quiet"], sudo=False)

            result = self.runner.run(
                ["hdiutil", "attach", dmg_path, "-nobrowse", "-quiet"],
                description="Montage du DMG",
                sudo=False,
                timeout=60
            )

            if not result.success:
                self.cli.print_error("Echec du montage du DMG")
                return False

            try:
                # Copy Docker.app to /Applications
                self.cli.print_info("Installation de Docker Desktop dans /Applications...")
                docker_app_src = "/Volumes/Docker/Docker.app"

                if not os.path.exists(docker_app_src):
                    self.cli.print_error("Docker.app non trouve dans le DMG")
                    return False

                # Remove existing installation if present
                docker_app_dest = "/Applications/Docker.app"
                if os.path.exists(docker_app_dest):
                    self.cli.print_info("Suppression de l'ancienne version...")
                    result = self.runner.run(
                        ["rm", "-rf", docker_app_dest],
                        sudo=False
                    )
                    if not result.success:
                        self.cli.print_error("Impossible de supprimer l'ancienne version")
                        return False

                # Copy the app
                result = self.runner.run(
                    ["cp", "-R", docker_app_src, docker_app_dest],
                    description="Copie de Docker.app",
                    sudo=False,
                    timeout=120
                )

                if not result.success:
                    self.cli.print_error("Echec de la copie de Docker.app")
                    return False

                self.cli.print_success("Docker Desktop installe dans /Applications")
                return True

            finally:
                # Always unmount the DMG
                self.cli.print_info("Demontage du DMG...")
                self.runner.run(
                    ["hdiutil", "detach", mount_point, "-quiet"],
                    sudo=False,
                    timeout=30
                )

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
