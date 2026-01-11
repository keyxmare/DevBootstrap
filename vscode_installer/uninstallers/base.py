"""Base uninstaller class for VS Code uninstallation."""

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
    """Options for the VS Code uninstallation process."""
    remove_extensions: bool = True
    remove_settings: bool = True
    remove_cache: bool = True
    backup_before_remove: bool = True


@dataclass
class UninstallResult:
    """Result of a VS Code uninstallation."""
    success: bool
    message: str
    removed_items: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class BaseUninstaller(ABC):
    """Abstract base class for platform-specific VS Code uninstallers."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the uninstaller."""
        self.system_info = system_info
        self.cli = cli
        self.runner = runner
        self.options: Optional[UninstallOptions] = None

    def check_installed(self) -> bool:
        """Check if VS Code is installed."""
        return self.runner.check_command_exists("code")

    def get_installed_version(self) -> Optional[str]:
        """Get the installed VS Code version."""
        return self.runner.get_command_version("code")

    @abstractmethod
    def get_vscode_paths(self) -> dict[str, str]:
        """Get VS Code related paths for this platform."""
        pass

    def backup_settings(self) -> Optional[str]:
        """Backup VS Code settings before removal."""
        import datetime

        paths = self.get_vscode_paths()
        settings_dir = paths.get("settings")

        if not settings_dir or not os.path.exists(settings_dir):
            return None

        timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_dir = f"{settings_dir}.backup.{timestamp}"

        self.cli.print_info(f"Sauvegarde des parametres vers {backup_dir}")
        try:
            shutil.copytree(settings_dir, backup_dir)
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

    def remove_extensions(self) -> bool:
        """Remove VS Code extensions directory."""
        paths = self.get_vscode_paths()
        extensions_dir = paths.get("extensions")

        if not extensions_dir:
            return True

        return self.remove_directory(extensions_dir, "Extensions")

    def remove_settings(self) -> bool:
        """Remove VS Code settings directory."""
        paths = self.get_vscode_paths()
        settings_dir = paths.get("settings")

        if not settings_dir:
            return True

        return self.remove_directory(settings_dir, "Parametres")

    def remove_cache(self) -> bool:
        """Remove VS Code cache directories."""
        paths = self.get_vscode_paths()
        cache_dirs = [
            paths.get("cache"),
            paths.get("crashpad"),
        ]

        success = True
        for cache_dir in cache_dirs:
            if cache_dir and os.path.exists(cache_dir):
                if not self.remove_directory(cache_dir, "Cache"):
                    success = False

        return success

    @abstractmethod
    def uninstall_vscode(self) -> bool:
        """Uninstall VS Code using the platform's package manager."""
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
                message="VS Code n'est pas installe",
                warnings=["VS Code n'etait pas installe"]
            )

        version = self.get_installed_version()
        self.cli.print_info(f"Version installee: {version}")

        # Step 1: Backup settings
        if options.backup_before_remove and options.remove_settings:
            self.cli.print_section("Sauvegarde des parametres")
            backup_path = self.backup_settings()
            if backup_path:
                self.cli.print_success(f"Parametres sauvegardes: {backup_path}")

        # Step 2: Uninstall VS Code
        self.cli.print_section("Desinstallation de VS Code")
        if not self.uninstall_vscode():
            errors.append("Echec de la desinstallation de VS Code")
            return UninstallResult(
                success=False,
                message="Echec de la desinstallation de VS Code",
                errors=errors
            )
        removed_items.append("vscode")

        # Step 3: Remove extensions
        if options.remove_extensions:
            self.cli.print_section("Suppression des extensions")
            if self.remove_extensions():
                removed_items.append("extensions")
            else:
                warnings.append("Extensions non supprimees")

        # Step 4: Remove settings
        if options.remove_settings:
            self.cli.print_section("Suppression des parametres")
            if self.remove_settings():
                removed_items.append("settings")
            else:
                warnings.append("Parametres non supprimes")

        # Step 5: Remove cache
        if options.remove_cache:
            self.cli.print_section("Suppression du cache")
            if self.remove_cache():
                removed_items.append("cache")
            else:
                warnings.append("Cache non supprime")

        # Verify uninstallation
        self.cli.print_section("Verification de la desinstallation")
        if self.check_installed():
            warnings.append("VS Code semble toujours present dans le PATH")
            self.cli.print_warning("VS Code semble toujours present")
        else:
            self.cli.print_success("VS Code desinstalle avec succes")

        return UninstallResult(
            success=True,
            message="Desinstallation de VS Code terminee avec succes!",
            removed_items=removed_items,
            warnings=warnings
        )
