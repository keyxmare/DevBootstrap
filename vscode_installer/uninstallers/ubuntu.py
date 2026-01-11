"""Ubuntu/Debian-specific VS Code uninstaller."""

import os
import shutil
from .base import BaseUninstaller, UninstallOptions, UninstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class UbuntuUninstaller(BaseUninstaller):
    """Uninstaller for VS Code on Ubuntu/Debian."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the Ubuntu VS Code uninstaller."""
        super().__init__(system_info, cli, runner)

    def get_vscode_paths(self) -> dict[str, str]:
        """Get VS Code related paths for Ubuntu/Debian."""
        home = self.system_info.home_dir
        return {
            "settings": os.path.join(home, ".config", "Code"),
            "extensions": os.path.join(home, ".vscode", "extensions"),
            "cache": os.path.join(home, ".cache", "Code"),
            "crashpad": os.path.join(home, ".config", "Code", "Crashpad"),
        }

    def _detect_install_method(self) -> str:
        """Detect how VS Code was installed."""
        # Check snap
        result = self.runner.run(
            ["snap", "list", "code"],
            sudo=False
        )
        if result.success:
            return "snap"

        # Check apt/dpkg
        result = self.runner.run(
            ["dpkg", "-s", "code"],
            sudo=False
        )
        if result.success:
            return "apt"

        # Check if binary exists
        if self.runner.check_command_exists("code"):
            return "manual"

        return "unknown"

    def _uninstall_apt(self) -> bool:
        """Uninstall VS Code via apt."""
        self.cli.print_info("Desinstallation de VS Code via apt...")

        # Remove package
        result = self.runner.run(
            ["apt-get", "remove", "-y", "code"],
            description="Desinstallation de VS Code",
            sudo=True,
            timeout=120
        )

        if result.success:
            # Purge configuration
            self.runner.run(
                ["apt-get", "purge", "-y", "code"],
                sudo=True
            )
            # Clean up
            self.runner.run(
                ["apt-get", "autoremove", "-y"],
                sudo=True
            )
            self.cli.print_success("VS Code desinstalle via apt")
        else:
            self.cli.print_warning("Echec de la desinstallation via apt")

        return result.success

    def _uninstall_snap(self) -> bool:
        """Uninstall VS Code via snap."""
        self.cli.print_info("Desinstallation de VS Code via snap...")

        result = self.runner.run(
            ["snap", "remove", "code"],
            description="Desinstallation de VS Code",
            sudo=True,
            timeout=120
        )

        if result.success:
            self.cli.print_success("VS Code desinstalle via snap")
        else:
            self.cli.print_warning("Echec de la desinstallation via snap")

        return result.success

    def _remove_microsoft_repo(self) -> bool:
        """Remove Microsoft repository for VS Code."""
        self.cli.print_info("Suppression du repository Microsoft...")

        # Remove repository file
        repo_files = [
            "/etc/apt/sources.list.d/vscode.list",
            "/etc/apt/sources.list.d/microsoft.list",
        ]

        for repo_file in repo_files:
            if os.path.exists(repo_file):
                try:
                    os.remove(repo_file)
                    self.cli.print_success(f"Repository supprime: {repo_file}")
                except PermissionError:
                    self.runner.run(["rm", "-f", repo_file], sudo=True)

        # Remove GPG key
        gpg_files = [
            "/etc/apt/trusted.gpg.d/microsoft.gpg",
            "/etc/apt/keyrings/microsoft.gpg",
            "/usr/share/keyrings/microsoft.gpg",
        ]

        for gpg_file in gpg_files:
            if os.path.exists(gpg_file):
                try:
                    os.remove(gpg_file)
                except PermissionError:
                    self.runner.run(["rm", "-f", gpg_file], sudo=True)

        # Update apt
        self.runner.run(["apt-get", "update"], sudo=True)

        return True

    def uninstall_vscode(self) -> bool:
        """Uninstall VS Code on Ubuntu/Debian."""
        install_method = self._detect_install_method()
        self.cli.print_info(f"Methode d'installation detectee: {install_method}")

        if install_method == "snap":
            return self._uninstall_snap()
        elif install_method == "apt":
            success = self._uninstall_apt()
            # Also remove the Microsoft repository
            self._remove_microsoft_repo()
            return success
        elif install_method == "manual":
            self.cli.print_warning("Installation manuelle detectee")
            self.cli.print_info("Suppression manuelle peut etre necessaire")
            return True

        return True

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process for Ubuntu/Debian."""
        # Run base uninstallation
        result = super().uninstall(options)

        if result.success:
            # Remove .vscode directory
            vscode_dir = os.path.join(self.system_info.home_dir, ".vscode")
            if os.path.exists(vscode_dir):
                if self.remove_directory(vscode_dir, ".vscode"):
                    result.removed_items.append(".vscode")

        return result
