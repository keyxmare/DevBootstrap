"""Application registry for DevBootstrap."""

from dataclasses import dataclass
from typing import Callable, Optional
from enum import Enum


class AppStatus(Enum):
    """Status of an application."""
    NOT_INSTALLED = "non installé"
    INSTALLED = "installé"
    UPDATE_AVAILABLE = "mise à jour disponible"


@dataclass
class AppInfo:
    """Information about an installable application."""
    id: str
    name: str
    description: str
    check_command: str  # Command to check if installed
    version_command: Optional[str] = None  # Command to get version
    module: Optional[str] = None  # Python module to import for installation


# Registry of all available applications
AVAILABLE_APPS = [
    AppInfo(
        id="docker",
        name="Docker",
        description="Plateforme de conteneurisation",
        check_command="docker",
        version_command="docker --version",
        module="docker_installer.app"
    ),
    AppInfo(
        id="vscode",
        name="Visual Studio Code",
        description="Éditeur de code source léger et puissant",
        check_command="code",
        version_command="code --version | head -1",
        module="vscode_installer.app"
    ),
    AppInfo(
        id="neovim",
        name="Neovim",
        description="Éditeur de texte moderne et extensible",
        check_command="nvim",
        version_command="nvim --version | head -1",
        module="nvim_installer.app"
    ),
]
