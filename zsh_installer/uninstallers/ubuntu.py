"""Ubuntu/Debian-specific Zsh uninstaller."""

import os
from .base import BaseUninstaller, UninstallOptions, UninstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class UbuntuUninstaller(BaseUninstaller):
    """Uninstaller for Zsh on Ubuntu/Debian."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the Ubuntu Zsh uninstaller."""
        super().__init__(system_info, cli, runner)

    def uninstall_zsh(self) -> bool:
        """Uninstall Zsh via apt."""
        # Check if Zsh is installed via apt
        result = self.runner.run(
            ["dpkg", "-s", "zsh"],
            sudo=False
        )

        if not result.success:
            self.cli.print_info("Zsh n'est pas installe via apt")
            return True

        self.cli.print_info("Desinstallation de Zsh via apt...")

        # Remove zsh package
        result = self.runner.run(
            ["apt-get", "remove", "-y", "zsh"],
            description="Desinstallation de Zsh",
            sudo=True,
            timeout=120
        )

        if result.success:
            # Purge configuration
            self.runner.run(
                ["apt-get", "purge", "-y", "zsh"],
                sudo=True
            )
            # Clean up
            self.runner.run(
                ["apt-get", "autoremove", "-y"],
                sudo=True
            )
            self.cli.print_success("Zsh desinstalle")
        else:
            self.cli.print_warning("Echec de la desinstallation de Zsh")

        return result.success

    def uninstall_bash_completion(self) -> bool:
        """Optionally uninstall bash-completion."""
        # We don't uninstall bash-completion as it's commonly used
        self.cli.print_info("bash-completion n'est pas supprime (utilise par d'autres programmes)")
        return True

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process for Ubuntu/Debian."""
        # First, make sure we're not removing Zsh if it's the default shell
        # without restoring bash first
        if options.restore_default_shell:
            # Restore bash as default shell before potentially removing Zsh
            self.cli.print_section("Verification du shell par defaut")
            current_shell = os.environ.get("SHELL", "")

            if "zsh" in current_shell:
                self.cli.print_info("Restauration de bash comme shell par defaut...")
                self.restore_default_shell()

        # Run base uninstallation
        result = super().uninstall(options)

        if result.success:
            # Ask if user wants to also uninstall Zsh itself
            if self.cli.ask_yes_no("Voulez-vous aussi desinstaller Zsh lui-meme?", default=False):
                self.cli.print_section("Desinstallation de Zsh")
                if self.uninstall_zsh():
                    result.removed_items.append("zsh")
                else:
                    result.warnings.append("Zsh non desinstalle")

        return result
