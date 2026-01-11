"""System detection and information module."""

import platform
import subprocess
import os
from dataclasses import dataclass
from enum import Enum
from typing import Optional


class OSType(Enum):
    """Supported operating system types."""
    MACOS = "macos"
    UBUNTU = "ubuntu"
    DEBIAN = "debian"
    LINUX_OTHER = "linux_other"
    UNSUPPORTED = "unsupported"


class Architecture(Enum):
    """CPU architecture types."""
    ARM64 = "arm64"
    X86_64 = "x86_64"
    UNKNOWN = "unknown"


@dataclass
class SystemInfo:
    """Container for system information."""
    os_type: OSType
    os_name: str
    os_version: str
    architecture: Architecture
    home_dir: str
    is_root: bool
    has_sudo: bool

    @classmethod
    def detect(cls) -> "SystemInfo":
        """Detect current system information."""
        system = platform.system().lower()
        machine = platform.machine().lower()

        # Detect OS type
        if system == "darwin":
            os_type = OSType.MACOS
            os_name = "macOS"
            os_version = platform.mac_ver()[0]
        elif system == "linux":
            os_name, os_version = cls._detect_linux_distro()
            if "ubuntu" in os_name.lower():
                os_type = OSType.UBUNTU
            elif "debian" in os_name.lower():
                os_type = OSType.DEBIAN
            else:
                os_type = OSType.LINUX_OTHER
        else:
            os_type = OSType.UNSUPPORTED
            os_name = system
            os_version = "unknown"

        # Detect architecture
        if machine in ("arm64", "aarch64"):
            architecture = Architecture.ARM64
        elif machine in ("x86_64", "amd64"):
            architecture = Architecture.X86_64
        else:
            architecture = Architecture.UNKNOWN

        # Get home directory
        home_dir = os.path.expanduser("~")

        # Check root/sudo
        is_root = os.geteuid() == 0 if hasattr(os, 'geteuid') else False
        has_sudo = cls._check_sudo_available()

        return cls(
            os_type=os_type,
            os_name=os_name,
            os_version=os_version,
            architecture=architecture,
            home_dir=home_dir,
            is_root=is_root,
            has_sudo=has_sudo
        )

    @staticmethod
    def _detect_linux_distro() -> tuple[str, str]:
        """Detect Linux distribution name and version."""
        try:
            with open("/etc/os-release", "r") as f:
                lines = f.readlines()

            info = {}
            for line in lines:
                if "=" in line:
                    key, value = line.strip().split("=", 1)
                    info[key] = value.strip('"')

            name = info.get("NAME", "Linux")
            version = info.get("VERSION_ID", "unknown")
            return name, version
        except FileNotFoundError:
            return "Linux", "unknown"

    @staticmethod
    def _check_sudo_available() -> bool:
        """Check if sudo is available and user can use it."""
        try:
            result = subprocess.run(
                ["sudo", "-n", "true"],
                capture_output=True,
                timeout=5
            )
            return result.returncode == 0
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return False

    def is_supported(self) -> bool:
        """Check if the current system is supported."""
        return self.os_type in (OSType.MACOS, OSType.UBUNTU, OSType.DEBIAN)

    def get_zshrc_path(self) -> str:
        """Get the .zshrc path."""
        return os.path.join(self.home_dir, ".zshrc")

    def get_oh_my_zsh_dir(self) -> str:
        """Get the Oh My Zsh installation directory."""
        return os.path.join(self.home_dir, ".oh-my-zsh")

    def get_zsh_custom_dir(self) -> str:
        """Get the custom directory for Zsh plugins/themes."""
        return os.path.join(self.get_oh_my_zsh_dir(), "custom")

    def __str__(self) -> str:
        """Human-readable system info."""
        return (
            f"{self.os_name} {self.os_version} "
            f"({self.architecture.value})"
        )
