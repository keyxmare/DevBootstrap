"""Ubuntu/Debian-specific Docker installer using official Docker repository."""

import os
from typing import Optional
from .base import BaseInstaller, InstallOptions, InstallResult
from ..utils.system import SystemInfo, OSType, Architecture
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class UbuntuInstaller(BaseInstaller):
    """Installer for Docker on Ubuntu/Debian using official Docker repository."""

    # Docker GPG key URL
    DOCKER_GPG_URL = "https://download.docker.com/linux/ubuntu/gpg"
    DOCKER_GPG_URL_DEBIAN = "https://download.docker.com/linux/debian/gpg"

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the Ubuntu Docker installer."""
        super().__init__(system_info, cli, runner)

    def _get_distro_codename(self) -> str:
        """Get the distribution codename for Docker repository."""
        result = self.runner.run(
            ["lsb_release", "-cs"],
            sudo=False
        )
        if result.success:
            return result.stdout.strip()

        # Fallback: try to read from /etc/os-release
        try:
            with open("/etc/os-release", "r") as f:
                for line in f:
                    if line.startswith("VERSION_CODENAME="):
                        return line.split("=")[1].strip().strip('"')
        except FileNotFoundError:
            pass

        # Default fallback
        return "jammy"  # Ubuntu 22.04

    def _get_arch_string(self) -> str:
        """Get the architecture string for Docker repository."""
        if self.system_info.architecture == Architecture.ARM64:
            return "arm64"
        return "amd64"

    def _remove_old_packages(self) -> bool:
        """Remove old Docker packages that might conflict."""
        old_packages = [
            "docker.io",
            "docker-doc",
            "docker-compose",
            "docker-compose-v2",
            "podman-docker",
            "containerd",
            "runc"
        ]

        self.cli.print_info("Suppression des anciens paquets Docker...")

        for package in old_packages:
            # Check if package is installed
            result = self.runner.run(
                ["dpkg", "-s", package],
                sudo=False
            )
            if result.success:
                self.runner.run(
                    ["apt-get", "remove", "-y", package],
                    sudo=True
                )

        return True

    def _install_prerequisites(self) -> bool:
        """Install prerequisites for Docker installation."""
        self.cli.print_info("Installation des prerequis...")

        packages = [
            "ca-certificates",
            "curl",
            "gnupg",
            "lsb-release"
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

    def _setup_docker_repository(self) -> bool:
        """Set up the official Docker repository."""
        self.cli.print_info("Configuration du repository Docker...")

        # Determine if Ubuntu or Debian
        is_debian = self.system_info.os_type == OSType.DEBIAN
        distro = "debian" if is_debian else "ubuntu"
        gpg_url = self.DOCKER_GPG_URL_DEBIAN if is_debian else self.DOCKER_GPG_URL

        # Create keyrings directory
        keyrings_dir = "/etc/apt/keyrings"
        self.runner.run(
            ["mkdir", "-p", keyrings_dir],
            sudo=True
        )

        # Download and add Docker GPG key
        gpg_file = f"{keyrings_dir}/docker.gpg"

        # Remove old key if exists
        if os.path.exists(gpg_file):
            self.runner.run(["rm", "-f", gpg_file], sudo=True)

        # Download GPG key
        result = self.runner.run(
            ["curl", "-fsSL", gpg_url, "-o", "/tmp/docker.gpg.tmp"],
            description="Telechargement de la cle GPG Docker",
            sudo=False
        )

        if not result.success:
            self.cli.print_error("Impossible de telecharger la cle GPG Docker")
            return False

        # Dearmor and install GPG key
        result = self.runner.run(
            ["gpg", "--dearmor", "-o", gpg_file, "/tmp/docker.gpg.tmp"],
            sudo=True
        )

        if not result.success:
            # Try alternative method
            result = self.runner.run(
                ["bash", "-c", f"cat /tmp/docker.gpg.tmp | gpg --dearmor -o {gpg_file}"],
                sudo=True
            )

        # Set proper permissions
        self.runner.run(
            ["chmod", "a+r", gpg_file],
            sudo=True
        )

        # Get architecture and codename
        arch = self._get_arch_string()
        codename = self._get_distro_codename()

        # Add Docker repository
        repo_line = f"deb [arch={arch} signed-by={gpg_file}] https://download.docker.com/linux/{distro} {codename} stable"

        result = self.runner.run(
            ["bash", "-c", f'echo "{repo_line}" > /etc/apt/sources.list.d/docker.list'],
            sudo=True
        )

        if not result.success:
            self.cli.print_error("Impossible d'ajouter le repository Docker")
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
        """Check if Docker is already installed."""
        if self.runner.check_command_exists("docker"):
            # Verify docker daemon is accessible
            result = self.runner.run(
                ["docker", "info"],
                sudo=True,
                timeout=10
            )
            return result.success

        return False

    def install_docker(self) -> bool:
        """Install Docker Engine from official repository."""
        # Remove old packages
        self._remove_old_packages()

        # Install prerequisites
        if not self._install_prerequisites():
            self.cli.print_error("Impossible d'installer les prerequis")
            return False

        # Setup Docker repository
        if not self._setup_docker_repository():
            self.cli.print_error("Impossible de configurer le repository Docker")
            return False

        # Install Docker packages
        self.cli.print_info("Installation de Docker Engine...")

        docker_packages = [
            "docker-ce",
            "docker-ce-cli",
            "containerd.io",
            "docker-buildx-plugin",
            "docker-compose-plugin"
        ]

        result = self.runner.run(
            ["apt-get", "install", "-y"] + docker_packages,
            description="Installation de Docker",
            sudo=True,
            timeout=600
        )

        if not result.success:
            self.cli.print_error("Echec de l'installation de Docker")
            return False

        self.cli.print_success("Docker Engine installe")
        return True

    def install_docker_compose(self) -> bool:
        """Docker Compose v2 is installed as a plugin with Docker Engine."""
        # Check if already installed
        result = self.runner.run(
            ["docker", "compose", "version"],
            sudo=False
        )

        if result.success:
            self.cli.print_success("Docker Compose est deja installe")
            return True

        # If not installed, try to install the plugin separately
        result = self.runner.run(
            ["apt-get", "install", "-y", "docker-compose-plugin"],
            description="Installation de Docker Compose",
            sudo=True,
            timeout=300
        )

        return result.success

    def configure_docker(self) -> bool:
        """Configure Docker for the current user."""
        success = True

        if self.options and self.options.add_user_to_docker_group:
            # Add current user to docker group
            username = os.environ.get("USER") or os.environ.get("LOGNAME")

            if username and username != "root":
                self.cli.print_info(f"Ajout de l'utilisateur '{username}' au groupe docker...")

                # Check if docker group exists
                result = self.runner.run(
                    ["getent", "group", "docker"],
                    sudo=False
                )

                if not result.success:
                    # Create docker group
                    self.runner.run(
                        ["groupadd", "docker"],
                        sudo=True
                    )

                # Add user to group
                result = self.runner.run(
                    ["usermod", "-aG", "docker", username],
                    sudo=True
                )

                if result.success:
                    self.cli.print_success(f"Utilisateur '{username}' ajoute au groupe docker")
                    self.cli.print_warning("Deconnectez-vous et reconnectez-vous pour appliquer les changements")
                else:
                    self.cli.print_warning("Impossible d'ajouter l'utilisateur au groupe docker")
                    success = False

        if self.options and self.options.start_on_boot:
            # Enable Docker to start on boot
            self.cli.print_info("Configuration du demarrage automatique...")

            result = self.runner.run(
                ["systemctl", "enable", "docker.service"],
                sudo=True
            )

            if result.success:
                self.cli.print_success("Docker demarre automatiquement au boot")
            else:
                self.cli.print_warning("Impossible de configurer le demarrage automatique")

            # Also enable containerd
            self.runner.run(
                ["systemctl", "enable", "containerd.service"],
                sudo=True
            )

        return success

    def start_docker(self) -> bool:
        """Start Docker service."""
        # Check if already running
        result = self.runner.run(
            ["systemctl", "is-active", "docker"],
            sudo=False
        )

        if result.success and "active" in result.stdout:
            self.cli.print_success("Docker est deja en cours d'execution")
            return True

        # Start Docker service
        self.cli.print_info("Demarrage du service Docker...")

        result = self.runner.run(
            ["systemctl", "start", "docker"],
            sudo=True
        )

        if result.success:
            self.cli.print_success("Docker demarre")
            return True
        else:
            self.cli.print_error("Impossible de demarrer Docker")
            return False

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete Docker installation for Ubuntu/Debian."""
        result = super().install(options)

        if result.success and options.add_user_to_docker_group:
            # Remind user about group changes
            result.warnings.append(
                "Deconnectez-vous et reconnectez-vous pour utiliser Docker sans sudo"
            )

        return result
