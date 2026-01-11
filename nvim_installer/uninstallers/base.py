"""Base uninstaller class for Neovim uninstallation."""

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
    """Options for the uninstallation process."""
    remove_config: bool = True
    remove_cache: bool = True
    remove_data: bool = True
    backup_before_remove: bool = True


@dataclass
class UninstallResult:
    """Result of an uninstallation."""
    success: bool
    message: str
    removed_paths: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class BaseUninstaller(ABC):
    """Abstract base class for platform-specific uninstallers."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the uninstaller."""
        self.system_info = system_info
        self.cli = cli
        self.runner = runner
        self.options: Optional[UninstallOptions] = None

    def get_neovim_paths(self) -> dict[str, str]:
        """Get all Neovim-related paths."""
        home = self.system_info.home_dir
        return {
            "config": os.path.join(home, ".config", "nvim"),
            "data": os.path.join(home, ".local", "share", "nvim"),
            "cache": os.path.join(home, ".cache", "nvim"),
            "state": os.path.join(home, ".local", "state", "nvim"),
        }

    def check_installed(self) -> bool:
        """Check if Neovim is installed."""
        return self.runner.check_command_exists("nvim")

    def get_installed_version(self) -> Optional[str]:
        """Get the installed Neovim version."""
        return self.runner.get_command_version("nvim")

    def backup_config(self) -> Optional[str]:
        """Backup Neovim configuration before removal."""
        import datetime

        config_dir = self.get_neovim_paths()["config"]
        if not os.path.exists(config_dir):
            return None

        timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_dir = f"{config_dir}.backup.{timestamp}"

        self.cli.print_info(f"Sauvegarde de la configuration vers {backup_dir}")
        try:
            shutil.copytree(config_dir, backup_dir)
            return backup_dir
        except Exception as e:
            self.cli.print_warning(f"Impossible de sauvegarder: {e}")
            return None

    def remove_directory(self, path: str, description: str) -> bool:
        """Remove a directory if it exists."""
        if not os.path.exists(path):
            return True

        self.cli.print_info(f"Suppression de {description}: {path}")
        try:
            shutil.rmtree(path)
            self.cli.print_success(f"{description} supprime")
            return True
        except Exception as e:
            self.cli.print_error(f"Erreur lors de la suppression: {e}")
            return False

    @abstractmethod
    def uninstall_neovim(self) -> bool:
        """Uninstall Neovim using the platform's package manager."""
        pass

    @abstractmethod
    def uninstall_dependencies(self) -> bool:
        """Optionally uninstall dependencies installed with Neovim."""
        pass

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process."""
        self.options = options
        errors = []
        warnings = []
        removed_paths = []

        # Step 0: Check if installed
        self.cli.print_section("Verification de l'installation")
        if not self.check_installed():
            return UninstallResult(
                success=True,
                message="Neovim n'est pas installe",
                warnings=["Neovim n'etait pas installe"]
            )

        version = self.get_installed_version()
        self.cli.print_info(f"Version installee: {version}")

        # Step 1: Backup config if requested
        if options.backup_before_remove and options.remove_config:
            self.cli.print_section("Sauvegarde de la configuration")
            backup_path = self.backup_config()
            if backup_path:
                self.cli.print_success(f"Configuration sauvegardee: {backup_path}")

        # Step 2: Uninstall Neovim
        self.cli.print_section("Desinstallation de Neovim")
        if not self.uninstall_neovim():
            errors.append("Echec de la desinstallation de Neovim")
            return UninstallResult(
                success=False,
                message="Echec de la desinstallation de Neovim",
                errors=errors
            )
        removed_paths.append("neovim")

        # Step 3: Remove config
        if options.remove_config:
            self.cli.print_section("Suppression de la configuration")
            config_path = self.get_neovim_paths()["config"]
            if self.remove_directory(config_path, "Configuration"):
                removed_paths.append(config_path)
            else:
                warnings.append(f"Impossible de supprimer {config_path}")

        # Step 4: Remove data
        if options.remove_data:
            self.cli.print_section("Suppression des donnees")
            data_path = self.get_neovim_paths()["data"]
            state_path = self.get_neovim_paths()["state"]

            if self.remove_directory(data_path, "Donnees"):
                removed_paths.append(data_path)
            else:
                warnings.append(f"Impossible de supprimer {data_path}")

            if self.remove_directory(state_path, "Etat"):
                removed_paths.append(state_path)

        # Step 5: Remove cache
        if options.remove_cache:
            self.cli.print_section("Suppression du cache")
            cache_path = self.get_neovim_paths()["cache"]
            if self.remove_directory(cache_path, "Cache"):
                removed_paths.append(cache_path)
            else:
                warnings.append(f"Impossible de supprimer {cache_path}")

        # Verify uninstallation
        self.cli.print_section("Verification de la desinstallation")
        if self.check_installed():
            warnings.append("Neovim semble toujours present dans le PATH")
            self.cli.print_warning("Neovim semble toujours present")
        else:
            self.cli.print_success("Neovim desinstalle avec succes")

        return UninstallResult(
            success=True,
            message="Desinstallation terminee avec succes!",
            removed_paths=removed_paths,
            warnings=warnings
        )
