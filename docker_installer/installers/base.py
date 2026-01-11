"""Base installer class for Docker installation."""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Optional
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


@dataclass
class InstallOptions:
    """Options for the Docker installation process."""
    install_compose: bool = True
    install_buildx: bool = True
    add_user_to_docker_group: bool = True
    start_on_boot: bool = True


@dataclass
class InstallResult:
    """Result of a Docker installation."""
    success: bool
    message: str
    docker_path: Optional[str] = None
    docker_version: Optional[str] = None
    compose_version: Optional[str] = None
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class BaseInstaller(ABC):
    """Abstract base class for platform-specific Docker installers."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the installer."""
        self.system_info = system_info
        self.cli = cli
        self.runner = runner
        self.options: Optional[InstallOptions] = None

    @abstractmethod
    def check_existing_installation(self) -> bool:
        """Check if Docker is already installed."""
        pass

    @abstractmethod
    def install_docker(self) -> bool:
        """Install Docker Engine."""
        pass

    @abstractmethod
    def install_docker_compose(self) -> bool:
        """Install Docker Compose."""
        pass

    @abstractmethod
    def configure_docker(self) -> bool:
        """Configure Docker (user groups, startup, etc.)."""
        pass

    @abstractmethod
    def start_docker(self) -> bool:
        """Start Docker service."""
        pass

    def verify_installation(self) -> tuple[Optional[str], Optional[str]]:
        """Verify Docker installation and return versions."""
        docker_version = self.runner.get_command_version("docker")
        compose_version = None

        # Check for docker compose (v2) or docker-compose (v1)
        result = self.runner.run(
            ["docker", "compose", "version"],
            sudo=False
        )
        if result.success:
            compose_version = result.stdout.strip().split("\n")[0]
        else:
            compose_version = self.runner.get_command_version("docker-compose")

        return docker_version, compose_version

    def test_docker(self) -> bool:
        """Test Docker by running hello-world container."""
        self.cli.print_info("Test de Docker avec hello-world...")

        result = self.runner.run(
            ["docker", "run", "--rm", "hello-world"],
            description="Test Docker hello-world",
            sudo=False,
            timeout=120
        )

        return result.success

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete Docker installation process."""
        self.options = options
        errors = []
        warnings = []

        # Step 1: Check existing installation
        self.cli.print_section("Verification de l'installation existante")
        if self.check_existing_installation():
            docker_version, compose_version = self.verify_installation()
            self.cli.print_info(f"Docker est deja installe: {docker_version}")

            if not self.cli.ask_yes_no("Voulez-vous reinstaller/mettre a jour Docker?", default=False):
                return InstallResult(
                    success=True,
                    message="Docker est deja installe",
                    docker_path=self.runner.get_command_path("docker"),
                    docker_version=docker_version,
                    compose_version=compose_version
                )

        # Step 2: Install Docker
        self.cli.print_section("Installation de Docker")
        if not self.install_docker():
            return InstallResult(
                success=False,
                message="Echec de l'installation de Docker",
                errors=["Docker installation failed"]
            )

        # Step 3: Install Docker Compose if requested
        if options.install_compose:
            self.cli.print_section("Installation de Docker Compose")
            if not self.install_docker_compose():
                warnings.append("Docker Compose n'a pas pu etre installe")
                self.cli.print_warning("Docker Compose n'a pas pu etre installe")

        # Step 4: Configure Docker
        self.cli.print_section("Configuration de Docker")
        if not self.configure_docker():
            warnings.append("Configuration partielle de Docker")

        # Step 5: Start Docker
        self.cli.print_section("Demarrage de Docker")
        if not self.start_docker():
            warnings.append("Docker n'a pas pu etre demarre automatiquement")
            self.cli.print_warning("Docker n'a pas pu etre demarre automatiquement")

        # Step 6: Verify installation
        self.cli.print_section("Verification de l'installation")
        docker_version, compose_version = self.verify_installation()

        if not docker_version:
            return InstallResult(
                success=False,
                message="Docker installe mais non trouve dans le PATH",
                errors=["Docker not found in PATH after installation"]
            )

        docker_path = self.runner.get_command_path("docker")

        return InstallResult(
            success=True,
            message="Installation de Docker terminee avec succes!",
            docker_path=docker_path,
            docker_version=docker_version,
            compose_version=compose_version,
            warnings=warnings
        )
