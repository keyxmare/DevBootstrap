"""Platform-specific installers for Neovim."""

from .base import BaseInstaller
from .macos import MacOSInstaller
from .ubuntu import UbuntuInstaller

__all__ = ["BaseInstaller", "MacOSInstaller", "UbuntuInstaller"]
