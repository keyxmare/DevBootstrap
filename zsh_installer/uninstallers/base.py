"""Base uninstaller class for Zsh uninstallation."""

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
    """Options for the Zsh uninstallation process."""
    remove_oh_my_zsh: bool = True
    remove_plugins: bool = True
    remove_zshrc: bool = True
    restore_default_shell: bool = True
    backup_before_remove: bool = True


@dataclass
class UninstallResult:
    """Result of a Zsh uninstallation."""
    success: bool
    message: str
    removed_items: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class BaseUninstaller(ABC):
    """Abstract base class for platform-specific Zsh uninstallers."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the uninstaller."""
        self.system_info = system_info
        self.cli = cli
        self.runner = runner
        self.options: Optional[UninstallOptions] = None

    def check_installed(self) -> bool:
        """Check if Zsh is installed."""
        return self.runner.check_command_exists("zsh")

    def check_oh_my_zsh_installed(self) -> bool:
        """Check if Oh My Zsh is installed."""
        return os.path.exists(self.system_info.get_oh_my_zsh_dir())

    def get_installed_version(self) -> Optional[str]:
        """Get the installed Zsh version."""
        return self.runner.get_command_version("zsh")

    def backup_zshrc(self) -> Optional[str]:
        """Backup .zshrc before removal."""
        import datetime

        zshrc_path = self.system_info.get_zshrc_path()
        if not os.path.exists(zshrc_path):
            return None

        timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_path = f"{zshrc_path}.backup.{timestamp}"

        self.cli.print_info(f"Sauvegarde de .zshrc vers {backup_path}")
        try:
            shutil.copy2(zshrc_path, backup_path)
            return backup_path
        except Exception as e:
            self.cli.print_warning(f"Impossible de sauvegarder: {e}")
            return None

    def remove_oh_my_zsh(self) -> bool:
        """Remove Oh My Zsh directory."""
        oh_my_zsh_dir = self.system_info.get_oh_my_zsh_dir()

        if not os.path.exists(oh_my_zsh_dir):
            self.cli.print_info("Oh My Zsh n'est pas installe")
            return True

        self.cli.print_info(f"Suppression de Oh My Zsh: {oh_my_zsh_dir}")
        try:
            shutil.rmtree(oh_my_zsh_dir)
            self.cli.print_success("Oh My Zsh supprime")
            return True
        except Exception as e:
            self.cli.print_error(f"Erreur lors de la suppression: {e}")
            return False

    def remove_zshrc(self) -> bool:
        """Remove .zshrc file."""
        zshrc_path = self.system_info.get_zshrc_path()

        if not os.path.exists(zshrc_path):
            return True

        self.cli.print_info(f"Suppression de .zshrc: {zshrc_path}")
        try:
            os.remove(zshrc_path)
            self.cli.print_success(".zshrc supprime")
            return True
        except Exception as e:
            self.cli.print_error(f"Erreur lors de la suppression: {e}")
            return False

    def remove_zsh_history(self) -> bool:
        """Remove Zsh history file."""
        history_path = os.path.join(self.system_info.home_dir, ".zsh_history")

        if not os.path.exists(history_path):
            return True

        self.cli.print_info(f"Suppression de l'historique: {history_path}")
        try:
            os.remove(history_path)
            self.cli.print_success("Historique Zsh supprime")
            return True
        except Exception as e:
            self.cli.print_warning(f"Impossible de supprimer l'historique: {e}")
            return False

    def remove_zsh_cache(self) -> bool:
        """Remove Zsh cache directories."""
        cache_dirs = [
            os.path.join(self.system_info.home_dir, ".zcompdump*"),
            os.path.join(self.system_info.home_dir, ".zsh_sessions"),
            os.path.join(self.system_info.home_dir, ".cache", "zsh"),
        ]

        import glob

        for pattern in cache_dirs:
            for path in glob.glob(pattern):
                try:
                    if os.path.isdir(path):
                        shutil.rmtree(path)
                    else:
                        os.remove(path)
                    self.cli.print_success(f"Supprime: {path}")
                except Exception as e:
                    self.cli.print_warning(f"Impossible de supprimer {path}: {e}")

        return True

    def restore_default_shell(self) -> bool:
        """Restore default shell to bash."""
        # Get current shell
        current_shell = os.environ.get("SHELL", "")

        if "zsh" not in current_shell:
            self.cli.print_info("Le shell par defaut n'est pas Zsh")
            return True

        # Find bash path
        bash_path = self.runner.get_command_path("bash")
        if not bash_path:
            bash_path = "/bin/bash"

        self.cli.print_info(f"Restauration du shell par defaut vers {bash_path}")

        # Use chsh to change shell
        result = self.runner.run_interactive(
            ["chsh", "-s", bash_path],
            description="Changement du shell par defaut"
        )

        if result:
            self.cli.print_success("Shell par defaut restaure vers bash")
        else:
            self.cli.print_warning("Impossible de changer le shell par defaut")

        return result

    @abstractmethod
    def uninstall_zsh(self) -> bool:
        """Uninstall Zsh using the platform's package manager."""
        pass

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process."""
        self.options = options
        errors = []
        warnings = []
        removed_items = []

        # Step 0: Check if installed
        self.cli.print_section("Verification de l'installation")
        has_zsh = self.check_installed()
        has_oh_my_zsh = self.check_oh_my_zsh_installed()

        if not has_zsh and not has_oh_my_zsh:
            return UninstallResult(
                success=True,
                message="Zsh et Oh My Zsh ne sont pas installes",
                warnings=["Rien a desinstaller"]
            )

        if has_zsh:
            version = self.get_installed_version()
            self.cli.print_info(f"Zsh installe: {version}")

        if has_oh_my_zsh:
            self.cli.print_info("Oh My Zsh installe")

        # Step 1: Backup .zshrc
        if options.backup_before_remove and options.remove_zshrc:
            self.cli.print_section("Sauvegarde de la configuration")
            backup_path = self.backup_zshrc()
            if backup_path:
                self.cli.print_success(f"Configuration sauvegardee: {backup_path}")

        # Step 2: Restore default shell
        if options.restore_default_shell:
            self.cli.print_section("Restauration du shell par defaut")
            if self.restore_default_shell():
                removed_items.append("default_shell_changed")
            else:
                warnings.append("Shell par defaut non restaure")

        # Step 3: Remove Oh My Zsh
        if options.remove_oh_my_zsh:
            self.cli.print_section("Suppression de Oh My Zsh")
            if self.remove_oh_my_zsh():
                removed_items.append("oh-my-zsh")
            else:
                warnings.append("Oh My Zsh non supprime")

        # Step 4: Remove .zshrc
        if options.remove_zshrc:
            self.cli.print_section("Suppression de .zshrc")
            if self.remove_zshrc():
                removed_items.append(".zshrc")
            else:
                warnings.append(".zshrc non supprime")

        # Step 5: Remove cache and history
        self.cli.print_section("Nettoyage du cache et de l'historique")
        self.remove_zsh_cache()
        self.remove_zsh_history()

        # Note: We don't uninstall Zsh itself by default as it may be
        # the system shell or used by other applications

        return UninstallResult(
            success=True,
            message="Desinstallation terminee avec succes!",
            removed_items=removed_items,
            warnings=warnings
        )
