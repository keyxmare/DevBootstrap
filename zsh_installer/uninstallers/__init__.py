"""Uninstallers package for Zsh."""

from .base import BaseUninstaller, UninstallOptions, UninstallResult
from .macos import MacOSUninstaller
from .ubuntu import UbuntuUninstaller

__all__ = [
    "BaseUninstaller",
    "UninstallOptions",
    "UninstallResult",
    "MacOSUninstaller",
    "UbuntuUninstaller",
]
