"""Ubuntu/Debian-specific Nerd Font uninstaller."""

import os
import glob
import shutil
from .base import BaseUninstaller, UninstallOptions, UninstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner
from ..installers.base import FontInfo


class UbuntuUninstaller(BaseUninstaller):
    """Uninstaller for Nerd Fonts on Ubuntu/Debian."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the Ubuntu font uninstaller."""
        super().__init__(system_info, cli, runner)

    def get_fonts_directory(self) -> str:
        """Get the fonts directory for Ubuntu/Debian."""
        return os.path.join(self.system_info.home_dir, ".local", "share", "fonts")

    def _get_font_patterns(self, font: FontInfo) -> list[str]:
        """Get file patterns for a font."""
        patterns = {
            "meslo": ["MesloLG*", "MesloLGS*"],
            "fira-code": ["FiraCode*", "Fira Code*"],
            "jetbrains-mono": ["JetBrainsMono*", "JetBrains Mono*"],
            "hack": ["Hack*"],
        }
        return patterns.get(font.id, [])

    def _get_font_directory_name(self, font: FontInfo) -> str:
        """Get the directory name for a font."""
        dir_names = {
            "meslo": "MesloLGNerdFont",
            "fira-code": "FiraCodeNerdFont",
            "jetbrains-mono": "JetBrainsMonoNerdFont",
            "hack": "HackNerdFont",
        }
        return dir_names.get(font.id, "")

    def check_font_installed(self, font: FontInfo) -> bool:
        """Check if a font is installed on Ubuntu/Debian."""
        fonts_dir = self.get_fonts_directory()

        # Check for font directory
        font_dir_name = self._get_font_directory_name(font)
        if font_dir_name:
            font_path = os.path.join(fonts_dir, font_dir_name)
            if os.path.exists(font_path):
                return True

        # Check for font files with patterns
        patterns = self._get_font_patterns(font)
        for pattern in patterns:
            # Check in fonts root
            matches = glob.glob(os.path.join(fonts_dir, pattern + "*.ttf"))
            if matches:
                return True

            # Check in subdirectories
            matches = glob.glob(os.path.join(fonts_dir, "**", pattern + "*.ttf"), recursive=True)
            if matches:
                return True

        return False

    def uninstall_font(self, font: FontInfo) -> bool:
        """Uninstall a font on Ubuntu/Debian."""
        fonts_dir = self.get_fonts_directory()
        success = True

        # Remove font directory if exists
        font_dir_name = self._get_font_directory_name(font)
        if font_dir_name:
            font_path = os.path.join(fonts_dir, font_dir_name)
            if os.path.exists(font_path):
                try:
                    shutil.rmtree(font_path)
                    self.cli.print_success(f"Supprime: {font_path}")
                except Exception as e:
                    self.cli.print_warning(f"Impossible de supprimer {font_path}: {e}")
                    success = False

        # Remove individual font files
        patterns = self._get_font_patterns(font)
        for pattern in patterns:
            # Remove from fonts root
            for font_file in glob.glob(os.path.join(fonts_dir, pattern + "*.ttf")):
                try:
                    os.remove(font_file)
                    self.cli.print_success(f"Supprime: {os.path.basename(font_file)}")
                except Exception as e:
                    self.cli.print_warning(f"Impossible de supprimer {font_file}: {e}")
                    success = False

            # Remove from subdirectories
            for font_file in glob.glob(os.path.join(fonts_dir, "**", pattern + "*.ttf"), recursive=True):
                try:
                    os.remove(font_file)
                    self.cli.print_success(f"Supprime: {os.path.basename(font_file)}")
                except Exception as e:
                    self.cli.print_warning(f"Impossible de supprimer {font_file}: {e}")
                    success = False

        return success

    def refresh_font_cache(self) -> bool:
        """Refresh font cache on Ubuntu/Debian."""
        self.cli.print_info("Mise a jour du cache de polices...")

        result = self.runner.run(
            ["fc-cache", "-fv"],
            description="Mise a jour du cache de polices",
            sudo=False,
            timeout=60
        )

        if result.success:
            self.cli.print_success("Cache de polices mis a jour")
        else:
            self.cli.print_warning("Echec de la mise a jour du cache")

        return result.success

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process for Ubuntu/Debian."""
        return super().uninstall(options)
