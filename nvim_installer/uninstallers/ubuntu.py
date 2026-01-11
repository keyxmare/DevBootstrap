"""Ubuntu/Debian-specific Neovim uninstaller."""

import os
from typing import Optional
from .base import BaseUninstaller, UninstallOptions, UninstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class UbuntuUninstaller(BaseUninstaller):
    """Uninstaller for Ubuntu/Debian."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the Ubuntu uninstaller."""
        super().__init__(system_info, cli, runner)
        self._install_method: Optional[str] = None

    def _detect_install_method(self) -> Optional[str]:
        """Detect how Neovim was installed."""
        # Check AppImage
        appimage_path = os.path.expanduser("~/.local/bin/nvim.appimage")
        nvim_link = os.path.expanduser("~/.local/bin/nvim")
        squashfs_path = os.path.expanduser("~/.local/bin/squashfs-root")

        if os.path.exists(appimage_path) or os.path.exists(squashfs_path):
            return "appimage"

        # Check snap
        result = self.runner.run(
            ["snap", "list", "nvim"],
            sudo=False
        )
        if result.success:
            return "snap"

        # Check apt/dpkg
        result = self.runner.run(
            ["dpkg", "-s", "neovim"],
            sudo=False
        )
        if result.success:
            return "apt"

        # Check if binary exists in common locations
        if os.path.exists(nvim_link) or self.runner.check_command_exists("nvim"):
            return "manual"

        return None

    def _uninstall_appimage(self) -> bool:
        """Uninstall Neovim AppImage."""
        local_bin = os.path.expanduser("~/.local/bin")
        paths_to_remove = [
            os.path.join(local_bin, "nvim.appimage"),
            os.path.join(local_bin, "nvim"),
            os.path.join(local_bin, "squashfs-root"),
        ]

        success = True
        for path in paths_to_remove:
            if os.path.exists(path) or os.path.islink(path):
                try:
                    if os.path.islink(path) or os.path.isfile(path):
                        os.remove(path)
                    else:
                        import shutil
                        shutil.rmtree(path)
                    self.cli.print_success(f"Supprime: {path}")
                except Exception as e:
                    self.cli.print_error(f"Erreur lors de la suppression de {path}: {e}")
                    success = False

        return success

    def _uninstall_snap(self) -> bool:
        """Uninstall Neovim snap."""
        result = self.runner.run(
            ["snap", "remove", "nvim"],
            description="Desinstallation de Neovim via snap",
            sudo=True
        )
        return result.success

    def _uninstall_apt(self) -> bool:
        """Uninstall Neovim via apt."""
        result = self.runner.run(
            ["apt-get", "remove", "-y", "neovim"],
            description="Desinstallation de Neovim via apt",
            sudo=True
        )

        if result.success:
            # Also remove config files
            self.runner.run(
                ["apt-get", "purge", "-y", "neovim"],
                description="Purge de la configuration apt",
                sudo=True
            )
            # Clean up
            self.runner.run(
                ["apt-get", "autoremove", "-y"],
                description="Nettoyage des paquets inutilises",
                sudo=True
            )

        return result.success

    def _uninstall_lazygit(self) -> bool:
        """Uninstall lazygit if installed in ~/.local/bin."""
        lazygit_path = os.path.expanduser("~/.local/bin/lazygit")
        if os.path.exists(lazygit_path):
            try:
                os.remove(lazygit_path)
                self.cli.print_success("lazygit supprime")
                return True
            except Exception as e:
                self.cli.print_warning(f"Impossible de supprimer lazygit: {e}")
                return False
        return True

    def uninstall_neovim(self) -> bool:
        """Uninstall Neovim using the detected method."""
        self._install_method = self._detect_install_method()

        if not self._install_method:
            self.cli.print_info("Aucune installation de Neovim detectee")
            return True

        self.cli.print_info(f"Methode d'installation detectee: {self._install_method}")

        if self._install_method == "appimage":
            return self._uninstall_appimage()
        elif self._install_method == "snap":
            return self._uninstall_snap()
        elif self._install_method == "apt":
            return self._uninstall_apt()
        elif self._install_method == "manual":
            nvim_path = self.runner.get_command_path("nvim")
            self.cli.print_warning(f"Installation manuelle detectee: {nvim_path}")
            self.cli.print_info("Suppression manuelle peut etre necessaire")
            return True

        return False

    def uninstall_dependencies(self) -> bool:
        """Optionally uninstall dependencies."""
        # Note: We don't uninstall dependencies by default as they might be
        # used by other applications
        self.cli.print_info("Les dependances (ripgrep, fzf, etc.) ne sont pas supprimees")
        self.cli.print_info("Utilisez 'apt remove <package>' pour les supprimer manuellement")

        # Optionally remove lazygit if it was installed manually
        self._uninstall_lazygit()

        return True

    def uninstall_python_provider(self) -> bool:
        """Uninstall Python provider for Neovim."""
        result = self.runner.run(
            ["pip3", "uninstall", "-y", "pynvim"],
            description="Desinstallation du provider Python (pynvim)",
            sudo=False
        )
        return result.success

    def uninstall_node_provider(self) -> bool:
        """Uninstall Node.js provider for Neovim."""
        if not self.runner.check_command_exists("npm"):
            return True

        result = self.runner.run(
            ["npm", "uninstall", "-g", "neovim"],
            description="Desinstallation du provider Node.js",
            sudo=True
        )
        return result.success

    def uninstall(self, options: UninstallOptions) -> UninstallResult:
        """Run the complete uninstallation process for Ubuntu."""
        # Run base uninstallation
        result = super().uninstall(options)

        if result.success:
            # Uninstall providers
            self.cli.print_section("Desinstallation des providers")

            if self.uninstall_python_provider():
                self.cli.print_success("Provider Python desinstalle")
            else:
                result.warnings.append("Provider Python non desinstalle")

            if self.uninstall_node_provider():
                self.cli.print_success("Provider Node.js desinstalle")
            else:
                result.warnings.append("Provider Node.js non desinstalle")

            # Uninstall dependencies
            self.uninstall_dependencies()

        return result
