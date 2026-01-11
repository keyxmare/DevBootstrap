"""macOS-specific Nerd Font uninstaller."""

import os
import glob
import shutil
from typing import Optional
from .base import BaseUninstaller, UninstallOptions, UninstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner
from ..installers.base import FontInfo


class MacOSUninstaller(BaseUninstaller):
    """Uninstaller for Nerd Fonts on macOS."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the macOS font uninstaller."""
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

    def get_fonts_directory(self) -> str:
        """Get the fonts directory for macOS."""
        return os.path.join(self.system_info.home_dir, "Library", "Fonts")

    def _get_font_patterns(self, font: FontInfo) -> list[str]:
        """Get file patterns for a font."""
        patterns = {
            "meslo": ["MesloLG*.ttf", "MesloLGS*.ttf"],
            "fira-code": ["FiraCode*.ttf", "Fira Code*.ttf"],
            "jetbrains-mono": ["JetBrainsMono*.ttf", "JetBrains Mono*.ttf"],
            "hack": ["Hack*.ttf"],
        }
        return patterns.get(font.id, [])

    def check_font_installed(self, font: FontInfo) -> bool:
        """Check if a font is installed on macOS."""
        brew_path = self._get_homebrew_path()

        # Check via Homebrew cask first
        if brew_path and font.homebrew_cask:
            result = self.runner.run(
                [brew_path, "list", "--cask", font.homebrew_cask],
                sudo=False
            )
            if result.success:
                return True

        # Check in fonts directory
        fonts_dir = self.get_fonts_directory()
        patterns = self._get_font_patterns(font)

        for pattern in patterns:
            matches = glob.glob(os.path.join(fonts_dir, pattern))
            if matches:
                return True

        return False

    def uninstall_font(self, font: FontInfo) -> bool:
        """Uninstall a font on macOS."""
        brew_path = self._get_homebrew_path()
        success = True

        # Try Homebrew cask first
        if brew_path and font.homebrew_cask:
            result = self.runner.run(
                [brew_path, "list", "--cask", font.homebrew_cask],
                sudo=False
            )

            if result.success:
                self.cli.print_info(f"Desinstallation de {font.name} via Homebrew...")
                result = self.runner.run(
                    [brew_path, "uninstall", "--cask", "--force", font.homebrew_cask],
                    description=f"Desinstallation de {font.name}",
                    sudo=False
                )

                if result.success:
                    return True
                else:
                    self.cli.print_warning("Echec via Homebrew, tentative manuelle...")

        # Manual removal from fonts directory
        fonts_dir = self.get_fonts_directory()
        patterns = self._get_font_patterns(font)

        for pattern in patterns:
            for font_file in glob.glob(os.path.join(fonts_dir, pattern)):
                try:
                    os.remove(font_file)
                    self.cli.print_success(f"Supprime: {os.path.basename(font_file)}")
                except Exception as e:
                    self.cli.print_warning(f"Impossible de supprimer {font_file}: {e}")
                    success = False

        return success

    def refresh_font_cache(self) -> bool:
        """Refresh font cache on macOS."""
        # macOS doesn't require manual font cache refresh
        # The system automatically updates when fonts are added/removed
        self.cli.print_info("Le cache de polices sera mis a jour automatiquement")
        return True

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process for macOS."""
        return super().uninstall(options)
