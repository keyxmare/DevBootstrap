"""Ubuntu/Debian-specific Docker uninstaller."""

import os
import shutil
from .base import BaseUninstaller, UninstallOptions, UninstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class UbuntuUninstaller(BaseUninstaller):
    """Uninstaller for Docker on Ubuntu/Debian."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the Ubuntu Docker uninstaller."""
        super().__init__(system_info, cli, runner)

    def _stop_docker_service(self) -> bool:
        """Stop Docker service."""
        self.cli.print_info("Arret du service Docker...")

        # Stop docker service
        self.runner.run(
            ["systemctl", "stop", "docker.service"],
            sudo=True
        )

        # Stop containerd
        self.runner.run(
            ["systemctl", "stop", "containerd.service"],
            sudo=True
        )

        # Disable services
        self.runner.run(
            ["systemctl", "disable", "docker.service"],
            sudo=True
        )
        self.runner.run(
            ["systemctl", "disable", "containerd.service"],
            sudo=True
        )

        return True

    def _remove_user_from_docker_group(self) -> bool:
        """Remove current user from docker group."""
        username = os.environ.get("USER") or os.environ.get("LOGNAME")

        if username and username != "root":
            self.cli.print_info(f"Suppression de l'utilisateur '{username}' du groupe docker...")

            result = self.runner.run(
                ["gpasswd", "-d", username, "docker"],
                sudo=True
            )

            if result.success:
                self.cli.print_success(f"Utilisateur retire du groupe docker")
            else:
                self.cli.print_info("L'utilisateur n'etait pas dans le groupe docker")

        return True

    def uninstall_docker(self) -> bool:
        """Uninstall Docker packages via apt."""
        # Stop services first
        self._stop_docker_service()

        # Remove user from docker group
        self._remove_user_from_docker_group()

        # Docker packages to remove
        docker_packages = [
            "docker-ce",
            "docker-ce-cli",
            "containerd.io",
            "docker-buildx-plugin",
            "docker-compose-plugin",
            "docker-ce-rootless-extras",
        ]

        self.cli.print_info("Desinstallation des paquets Docker...")

        # Remove packages
        result = self.runner.run(
            ["apt-get", "remove", "-y"] + docker_packages,
            description="Suppression des paquets Docker",
            sudo=True,
            timeout=300
        )

        if not result.success:
            self.cli.print_warning("Certains paquets n'ont pas pu etre supprimes")

        # Purge configuration
        result = self.runner.run(
            ["apt-get", "purge", "-y"] + docker_packages,
            description="Purge de la configuration",
            sudo=True,
            timeout=300
        )

        # Clean up
        self.runner.run(
            ["apt-get", "autoremove", "-y"],
            description="Nettoyage des paquets inutilises",
            sudo=True
        )

        # Remove Docker repository
        self._remove_docker_repository()

        return True

    def _remove_docker_repository(self) -> bool:
        """Remove Docker apt repository and GPG key."""
        self.cli.print_info("Suppression du repository Docker...")

        # Remove repository file
        repo_file = "/etc/apt/sources.list.d/docker.list"
        if os.path.exists(repo_file):
            try:
                os.remove(repo_file)
                self.cli.print_success("Repository Docker supprime")
            except PermissionError:
                self.runner.run(["rm", "-f", repo_file], sudo=True)

        # Remove GPG key
        gpg_file = "/etc/apt/keyrings/docker.gpg"
        if os.path.exists(gpg_file):
            try:
                os.remove(gpg_file)
                self.cli.print_success("Cle GPG Docker supprimee")
            except PermissionError:
                self.runner.run(["rm", "-f", gpg_file], sudo=True)

        # Update apt
        self.runner.run(
            ["apt-get", "update"],
            sudo=True
        )

        return True

    def remove_docker_data(self) -> bool:
        """Remove Docker data directories."""
        paths_to_remove = [
            "/var/lib/docker",
            "/var/lib/containerd",
            "/etc/docker",
            os.path.expanduser("~/.docker"),
        ]

        success = True
        for path in paths_to_remove:
            if os.path.exists(path):
                self.cli.print_info(f"Suppression de {path}...")
                try:
                    if os.path.isdir(path):
                        shutil.rmtree(path)
                    else:
                        os.remove(path)
                    self.cli.print_success(f"Supprime: {path}")
                except PermissionError:
                    # Try with sudo
                    result = self.runner.run(
                        ["rm", "-rf", path],
                        sudo=True
                    )
                    if result.success:
                        self.cli.print_success(f"Supprime: {path}")
                    else:
                        self.cli.print_warning(f"Impossible de supprimer {path}")
                        success = False
                except Exception as e:
                    self.cli.print_warning(f"Impossible de supprimer {path}: {e}")
                    success = False

        return success

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process for Ubuntu/Debian."""
        # Run base uninstallation
        result = super().uninstall(options)

        if result.success:
            self.cli.print_section("Nettoyage final")

            # Remove docker group (optional)
            self.runner.run(
                ["groupdel", "docker"],
                sudo=True
            )

            result.warnings.append(
                "Redemarrez le systeme pour completer la desinstallation"
            )

        return result
