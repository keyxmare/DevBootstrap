"""Base installer class for Neovim installation."""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Optional
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


@dataclass
class Dependency:
    """Represents a dependency that needs to be installed."""
    name: str
    description: str
    check_command: str
    required: bool = True
    version_arg: str = "--version"


@dataclass
class InstallOptions:
    """Options for the installation process."""
    config_dir: str = ""
    install_dependencies: bool = True
    install_config: bool = True
    backup_existing: bool = True
    neovim_version: str = "stable"  # "stable", "nightly", or specific version


@dataclass
class InstallResult:
    """Result of an installation."""
    success: bool
    message: str
    neovim_path: Optional[str] = None
    neovim_version: Optional[str] = None
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class BaseInstaller(ABC):
    """Abstract base class for platform-specific installers."""

    # Dependencies that should be installed for a good Neovim experience
    COMMON_DEPENDENCIES: list[Dependency] = [
        Dependency(
            name="git",
            description="Système de contrôle de version",
            check_command="git"
        ),
        Dependency(
            name="node",
            description="Runtime JavaScript (pour LSP, plugins)",
            check_command="node",
            required=False
        ),
        Dependency(
            name="npm",
            description="Gestionnaire de paquets Node.js",
            check_command="npm",
            required=False
        ),
        Dependency(
            name="python3",
            description="Python 3 (pour plugins Python)",
            check_command="python3"
        ),
        Dependency(
            name="pip3",
            description="Gestionnaire de paquets Python",
            check_command="pip3",
            required=False
        ),
        Dependency(
            name="ripgrep",
            description="Recherche ultra-rapide (pour Telescope)",
            check_command="rg"
        ),
        Dependency(
            name="fd",
            description="Alternative à find (pour Telescope)",
            check_command="fd",
            required=False
        ),
        Dependency(
            name="fzf",
            description="Fuzzy finder",
            check_command="fzf",
            required=False
        ),
        Dependency(
            name="lazygit",
            description="Interface Git en terminal",
            check_command="lazygit",
            required=False
        ),
    ]

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the installer."""
        self.system_info = system_info
        self.cli = cli
        self.runner = runner
        self.options: Optional[InstallOptions] = None

    @abstractmethod
    def get_dependencies(self) -> list[Dependency]:
        """Get the list of dependencies for this platform."""
        pass

    @abstractmethod
    def install_package_manager(self) -> bool:
        """Install the package manager if not present."""
        pass

    @abstractmethod
    def update_package_manager(self) -> bool:
        """Update the package manager."""
        pass

    @abstractmethod
    def install_package(self, package: str, description: Optional[str] = None) -> bool:
        """Install a package using the platform's package manager."""
        pass

    @abstractmethod
    def install_neovim(self) -> bool:
        """Install Neovim."""
        pass

    def check_dependency(self, dep: Dependency) -> bool:
        """Check if a dependency is installed."""
        return self.runner.check_command_exists(dep.check_command)

    def check_all_dependencies(self) -> dict[str, bool]:
        """Check all dependencies and return their status."""
        results = {}
        for dep in self.get_dependencies():
            results[dep.name] = self.check_dependency(dep)
        return results

    def install_missing_dependencies(self) -> bool:
        """Install all missing dependencies."""
        dependencies = self.get_dependencies()
        total = len(dependencies)
        success = True

        for i, dep in enumerate(dependencies, 1):
            if self.check_dependency(dep):
                self.cli.print_success(f"{dep.name} - déjà installé")
                continue

            self.cli.print_step(i, total, f"Installation de {dep.name} ({dep.description})")

            if not self.install_package(dep.name, dep.description):
                if dep.required:
                    self.cli.print_error(f"Échec de l'installation de {dep.name}")
                    success = False
                else:
                    self.cli.print_warning(f"Impossible d'installer {dep.name} (optionnel)")
            else:
                self.cli.print_success(f"{dep.name} installé")

        return success

    def backup_existing_config(self) -> bool:
        """Backup existing Neovim configuration."""
        import os
        import datetime

        config_dir = self.system_info.get_config_dir()
        if not os.path.exists(config_dir):
            return True

        timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_dir = f"{config_dir}.backup.{timestamp}"

        self.cli.print_info(f"Sauvegarde de la configuration existante vers {backup_dir}")
        return self.runner.copy_directory(config_dir, backup_dir)

    def verify_installation(self) -> Optional[str]:
        """Verify Neovim installation and return version."""
        version = self.runner.get_command_version("nvim")
        if version:
            return version
        return None

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete installation process."""
        self.options = options
        errors = []
        warnings = []

        # Step 1: Check/install package manager
        self.cli.print_section("Vérification du gestionnaire de paquets")
        if not self.install_package_manager():
            return InstallResult(
                success=False,
                message="Impossible d'installer le gestionnaire de paquets",
                errors=["Package manager installation failed"]
            )

        # Step 2: Update package manager
        self.cli.print_section("Mise à jour du gestionnaire de paquets")
        if not self.update_package_manager():
            warnings.append("Impossible de mettre à jour le gestionnaire de paquets")
            self.cli.print_warning("Impossible de mettre à jour le gestionnaire de paquets")

        # Step 3: Install dependencies
        if options.install_dependencies:
            self.cli.print_section("Installation des dépendances")
            if not self.install_missing_dependencies():
                warnings.append("Certaines dépendances n'ont pas pu être installées")

        # Step 4: Backup existing config
        if options.backup_existing and options.install_config:
            self.cli.print_section("Sauvegarde de la configuration existante")
            if not self.backup_existing_config():
                warnings.append("Impossible de sauvegarder la configuration existante")

        # Step 5: Install Neovim
        self.cli.print_section("Installation de Neovim")
        if not self.install_neovim():
            return InstallResult(
                success=False,
                message="Échec de l'installation de Neovim",
                errors=["Neovim installation failed"]
            )

        # Step 6: Verify installation
        self.cli.print_section("Vérification de l'installation")
        version = self.verify_installation()
        if not version:
            return InstallResult(
                success=False,
                message="Neovim installé mais non trouvé dans le PATH",
                errors=["Neovim not found in PATH after installation"]
            )

        nvim_path = self.runner.get_command_path("nvim")

        return InstallResult(
            success=True,
            message="Installation terminée avec succès!",
            neovim_path=nvim_path,
            neovim_version=version,
            warnings=warnings
        )
