"""Base uninstaller class for Nerd Font uninstallation."""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Optional
import os
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner
from ..installers.base import FontInfo, AVAILABLE_FONTS


@dataclass
class UninstallOptions:
    """Options for the font uninstallation process."""
    fonts: list[FontInfo] = field(default_factory=list)
    remove_all: bool = False


@dataclass
class UninstallResult:
    """Result of a font uninstallation."""
    success: bool
    message: str
    removed_fonts: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class BaseUninstaller(ABC):
    """Abstract base class for platform-specific font uninstallers."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the uninstaller."""
        self.system_info = system_info
        self.cli = cli
        self.runner = runner
        self.options: Optional[UninstallOptions] = None

    @abstractmethod
    def get_fonts_directory(self) -> str:
        """Get the fonts directory for this platform."""
        pass

    @abstractmethod
    def check_font_installed(self, font: FontInfo) -> bool:
        """Check if a font is installed."""
        pass

    @abstractmethod
    def uninstall_font(self, font: FontInfo) -> bool:
        """Uninstall a specific font."""
        pass

    def get_installed_fonts(self) -> list[FontInfo]:
        """Get list of installed Nerd Fonts."""
        installed = []
        for font in AVAILABLE_FONTS:
            if self.check_font_installed(font):
                installed.append(font)
        return installed

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process."""
        self.options = options
        errors = []
        warnings = []
        removed_fonts = []

        # Determine which fonts to uninstall
        fonts_to_remove = options.fonts
        if options.remove_all:
            fonts_to_remove = self.get_installed_fonts()

        if not fonts_to_remove:
            return UninstallResult(
                success=True,
                message="Aucune police Nerd Font a desinstaller",
                warnings=["Aucune police selectionnee ou installee"]
            )

        # Step 1: Show fonts to remove
        self.cli.print_section("Polices a desinstaller")
        for font in fonts_to_remove:
            if self.check_font_installed(font):
                self.cli.print_info(f"  - {font.name}")
            else:
                self.cli.print_info(f"  - {font.name} (non installee)")

        # Step 2: Uninstall each font
        self.cli.print_section("Desinstallation des polices")
        for font in fonts_to_remove:
            if not self.check_font_installed(font):
                self.cli.print_info(f"{font.name} n'est pas installee")
                continue

            self.cli.print_info(f"Desinstallation de {font.name}...")
            if self.uninstall_font(font):
                self.cli.print_success(f"{font.name} desinstallee")
                removed_fonts.append(font.name)
            else:
                self.cli.print_error(f"Echec de la desinstallation de {font.name}")
                errors.append(f"Echec: {font.name}")

        # Step 3: Refresh font cache
        self.cli.print_section("Mise a jour du cache de polices")
        self.refresh_font_cache()

        if errors:
            return UninstallResult(
                success=False,
                message=f"{len(removed_fonts)} police(s) desinstallee(s), {len(errors)} echec(s)",
                removed_fonts=removed_fonts,
                errors=errors,
                warnings=warnings
            )

        return UninstallResult(
            success=True,
            message=f"{len(removed_fonts)} police(s) desinstallee(s) avec succes",
            removed_fonts=removed_fonts,
            warnings=warnings
        )

    @abstractmethod
    def refresh_font_cache(self) -> bool:
        """Refresh the system font cache."""
        pass
