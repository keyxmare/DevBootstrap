"""macOS-specific Zsh installer using Homebrew."""

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
            description="Systeme de controle de version",
            check_command="git"
        ),
        Dependency(
            name="curl",
            description="Outil de transfert de donnees",
            check_command="curl"
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
            self.cli.print_success("Homebrew est deja installe")
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
            description="Mise a jour de Homebrew",
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
            self.cli.print_info(f"{package} est deja installe via Homebrew")
            return True

        # Install the package
        result = self.runner.run(
            [brew_path, "install", package],
            description=desc,
            sudo=False,
            timeout=600  # 10 minutes timeout
        )

        return result.success

    def install_zsh(self) -> bool:
        """Install Zsh using Homebrew."""
        # Note: macOS comes with zsh pre-installed since Catalina
        # but we can still install the latest version via Homebrew

        if self.runner.check_command_exists("zsh"):
            version = self.runner.get_command_version("zsh")
            self.cli.print_info(f"Zsh est deja installe: {version}")

            # Ask if user wants to install Homebrew version for latest
            if self.cli.ask_yes_no("Installer la derniere version via Homebrew?", default=False):
                return self.install_package("zsh", "Installation de Zsh (Homebrew)")
            return True

        return self.install_package("zsh", "Installation de Zsh")

    def install_bash_completion(self) -> bool:
        """Install bash-completion package."""
        brew_path = self._get_homebrew_path()
        if not brew_path:
            return False

        # Check if bash-completion@2 is installed (preferred for bash 4+)
        result = self.runner.run([brew_path, "list", "bash-completion@2"], sudo=False)
        if result.success:
            self.cli.print_info("bash-completion@2 est deja installe")
            return True

        # Try to install bash-completion@2
        result = self.runner.run(
            [brew_path, "install", "bash-completion@2"],
            description="Installation de bash-completion@2",
            sudo=False,
            timeout=300
        )

        if result.success:
            self._configure_bash_completion()
            return True

        # Fall back to bash-completion (v1)
        result = self.runner.run(
            [brew_path, "install", "bash-completion"],
            description="Installation de bash-completion",
            sudo=False,
            timeout=300
        )

        if result.success:
            self._configure_bash_completion()

        return result.success

    def _configure_bash_completion(self) -> bool:
        """Configure bash-completion in .bashrc or .bash_profile."""
        bashrc_path = os.path.join(self.system_info.home_dir, ".bashrc")
        bash_profile_path = os.path.join(self.system_info.home_dir, ".bash_profile")

        # Bash completion configuration for macOS
        bash_completion_config = '''
# Bash completion (Homebrew)
if type brew &>/dev/null; then
    HOMEBREW_PREFIX="$(brew --prefix)"
    if [[ -r "${HOMEBREW_PREFIX}/etc/profile.d/bash_completion.sh" ]]; then
        source "${HOMEBREW_PREFIX}/etc/profile.d/bash_completion.sh"
    else
        for COMPLETION in "${HOMEBREW_PREFIX}/etc/bash_completion.d/"*; do
            [[ -r "${COMPLETION}" ]] && source "${COMPLETION}"
        done
    fi
fi
'''

        try:
            # Check if already configured
            for rc_file in [bashrc_path, bash_profile_path]:
                if os.path.exists(rc_file):
                    with open(rc_file, "r") as f:
                        content = f.read()
                    if "bash_completion" in content:
                        self.cli.print_info("bash-completion deja configure")
                        return True

            # Add to .bashrc (create if doesn't exist)
            if self.runner.dry_run:
                self.cli.print_info(f"[DRY RUN] Configuration de bash-completion dans {bashrc_path}")
                return True

            with open(bashrc_path, "a") as f:
                f.write(bash_completion_config)

            # Also source .bashrc from .bash_profile if not already done
            if os.path.exists(bash_profile_path):
                with open(bash_profile_path, "r") as f:
                    profile_content = f.read()
                if ".bashrc" not in profile_content:
                    with open(bash_profile_path, "a") as f:
                        f.write('\n# Source .bashrc\n[[ -r ~/.bashrc ]] && source ~/.bashrc\n')

            self.cli.print_success("bash-completion configure dans .bashrc")
            return True

        except Exception as e:
            self.cli.print_error(f"Erreur lors de la configuration de bash-completion: {e}")
            return False

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete installation process for macOS."""
        # Run base installation
        result = super().install(options)

        if result.success:
            # Add Zsh to /etc/shells if needed (macOS specific)
            self._ensure_zsh_in_shells()

        return result

    def _ensure_zsh_in_shells(self) -> bool:
        """Ensure Zsh is listed in /etc/shells."""
        zsh_path = self.runner.get_command_path("zsh")
        if not zsh_path:
            return False

        try:
            with open("/etc/shells", "r") as f:
                shells = f.read()

            if zsh_path in shells:
                return True

            # Need to add zsh to /etc/shells
            self.cli.print_info(f"Ajout de {zsh_path} a /etc/shells")

            if self.runner.dry_run:
                self.cli.print_info(f"[DRY RUN] echo '{zsh_path}' >> /etc/shells")
                return True

            result = self.runner.run(
                ["sh", "-c", f"echo '{zsh_path}' >> /etc/shells"],
                sudo=True,
                description="Ajout de Zsh a /etc/shells"
            )

            return result.success

        except Exception as e:
            self.cli.print_warning(f"Impossible de modifier /etc/shells: {e}")
            return False
