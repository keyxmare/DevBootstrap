"""macOS-specific Zsh uninstaller."""

import os
from typing import Optional
from .base import BaseUninstaller, UninstallOptions, UninstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class MacOSUninstaller(BaseUninstaller):
    """Uninstaller for Zsh on macOS."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the macOS Zsh uninstaller."""
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

    def uninstall_zsh(self) -> bool:
        """
        On macOS, Zsh is the default shell and should not be uninstalled.
        We only remove Oh My Zsh and configuration.
        """
        self.cli.print_info("Zsh est le shell par defaut de macOS")
        self.cli.print_info("Zsh ne sera pas desinstalle (systeme)")
        self.cli.print_info("Seuls Oh My Zsh et la configuration seront supprimes")
        return True

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process for macOS."""
        # Run base uninstallation
        result = super().uninstall(options)

        if result.success:
            # On macOS, remind user that Zsh is the system shell
            result.warnings.append(
                "Zsh est le shell par defaut de macOS et n'a pas ete desinstalle"
            )

        return result
