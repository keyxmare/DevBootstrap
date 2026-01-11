"""Base installer interface for fonts."""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Optional
from enum import Enum

from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class FontFamily(Enum):
    """Available Nerd Font families."""
    MESLO = "meslo"
    FIRA_CODE = "fira-code"
    JETBRAINS_MONO = "jetbrains-mono"
    HACK = "hack"
    SOURCE_CODE_PRO = "source-code-pro"


@dataclass
class FontInfo:
    """Information about a font."""
    id: str
    name: str
    description: str
    homebrew_cask: str  # Homebrew cask name for macOS
    apt_package: Optional[str] = None  # APT package name for Ubuntu/Debian


# Available Nerd Fonts
AVAILABLE_FONTS = [
    FontInfo(
        id="meslo",
        name="MesloLG Nerd Font",
        description="Police recommandee pour le theme agnoster (Powerline)",
        homebrew_cask="font-meslo-lg-nerd-font",
        apt_package=None  # Manual installation on Linux
    ),
    FontInfo(
        id="fira-code",
        name="FiraCode Nerd Font",
        description="Police avec ligatures pour le code",
        homebrew_cask="font-fira-code-nerd-font",
        apt_package=None
    ),
    FontInfo(
        id="jetbrains-mono",
        name="JetBrainsMono Nerd Font",
        description="Police JetBrains avec icones Nerd Font",
        homebrew_cask="font-jetbrains-mono-nerd-font",
        apt_package=None
    ),
    FontInfo(
        id="hack",
        name="Hack Nerd Font",
        description="Police Hack avec icones Nerd Font",
        homebrew_cask="font-hack-nerd-font",
        apt_package=None
    ),
]


@dataclass
class InstallOptions:
    """Options for font installation."""
    fonts: list[FontInfo] = field(default_factory=list)


@dataclass
class InstallResult:
    """Result of font installation."""
    success: bool
    message: str
    installed_fonts: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)


class BaseInstaller(ABC):
    """Abstract base class for font installers."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the installer."""
        self.system_info = system_info
        self.cli = cli
        self.runner = runner

    @abstractmethod
    def install_font(self, font: FontInfo) -> bool:
        """Install a specific font."""
        pass

    @abstractmethod
    def check_font_installed(self, font: FontInfo) -> bool:
        """Check if a font is already installed."""
        pass

    def install(self, options: InstallOptions) -> InstallResult:
        """Install the selected fonts."""
        if not options.fonts:
            return InstallResult(
                success=True,
                message="Aucune police selectionnee"
            )

        installed = []
        errors = []
        warnings = []

        for font in options.fonts:
            self.cli.print_info(f"Installation de {font.name}...")

            if self.check_font_installed(font):
                self.cli.print_info(f"{font.name} est deja installee")
                installed.append(font.name)
                continue

            if self.install_font(font):
                self.cli.print_success(f"{font.name} installee avec succes")
                installed.append(font.name)
            else:
                self.cli.print_error(f"Echec de l'installation de {font.name}")
                errors.append(f"Echec: {font.name}")

        if errors:
            return InstallResult(
                success=False,
                message=f"{len(installed)} police(s) installee(s), {len(errors)} echec(s)",
                installed_fonts=installed,
                errors=errors,
                warnings=warnings
            )

        return InstallResult(
            success=True,
            message=f"{len(installed)} police(s) installee(s) avec succes",
            installed_fonts=installed,
            warnings=warnings
        )
