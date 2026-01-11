"""Platform-specific installers for Zsh."""

from .base import BaseInstaller, InstallOptions, InstallResult, Dependency
from .macos import MacOSInstaller
from .ubuntu import UbuntuInstaller

__all__ = [
    "BaseInstaller",
    "InstallOptions",
    "InstallResult",
    "Dependency",
    "MacOSInstaller",
    "UbuntuInstaller",
]
