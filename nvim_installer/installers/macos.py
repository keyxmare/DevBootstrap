"""macOS-specific Neovim installer using Homebrew."""

import os
from typing import Optional
from .base import BaseInstaller, Dependency, InstallOptions, InstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class MacOSInstaller(BaseInstaller):
    """Installer for macOS using Homebrew."""

    # macOS-specific dependencies (using Homebrew package names)
    MACOS_DEPENDENCIES: list[Dependency] = [
        Dependency(
            name="git",
            description="Système de contrôle de version",
            check_command="git"
        ),
        Dependency(
            name="node",
            description="Runtime JavaScript (pour LSP, plugins)",
            check_command="node",
            required=False
        ),
        Dependency(
            name="python@3.12",
            description="Python 3 (pour plugins Python)",
            check_command="python3"
        ),
        Dependency(
            name="ripgrep",
            description="Recherche ultra-rapide (pour Telescope)",
            check_command="rg"
        ),
        Dependency(
            name="fd",
            description="Alternative à find (pour Telescope)",
            check_command="fd",
            required=False
        ),
        Dependency(
            name="fzf",
            description="Fuzzy finder",
            check_command="fzf",
            required=False
        ),
        Dependency(
            name="lazygit",
            description="Interface Git en terminal",
            check_command="lazygit",
            required=False
        ),
        Dependency(
            name="lua",
            description="Langage Lua (pour la configuration)",
            check_command="lua",
            required=False
        ),
        Dependency(
            name="luarocks",
            description="Gestionnaire de paquets Lua",
            check_command="luarocks",
            required=False
        ),
    ]

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the macOS installer."""
        super().__init__(system_info, cli, runner)
        self._homebrew_path: Optional[str] = None

    def get_dependencies(self) -> list[Dependency]:
        """Get macOS-specific dependencies."""
        return self.MACOS_DEPENDENCIES

    def _get_homebrew_path(self) -> Optional[str]:
        """Get the Homebrew installation path."""
        if self._homebrew_path:
            return self._homebrew_path

        # Check common Homebrew locations
        possible_paths = [
            "/opt/homebrew/bin/brew",  # Apple Silicon
            "/usr/local/bin/brew",      # Intel
        ]

        for path in possible_paths:
            if os.path.exists(path):
                self._homebrew_path = path
                return path

        # Check if brew is in PATH
        brew_path = self.runner.get_command_path("brew")
        if brew_path:
            self._homebrew_path = brew_path
            return brew_path

        return None

    def _ensure_homebrew_in_path(self) -> bool:
        """Ensure Homebrew is in the PATH for this session."""
        brew_path = self._get_homebrew_path()
        if not brew_path:
            return False

        brew_dir = os.path.dirname(brew_path)
        current_path = os.environ.get("PATH", "")

        if brew_dir not in current_path:
            os.environ["PATH"] = f"{brew_dir}:{current_path}"

        return True

    def install_package_manager(self) -> bool:
        """Install Homebrew if not present."""
        if self._get_homebrew_path():
            self._ensure_homebrew_in_path()
            self.cli.print_success("Homebrew est déjà installé")
            return True

        self.cli.print_info("Installation de Homebrew...")

        # Homebrew installation script
        install_script = '/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'

        # Run interactively as it needs user input
        result = self.runner.run_interactive(
            ["/bin/bash", "-c", install_script],
            description="Installation de Homebrew",
            sudo=False
        )

        if not result:
            return False

        # Add Homebrew to PATH
        self._ensure_homebrew_in_path()

        return self._get_homebrew_path() is not None

    def update_package_manager(self) -> bool:
        """Update Homebrew."""
        brew_path = self._get_homebrew_path()
        if not brew_path:
            return False

        result = self.runner.run(
            [brew_path, "update"],
            description="Mise à jour de Homebrew",
            sudo=False,
            timeout=300  # 5 minutes timeout
        )

        return result.success

    def install_package(self, package: str, description: Optional[str] = None) -> bool:
        """Install a package using Homebrew."""
        brew_path = self._get_homebrew_path()
        if not brew_path:
            return False

        desc = description or f"Installation de {package}"

        # Check if already installed
        result = self.runner.run(
            [brew_path, "list", package],
            sudo=False
        )

        if result.success:
            self.cli.print_info(f"{package} est déjà installé via Homebrew")
            return True

        # Install the package
        result = self.runner.run(
            [brew_path, "install", package],
            description=desc,
            sudo=False,
            timeout=600  # 10 minutes timeout
        )

        return result.success

    def install_neovim(self) -> bool:
        """Install Neovim using Homebrew."""
        brew_path = self._get_homebrew_path()
        if not brew_path:
            return False

        version = self.options.neovim_version if self.options else "stable"

        # Check if already installed
        result = self.runner.run(
            [brew_path, "list", "neovim"],
            sudo=False
        )

        if result.success:
            # Upgrade if already installed
            self.cli.print_info("Neovim est déjà installé, mise à jour...")
            result = self.runner.run(
                [brew_path, "upgrade", "neovim"],
                description="Mise à jour de Neovim",
                sudo=False,
                timeout=600
            )
            # Upgrade returns non-zero if already up to date, that's ok
            return True

        # Install based on version preference
        if version == "nightly":
            # Use HEAD for nightly
            result = self.runner.run(
                [brew_path, "install", "--HEAD", "neovim"],
                description="Installation de Neovim (nightly)",
                sudo=False,
                timeout=900  # 15 minutes for build
            )
        else:
            # Stable release
            result = self.runner.run(
                [brew_path, "install", "neovim"],
                description="Installation de Neovim (stable)",
                sudo=False,
                timeout=600
            )

        return result.success

    def install_python_provider(self) -> bool:
        """Install Python provider for Neovim."""
        # Install pynvim
        result = self.runner.run(
            ["pip3", "install", "--user", "pynvim"],
            description="Installation du provider Python (pynvim)",
            sudo=False
        )

        return result.success

    def install_node_provider(self) -> bool:
        """Install Node.js provider for Neovim."""
        if not self.runner.check_command_exists("npm"):
            return False

        result = self.runner.run(
            ["npm", "install", "-g", "neovim"],
            description="Installation du provider Node.js",
            sudo=False
        )

        return result.success

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete installation process for macOS."""
        # Run base installation
        result = super().install(options)

        if result.success:
            # Install providers
            self.cli.print_section("Installation des providers")

            if not self.install_python_provider():
                result.warnings.append("Provider Python non installé")
                self.cli.print_warning("Provider Python non installé")
            else:
                self.cli.print_success("Provider Python installé")

            if self.runner.check_command_exists("npm"):
                if not self.install_node_provider():
                    result.warnings.append("Provider Node.js non installé")
                    self.cli.print_warning("Provider Node.js non installé")
                else:
                    self.cli.print_success("Provider Node.js installé")

        return result
