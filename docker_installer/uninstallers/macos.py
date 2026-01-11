"""macOS-specific Docker uninstaller using Homebrew."""

import os
import shutil
from typing import Optional
from .base import BaseUninstaller, UninstallOptions, UninstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class MacOSUninstaller(BaseUninstaller):
    """Uninstaller for Docker on macOS (Docker Desktop)."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the macOS Docker uninstaller."""
        super().__init__(system_info, cli, runner)
        self._homebrew_path: Optional[str] = None

    def _get_homebrew_path(self) -> Optional[str]:
        """Get the Homebrew installation path."""
        if self._homebrew_path:
            return self._homebrew_path

        possible_paths = [
            "/opt/homebrew/bin/brew",  # Apple Silicon
            "/usr/local/bin/brew",      # Intel
        ]

        for path in possible_paths:
            if os.path.exists(path):
                self._homebrew_path = path
                return path

        brew_path = self.runner.get_command_path("brew")
        if brew_path:
            self._homebrew_path = brew_path
            return brew_path

        return None

    def _quit_docker_desktop(self) -> bool:
        """Quit Docker Desktop application."""
        self.cli.print_info("Fermeture de Docker Desktop...")

        # Try to quit gracefully using osascript
        result = self.runner.run(
            ["osascript", "-e", 'quit app "Docker"'],
            sudo=False,
            timeout=10
        )

        # Also try killall as backup
        self.runner.run(
            ["killall", "Docker"],
            sudo=False,
            timeout=5
        )

        return True

    def uninstall_docker(self) -> bool:
        """Uninstall Docker Desktop using Homebrew."""
        # First quit Docker Desktop
        self._quit_docker_desktop()

        brew_path = self._get_homebrew_path()
        if not brew_path:
            self.cli.print_warning("Homebrew non trouve")
            return self._uninstall_manual()

        # Check if docker is installed via Homebrew (cask)
        result = self.runner.run(
            [brew_path, "list", "--cask", "docker"],
            sudo=False
        )

        if result.success:
            # Uninstall via Homebrew cask
            self.cli.print_info("Desinstallation de Docker Desktop via Homebrew...")
            result = self.runner.run(
                [brew_path, "uninstall", "--cask", "--force", "docker"],
                description="Desinstallation de Docker Desktop",
                sudo=False,
                timeout=120
            )

            if result.success:
                self.cli.print_success("Docker Desktop desinstalle via Homebrew")
                return True
            else:
                self.cli.print_warning("Echec via Homebrew, tentative manuelle...")
                return self._uninstall_manual()
        else:
            self.cli.print_info("Docker n'est pas installe via Homebrew")
            return self._uninstall_manual()

    def _uninstall_manual(self) -> bool:
        """Manually uninstall Docker Desktop by removing application."""
        self.cli.print_info("Desinstallation manuelle de Docker Desktop...")

        # Remove Docker.app
        docker_app_paths = [
            "/Applications/Docker.app",
            os.path.expanduser("~/Applications/Docker.app"),
        ]

        for app_path in docker_app_paths:
            if os.path.exists(app_path):
                try:
                    shutil.rmtree(app_path)
                    self.cli.print_success(f"Supprime: {app_path}")
                except Exception as e:
                    self.cli.print_warning(f"Impossible de supprimer {app_path}: {e}")

        return True

    def remove_docker_data(self) -> bool:
        """Remove Docker data directories on macOS."""
        home = self.system_info.home_dir

        paths_to_remove = [
            # Docker Desktop directories
            os.path.join(home, ".docker"),
            os.path.join(home, "Library", "Group Containers", "group.com.docker"),
            os.path.join(home, "Library", "Containers", "com.docker.docker"),
            os.path.join(home, "Library", "Application Support", "Docker Desktop"),
            os.path.join(home, "Library", "Preferences", "com.docker.docker.plist"),
            os.path.join(home, "Library", "Saved Application State", "com.electron.docker-frontend.savedState"),
            os.path.join(home, "Library", "Logs", "Docker Desktop"),
            os.path.join(home, "Library", "Cookies", "com.docker.docker.binarycookies"),
            # Docker CLI config
            os.path.join(home, ".docker"),
        ]

        success = True
        for path in paths_to_remove:
            if os.path.exists(path):
                try:
                    if os.path.isdir(path):
                        shutil.rmtree(path)
                    else:
                        os.remove(path)
                    self.cli.print_success(f"Supprime: {path}")
                except Exception as e:
                    self.cli.print_warning(f"Impossible de supprimer {path}: {e}")
                    success = False

        return success

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process for macOS."""
        # Quit Docker Desktop first
        self._quit_docker_desktop()

        # Run base uninstallation
        return super().uninstall(options)
