"""macOS-specific VS Code uninstaller."""

import os
import shutil
from typing import Optional
from .base import BaseUninstaller, UninstallOptions, UninstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class MacOSUninstaller(BaseUninstaller):
    """Uninstaller for VS Code on macOS."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the macOS VS Code uninstaller."""
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

    def get_vscode_paths(self) -> dict[str, str]:
        """Get VS Code related paths for macOS."""
        home = self.system_info.home_dir
        return {
            "app": "/Applications/Visual Studio Code.app",
            "settings": os.path.join(home, "Library", "Application Support", "Code"),
            "extensions": os.path.join(home, ".vscode", "extensions"),
            "cache": os.path.join(home, "Library", "Caches", "com.microsoft.VSCode"),
            "crashpad": os.path.join(home, "Library", "Application Support", "Code", "Crashpad"),
            "logs": os.path.join(home, "Library", "Logs", "Code"),
            "preferences": os.path.join(home, "Library", "Preferences", "com.microsoft.VSCode.plist"),
            "saved_state": os.path.join(home, "Library", "Saved Application State", "com.microsoft.VSCode.savedState"),
        }

    def _quit_vscode(self) -> bool:
        """Quit VS Code application."""
        self.cli.print_info("Fermeture de VS Code...")

        # Try to quit gracefully using osascript
        self.runner.run(
            ["osascript", "-e", 'quit app "Visual Studio Code"'],
            sudo=False,
            timeout=10
        )

        # Also try killall as backup
        self.runner.run(
            ["killall", "Electron"],
            sudo=False,
            timeout=5
        )

        return True

    def uninstall_vscode(self) -> bool:
        """Uninstall VS Code on macOS."""
        # First quit VS Code
        self._quit_vscode()

        brew_path = self._get_homebrew_path()

        if brew_path:
            # Check if VS Code is installed via Homebrew cask
            result = self.runner.run(
                [brew_path, "list", "--cask", "visual-studio-code"],
                sudo=False
            )

            if result.success:
                # Uninstall via Homebrew cask
                self.cli.print_info("Desinstallation de VS Code via Homebrew...")
                result = self.runner.run(
                    [brew_path, "uninstall", "--cask", "--force", "visual-studio-code"],
                    description="Desinstallation de VS Code",
                    sudo=False,
                    timeout=120
                )

                if result.success:
                    self.cli.print_success("VS Code desinstalle via Homebrew")
                    return True

        # Manual uninstallation - remove application
        return self._uninstall_manual()

    def _uninstall_manual(self) -> bool:
        """Manually uninstall VS Code by removing application."""
        self.cli.print_info("Desinstallation manuelle de VS Code...")

        paths = self.get_vscode_paths()
        app_path = paths.get("app")

        if app_path and os.path.exists(app_path):
            try:
                shutil.rmtree(app_path)
                self.cli.print_success(f"Application supprimee: {app_path}")
            except Exception as e:
                self.cli.print_warning(f"Impossible de supprimer {app_path}: {e}")
                return False

        # Remove 'code' command from PATH
        code_link = "/usr/local/bin/code"
        if os.path.exists(code_link) or os.path.islink(code_link):
            try:
                os.remove(code_link)
                self.cli.print_success("Commande 'code' supprimee")
            except Exception:
                pass

        return True

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process for macOS."""
        # Quit VS Code first
        self._quit_vscode()

        # Run base uninstallation
        result = super().uninstall(options)

        if result.success:
            # Remove additional macOS-specific paths
            paths = self.get_vscode_paths()
            for key in ["logs", "preferences", "saved_state"]:
                path = paths.get(key)
                if path and os.path.exists(path):
                    try:
                        if os.path.isdir(path):
                            shutil.rmtree(path)
                        else:
                            os.remove(path)
                        self.cli.print_success(f"Supprime: {path}")
                    except Exception:
                        pass

        return result
