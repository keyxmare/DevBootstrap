"""macOS-specific VS Code installer using Homebrew."""

import os
from typing import Optional
from .base import BaseInstaller, InstallOptions, InstallResult
from ..utils.system import SystemInfo, Architecture
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class MacOSInstaller(BaseInstaller):
    """Installer for VS Code on macOS using Homebrew."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the macOS VS Code installer."""
        super().__init__(system_info, cli, runner)
        self._homebrew_path: Optional[str] = None

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

    def _install_homebrew(self) -> bool:
        """Install Homebrew if not present."""
        if self._get_homebrew_path():
            self._ensure_homebrew_in_path()
            self.cli.print_success("Homebrew est deja installe")
            return True

        self.cli.print_info("Installation de Homebrew...")

        # Homebrew installation script
        install_script = '/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'

        result = self.runner.run_interactive(
            ["/bin/bash", "-c", install_script],
            description="Installation de Homebrew",
            sudo=False
        )

        if not result:
            return False

        self._ensure_homebrew_in_path()
        return self._get_homebrew_path() is not None

    def check_existing_installation(self) -> bool:
        """Check if VS Code is already installed."""
        # Check for 'code' command
        if self.runner.check_command_exists("code"):
            return True

        # Check if VS Code app exists
        vscode_app_paths = [
            "/Applications/Visual Studio Code.app",
            os.path.expanduser("~/Applications/Visual Studio Code.app")
        ]

        for path in vscode_app_paths:
            if os.path.exists(path):
                return True

        return False

    def _get_code_command(self) -> Optional[str]:
        """Get the VS Code CLI command."""
        # Check for 'code' in PATH
        if self.runner.check_command_exists("code"):
            return "code"

        # Try VS Code's shell command location
        vscode_cli_paths = [
            "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
            os.path.expanduser("~/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code")
        ]

        for path in vscode_cli_paths:
            if os.path.exists(path):
                return path

        return None

    def install_vscode(self) -> bool:
        """Install VS Code using Homebrew Cask."""
        # Ensure Homebrew is available
        if not self._install_homebrew():
            self.cli.print_error("Homebrew est requis pour installer VS Code")
            return False

        brew_path = self._get_homebrew_path()

        # Check if VS Code is already installed via brew
        result = self.runner.run(
            [brew_path, "list", "--cask", "visual-studio-code"],
            sudo=False
        )

        if result.success:
            self.cli.print_info("VS Code est deja installe via Homebrew")
            # Try to upgrade
            result = self.runner.run(
                [brew_path, "upgrade", "--cask", "visual-studio-code"],
                description="Mise a jour de VS Code",
                sudo=False,
                timeout=600
            )
            self._setup_code_command()
            return True

        # Install VS Code
        self.cli.print_info("Installation de VS Code via Homebrew...")
        result = self.runner.run(
            [brew_path, "install", "--cask", "visual-studio-code"],
            description="Installation de VS Code",
            sudo=False,
            timeout=600
        )

        if not result.success:
            self.cli.print_error("Echec de l'installation de VS Code")
            self.cli.print_info("Vous pouvez telecharger VS Code manuellement:")
            self.cli.print_info("https://code.visualstudio.com/download")
            return False

        self.cli.print_success("VS Code installe")

        # Setup 'code' command in PATH
        self._setup_code_command()

        return True

    def _setup_code_command(self) -> bool:
        """Setup the 'code' command in PATH."""
        if self.runner.check_command_exists("code"):
            self.cli.print_success("Commande 'code' deja disponible")
            return True

        self.cli.print_info("Configuration de la commande 'code'...")

        # VS Code has a built-in command to install shell command
        vscode_cli = "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code"

        if os.path.exists(vscode_cli):
            # Create symlink in /usr/local/bin
            target = "/usr/local/bin/code"

            # Ensure /usr/local/bin exists
            self.runner.run(["mkdir", "-p", "/usr/local/bin"], sudo=True)

            # Remove existing symlink if any
            if os.path.exists(target) or os.path.islink(target):
                self.runner.run(["rm", "-f", target], sudo=True)

            # Create new symlink
            result = self.runner.run(
                ["ln", "-s", vscode_cli, target],
                sudo=True
            )

            if result.success:
                self.cli.print_success("Commande 'code' configuree")
                return True

        self.cli.print_warning("Impossible de configurer la commande 'code' automatiquement")
        self.cli.print_info("Vous pouvez l'ajouter manuellement depuis VS Code:")
        self.cli.print_info("  Cmd+Shift+P > 'Shell Command: Install code command'")
        return False

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete VS Code installation for macOS."""
        result = super().install(options)

        if result.success:
            self.cli.print_section("Informations supplementaires")
            self.cli.print_info("Pour lancer VS Code depuis le terminal:")
            self.cli.print("  code .              # Ouvrir le dossier courant")
            self.cli.print("  code fichier.txt    # Ouvrir un fichier")
            self.cli.print()
            self.cli.print_info("Extensions recommandees:")
            self.cli.print("  code --install-extension <extension-id>")

        return result
