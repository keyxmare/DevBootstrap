"""macOS-specific Neovim uninstaller using Homebrew."""

import os
from typing import Optional
from .base import BaseUninstaller, UninstallOptions, UninstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class MacOSUninstaller(BaseUninstaller):
    """Uninstaller for macOS using Homebrew."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the macOS uninstaller."""
        super().__init__(system_info, cli, runner)
        self._homebrew_path: Optional[str] = None

    def _get_homebrew_path(self) -> Optional[str]:
        """Get the Homebrew installation path."""
        if self._homebrew_path:
            return self._homebrew_path

        # Check common Homebrew locations
        possible_paths = [
            "/opt/homebrew/bin/brew",  # Apple Silicon
            "/usr/local/bin/brew",      # Intel
        ]

        for path in possible_paths:
            if os.path.exists(path):
                self._homebrew_path = path
                return path

        # Check if brew is in PATH
        brew_path = self.runner.get_command_path("brew")
        if brew_path:
            self._homebrew_path = brew_path
            return brew_path

        return None

    def uninstall_neovim(self) -> bool:
        """Uninstall Neovim using Homebrew."""
        brew_path = self._get_homebrew_path()
        if not brew_path:
            self.cli.print_error("Homebrew non trouve")
            return False

        # Check if neovim is installed via Homebrew
        result = self.runner.run(
            [brew_path, "list", "neovim"],
            sudo=False
        )

        if not result.success:
            self.cli.print_info("Neovim n'est pas installe via Homebrew")
            # Try to find other installation methods
            nvim_path = self.runner.get_command_path("nvim")
            if nvim_path:
                self.cli.print_warning(f"Neovim trouve a: {nvim_path}")
                self.cli.print_info("Suppression manuelle peut etre necessaire")
            return True

        # Uninstall neovim
        self.cli.print_info("Desinstallation de Neovim via Homebrew...")
        result = self.runner.run(
            [brew_path, "uninstall", "--force", "neovim"],
            description="Desinstallation de Neovim",
            sudo=False
        )

        if result.success:
            self.cli.print_success("Neovim desinstalle via Homebrew")
        else:
            self.cli.print_error("Echec de la desinstallation")

        return result.success

    def uninstall_dependencies(self) -> bool:
        """Optionally uninstall dependencies."""
        # Note: We don't uninstall dependencies by default as they might be
        # used by other applications
        self.cli.print_info("Les dependances (ripgrep, fzf, etc.) ne sont pas supprimees")
        self.cli.print_info("Utilisez 'brew uninstall <package>' pour les supprimer manuellement")
        return True

    def uninstall_python_provider(self) -> bool:
        """Uninstall Python provider for Neovim."""
        result = self.runner.run(
            ["pip3", "uninstall", "-y", "pynvim"],
            description="Desinstallation du provider Python (pynvim)",
            sudo=False
        )
        return result.success

    def uninstall_node_provider(self) -> bool:
        """Uninstall Node.js provider for Neovim."""
        if not self.runner.check_command_exists("npm"):
            return True

        result = self.runner.run(
            ["npm", "uninstall", "-g", "neovim"],
            description="Desinstallation du provider Node.js",
            sudo=False
        )
        return result.success

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process for macOS."""
        # Run base uninstallation
        result = super().uninstall(options)

        if result.success:
            # Uninstall providers
            self.cli.print_section("Desinstallation des providers")

            if self.uninstall_python_provider():
                self.cli.print_success("Provider Python desinstalle")
            else:
                result.warnings.append("Provider Python non desinstalle")

            if self.uninstall_node_provider():
                self.cli.print_success("Provider Node.js desinstalle")
            else:
                result.warnings.append("Provider Node.js non desinstalle")

        return result
