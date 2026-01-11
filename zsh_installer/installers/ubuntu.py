"""Ubuntu/Debian-specific Zsh installer using apt."""

import os
from typing import Optional
from .base import BaseInstaller, Dependency, InstallOptions, InstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class UbuntuInstaller(BaseInstaller):
    """Installer for Ubuntu/Debian using apt."""

    # Ubuntu-specific dependencies (using apt package names)
    UBUNTU_DEPENDENCIES: list[Dependency] = [
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
        Dependency(
            name="wget",
            description="Telechargement de fichiers",
            check_command="wget"
        ),
    ]

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the Ubuntu installer."""
        super().__init__(system_info, cli, runner)

    def get_dependencies(self) -> list[Dependency]:
        """Get Ubuntu-specific dependencies."""
        return self.UBUNTU_DEPENDENCIES

    def install_package_manager(self) -> bool:
        """apt is always available on Ubuntu/Debian."""
        if self.runner.check_command_exists("apt-get"):
            self.cli.print_success("apt est disponible")
            return True
        return False

    def update_package_manager(self) -> bool:
        """Update apt package lists."""
        result = self.runner.run(
            ["apt-get", "update"],
            description="Mise a jour de la liste des paquets",
            sudo=True,
            timeout=300
        )
        return result.success

    def install_package(self, package: str, description: Optional[str] = None) -> bool:
        """Install a package using apt."""
        desc = description or f"Installation de {package}"

        # Check if already installed using dpkg
        result = self.runner.run(
            ["dpkg", "-s", package],
            sudo=False
        )

        if result.success:
            self.cli.print_info(f"{package} est deja installe")
            return True

        # Install the package
        result = self.runner.run(
            ["apt-get", "install", "-y", package],
            description=desc,
            sudo=True,
            timeout=600
        )

        return result.success

    def install_zsh(self) -> bool:
        """Install Zsh using apt."""
        return self.install_package("zsh", "Installation de Zsh")

    def install_bash_completion(self) -> bool:
        """Install bash-completion package."""
        # Install bash-completion
        if not self.install_package("bash-completion", "Installation de bash-completion"):
            return False

        # Configure bash-completion
        self._configure_bash_completion()
        return True

    def _configure_bash_completion(self) -> bool:
        """Configure bash-completion in .bashrc."""
        bashrc_path = os.path.join(self.system_info.home_dir, ".bashrc")

        # Bash completion configuration for Ubuntu/Debian
        bash_completion_config = '''
# Enable bash completion
if ! shopt -oq posix; then
    if [ -f /usr/share/bash-completion/bash_completion ]; then
        . /usr/share/bash-completion/bash_completion
    elif [ -f /etc/bash_completion ]; then
        . /etc/bash_completion
    fi
fi
'''

        try:
            # Check if already configured
            if os.path.exists(bashrc_path):
                with open(bashrc_path, "r") as f:
                    content = f.read()
                if "bash_completion" in content or "bash-completion" in content:
                    self.cli.print_info("bash-completion deja configure")
                    return True

            # Add to .bashrc
            if self.runner.dry_run:
                self.cli.print_info(f"[DRY RUN] Configuration de bash-completion dans {bashrc_path}")
                return True

            with open(bashrc_path, "a") as f:
                f.write(bash_completion_config)

            self.cli.print_success("bash-completion configure dans .bashrc")
            return True

        except Exception as e:
            self.cli.print_error(f"Erreur lors de la configuration de bash-completion: {e}")
            return False

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete installation process for Ubuntu."""
        # Run base installation
        result = super().install(options)

        if result.success:
            # Ensure Zsh is in /etc/shells
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
