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
    macos_app_paths: Optional[list[str]] = None  # macOS .app paths to check


# Registry of all available applications
AVAILABLE_APPS = [
    AppInfo(
        id="docker",
        name="Docker",
        description="Plateforme de conteneurisation",
        check_command="docker",
        version_command="docker --version",
        module="docker_installer.app",
        macos_app_paths=["/Applications/Docker.app"]
    ),
    AppInfo(
        id="vscode",
        name="Visual Studio Code",
        description="Editeur de code source leger et puissant",
        check_command="code",
        version_command="code --version | head -1",
        module="vscode_installer.app",
        macos_app_paths=["/Applications/Visual Studio Code.app"]
    ),
    AppInfo(
        id="neovim",
        name="Neovim",
        description="Editeur de texte moderne (sans configuration)",
        check_command="nvim",
        version_command="nvim --version | head -1",
        module="nvim_installer.app"
    ),
    AppInfo(
        id="neovim-config",
        name="Neovim Config",
        description="Configuration et plugins pour Neovim (necessite Neovim)",
        check_command="nvim",
        version_command=None,
        module="nvim_installer.app"
    ),
    AppInfo(
        id="zsh",
        name="Zsh",
        description="Shell Z moderne (sans Oh My Zsh)",
        check_command="zsh",
        version_command="zsh --version | head -1",
        module="zsh_installer.app"
    ),
    AppInfo(
        id="oh-my-zsh",
        name="Oh My Zsh",
        description="Framework de configuration pour Zsh avec plugins (necessite Zsh)",
        check_command="zsh",
        version_command=None,
        module="zsh_installer.app"
    ),
    AppInfo(
        id="alias",
        name="Commande devbootstrap",
        description="Installe la commande 'devbootstrap' pour lancer l'installation",
        check_command="devbootstrap",
        version_command=None,
        module="alias_installer.app"
    ),
]
