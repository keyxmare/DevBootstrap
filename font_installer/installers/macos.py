"""macOS-specific font installer using Homebrew."""

import os
from typing import Optional
from .base import BaseInstaller, FontInfo, InstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class MacOSInstaller(BaseInstaller):
    """Font installer for macOS using Homebrew Cask."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the macOS installer."""
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

    def _ensure_homebrew_cask_fonts(self) -> bool:
        """Ensure homebrew-cask-fonts tap is added."""
        brew_path = self._get_homebrew_path()
        if not brew_path:
            return False

        # Check if tap already exists
        result = self.runner.run(
            [brew_path, "tap"],
            sudo=False
        )

        if result.success and "homebrew/cask-fonts" in result.stdout:
            return True

        # Add the tap
        self.cli.print_info("Ajout du tap homebrew/cask-fonts...")
        result = self.runner.run(
            [brew_path, "tap", "homebrew/cask-fonts"],
            description="Ajout de homebrew/cask-fonts",
            sudo=False,
            timeout=120
        )

        return result.success

    def check_font_installed(self, font: FontInfo) -> bool:
        """Check if a font is installed via Homebrew cask."""
        brew_path = self._get_homebrew_path()
        if not brew_path:
            return False

        result = self.runner.run(
            [brew_path, "list", "--cask", font.homebrew_cask],
            sudo=False
        )

        return result.success

    def install_font(self, font: FontInfo) -> bool:
        """Install a font using Homebrew cask."""
        brew_path = self._get_homebrew_path()
        if not brew_path:
            self.cli.print_error("Homebrew n'est pas installe")
            return False

        # Ensure cask-fonts tap is available
        if not self._ensure_homebrew_cask_fonts():
            self.cli.print_error("Impossible d'ajouter le tap homebrew/cask-fonts")
            return False

        # Install the font
        result = self.runner.run(
            [brew_path, "install", "--cask", font.homebrew_cask],
            description=f"Installation de {font.name}",
            sudo=False,
            timeout=300
        )

        return result.success

    def install(self, options) -> InstallResult:
        """Run the complete installation process for macOS."""
        # Check Homebrew is available
        if not self._get_homebrew_path():
            return InstallResult(
                success=False,
                message="Homebrew n'est pas installe. Veuillez d'abord installer Homebrew.",
                errors=["Homebrew requis pour installer les polices sur macOS"]
            )

        return super().install(options)
