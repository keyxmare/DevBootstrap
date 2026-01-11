"""Base uninstaller class for Docker uninstallation."""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Optional
import os
import shutil
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


@dataclass
class UninstallOptions:
    """Options for the Docker uninstallation process."""
    remove_images: bool = True
    remove_containers: bool = True
    remove_volumes: bool = False  # Default False - volumes contain data
    remove_config: bool = True
    stop_containers: bool = True


@dataclass
class UninstallResult:
    """Result of a Docker uninstallation."""
    success: bool
    message: str
    removed_items: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class BaseUninstaller(ABC):
    """Abstract base class for platform-specific Docker uninstallers."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the uninstaller."""
        self.system_info = system_info
        self.cli = cli
        self.runner = runner
        self.options: Optional[UninstallOptions] = None

    def check_installed(self) -> bool:
        """Check if Docker is installed."""
        return self.runner.check_command_exists("docker")

    def get_installed_version(self) -> Optional[str]:
        """Get the installed Docker version."""
        return self.runner.get_command_version("docker")

    def stop_all_containers(self) -> bool:
        """Stop all running containers."""
        self.cli.print_info("Arret de tous les conteneurs en cours d'execution...")

        # Get running containers
        result = self.runner.run(
            ["docker", "ps", "-q"],
            sudo=False
        )

        if not result.success or not result.stdout.strip():
            self.cli.print_info("Aucun conteneur en cours d'execution")
            return True

        container_ids = result.stdout.strip().split("\n")
        if container_ids and container_ids[0]:
            result = self.runner.run(
                ["docker", "stop"] + container_ids,
                description="Arret des conteneurs",
                sudo=False,
                timeout=120
            )

            if result.success:
                self.cli.print_success(f"{len(container_ids)} conteneur(s) arrete(s)")
            else:
                self.cli.print_warning("Certains conteneurs n'ont pas pu etre arretes")

        return True

    def remove_all_containers(self) -> bool:
        """Remove all containers."""
        self.cli.print_info("Suppression de tous les conteneurs...")

        # Get all containers
        result = self.runner.run(
            ["docker", "ps", "-aq"],
            sudo=False
        )

        if not result.success or not result.stdout.strip():
            self.cli.print_info("Aucun conteneur a supprimer")
            return True

        container_ids = result.stdout.strip().split("\n")
        if container_ids and container_ids[0]:
            result = self.runner.run(
                ["docker", "rm", "-f"] + container_ids,
                description="Suppression des conteneurs",
                sudo=False,
                timeout=120
            )

            if result.success:
                self.cli.print_success(f"{len(container_ids)} conteneur(s) supprime(s)")
            else:
                self.cli.print_warning("Certains conteneurs n'ont pas pu etre supprimes")

        return True

    def remove_all_images(self) -> bool:
        """Remove all Docker images."""
        self.cli.print_info("Suppression de toutes les images...")

        # Get all images
        result = self.runner.run(
            ["docker", "images", "-q"],
            sudo=False
        )

        if not result.success or not result.stdout.strip():
            self.cli.print_info("Aucune image a supprimer")
            return True

        image_ids = list(set(result.stdout.strip().split("\n")))  # Remove duplicates
        if image_ids and image_ids[0]:
            result = self.runner.run(
                ["docker", "rmi", "-f"] + image_ids,
                description="Suppression des images",
                sudo=False,
                timeout=300
            )

            if result.success:
                self.cli.print_success(f"{len(image_ids)} image(s) supprimee(s)")
            else:
                self.cli.print_warning("Certaines images n'ont pas pu etre supprimees")

        return True

    def remove_all_volumes(self) -> bool:
        """Remove all Docker volumes."""
        self.cli.print_info("Suppression de tous les volumes...")

        # Get all volumes
        result = self.runner.run(
            ["docker", "volume", "ls", "-q"],
            sudo=False
        )

        if not result.success or not result.stdout.strip():
            self.cli.print_info("Aucun volume a supprimer")
            return True

        volume_names = result.stdout.strip().split("\n")
        if volume_names and volume_names[0]:
            result = self.runner.run(
                ["docker", "volume", "rm", "-f"] + volume_names,
                description="Suppression des volumes",
                sudo=False,
                timeout=120
            )

            if result.success:
                self.cli.print_success(f"{len(volume_names)} volume(s) supprime(s)")
            else:
                self.cli.print_warning("Certains volumes n'ont pas pu etre supprimes")

        return True

    def prune_system(self) -> bool:
        """Run docker system prune to clean up."""
        self.cli.print_info("Nettoyage du systeme Docker...")

        result = self.runner.run(
            ["docker", "system", "prune", "-af"],
            description="Nettoyage Docker",
            sudo=False,
            timeout=300
        )

        return result.success

    @abstractmethod
    def uninstall_docker(self) -> bool:
        """Uninstall Docker using the platform's package manager."""
        pass

    @abstractmethod
    def remove_docker_data(self) -> bool:
        """Remove Docker data directories."""
        pass

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process."""
        self.options = options
        errors = []
        warnings = []
        removed_items = []

        # Step 0: Check if installed
        self.cli.print_section("Verification de l'installation")
        if not self.check_installed():
            return UninstallResult(
                success=True,
                message="Docker n'est pas installe",
                warnings=["Docker n'etait pas installe"]
            )

        version = self.get_installed_version()
        self.cli.print_info(f"Version installee: {version}")

        # Step 1: Stop containers
        if options.stop_containers:
            self.cli.print_section("Arret des conteneurs")
            self.stop_all_containers()

        # Step 2: Remove containers
        if options.remove_containers:
            self.cli.print_section("Suppression des conteneurs")
            if self.remove_all_containers():
                removed_items.append("containers")

        # Step 3: Remove images
        if options.remove_images:
            self.cli.print_section("Suppression des images")
            if self.remove_all_images():
                removed_items.append("images")

        # Step 4: Remove volumes (optional, data loss warning)
        if options.remove_volumes:
            self.cli.print_section("Suppression des volumes")
            if self.remove_all_volumes():
                removed_items.append("volumes")
        else:
            warnings.append("Les volumes Docker n'ont pas ete supprimes (contiennent des donnees)")

        # Step 5: Prune system
        self.cli.print_section("Nettoyage du systeme")
        self.prune_system()

        # Step 6: Uninstall Docker
        self.cli.print_section("Desinstallation de Docker")
        if not self.uninstall_docker():
            errors.append("Echec de la desinstallation de Docker")
            return UninstallResult(
                success=False,
                message="Echec de la desinstallation de Docker",
                errors=errors
            )
        removed_items.append("docker")

        # Step 7: Remove data directories
        if options.remove_config:
            self.cli.print_section("Suppression des donnees")
            if self.remove_docker_data():
                removed_items.append("data")
            else:
                warnings.append("Certaines donnees n'ont pas pu etre supprimees")

        # Verify uninstallation
        self.cli.print_section("Verification de la desinstallation")
        if self.check_installed():
            warnings.append("Docker semble toujours present dans le PATH")
            self.cli.print_warning("Docker semble toujours present")
        else:
            self.cli.print_success("Docker desinstalle avec succes")

        return UninstallResult(
            success=True,
            message="Desinstallation de Docker terminee avec succes!",
            removed_items=removed_items,
            warnings=warnings
        )
