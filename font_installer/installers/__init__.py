"""Font installers for different platforms."""

from .base import BaseInstaller, InstallOptions, InstallResult, FontInfo
from .macos import MacOSInstaller
from .ubuntu import UbuntuInstaller

__all__ = [
    "BaseInstaller",
    "InstallOptions",
    "InstallResult",
    "FontInfo",
    "MacOSInstaller",
    "UbuntuInstaller",
]
