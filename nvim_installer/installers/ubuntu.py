"""Ubuntu/Debian-specific Neovim installer using apt and AppImage/snap."""

import os
import tempfile
from typing import Optional
from .base import BaseInstaller, Dependency, InstallOptions, InstallResult
from ..utils.system import SystemInfo, Architecture
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


class UbuntuInstaller(BaseInstaller):
    """Installer for Ubuntu/Debian using apt and AppImage."""

    # Ubuntu-specific dependencies (using apt package names)
    UBUNTU_DEPENDENCIES: list[Dependency] = [
        Dependency(
            name="git",
            description="Système de contrôle de version",
            check_command="git"
        ),
        Dependency(
            name="curl",
            description="Outil de transfert de données",
            check_command="curl"
        ),
        Dependency(
            name="wget",
            description="Téléchargement de fichiers",
            check_command="wget"
        ),
        Dependency(
            name="build-essential",
            description="Outils de compilation",
            check_command="gcc"
        ),
        Dependency(
            name="nodejs",
            description="Runtime JavaScript (pour LSP, plugins)",
            check_command="node",
            required=False
        ),
        Dependency(
            name="npm",
            description="Gestionnaire de paquets Node.js",
            check_command="npm",
            required=False
        ),
        Dependency(
            name="python3",
            description="Python 3 (pour plugins Python)",
            check_command="python3"
        ),
        Dependency(
            name="python3-pip",
            description="Gestionnaire de paquets Python",
            check_command="pip3"
        ),
        Dependency(
            name="python3-venv",
            description="Environnements virtuels Python",
            check_command="python3",  # No direct command check
            required=False
        ),
        Dependency(
            name="ripgrep",
            description="Recherche ultra-rapide (pour Telescope)",
            check_command="rg"
        ),
        Dependency(
            name="fd-find",
            description="Alternative à find (pour Telescope)",
            check_command="fdfind",  # fd is called fdfind on Ubuntu
            required=False
        ),
        Dependency(
            name="fzf",
            description="Fuzzy finder",
            check_command="fzf",
            required=False
        ),
        Dependency(
            name="unzip",
            description="Décompression de fichiers ZIP",
            check_command="unzip"
        ),
        Dependency(
            name="xclip",
            description="Accès au presse-papiers",
            check_command="xclip",
            required=False
        ),
    ]

    # Neovim AppImage URLs
    NEOVIM_APPIMAGE_URL_STABLE = "https://github.com/neovim/neovim/releases/latest/download/nvim.appimage"
    NEOVIM_APPIMAGE_URL_NIGHTLY = "https://github.com/neovim/neovim/releases/download/nightly/nvim.appimage"

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the Ubuntu installer."""
        super().__init__(system_info, cli, runner)
        self._install_method: Optional[str] = None  # 'apt', 'appimage', 'snap', 'ppa', 'build'

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
            description="Mise à jour de la liste des paquets",
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
            self.cli.print_info(f"{package} est déjà installé")
            return True

        # Install the package
        result = self.runner.run(
            ["apt-get", "install", "-y", package],
            description=desc,
            sudo=True,
            timeout=600
        )

        return result.success

    def _add_neovim_ppa(self) -> bool:
        """Add the Neovim PPA for Ubuntu."""
        # Install software-properties-common if needed
        if not self.runner.check_command_exists("add-apt-repository"):
            result = self.runner.run(
                ["apt-get", "install", "-y", "software-properties-common"],
                description="Installation de software-properties-common",
                sudo=True
            )
            if not result.success:
                return False

        # Add Neovim PPA
        result = self.runner.run(
            ["add-apt-repository", "-y", "ppa:neovim-ppa/unstable"],
            description="Ajout du PPA Neovim",
            sudo=True,
            timeout=120
        )

        if not result.success:
            return False

        # Update apt after adding PPA
        return self.update_package_manager()

    def _install_neovim_apt(self) -> bool:
        """Install Neovim using apt (may be an older version)."""
        return self.install_package("neovim", "Installation de Neovim via apt")

    def _install_neovim_ppa(self) -> bool:
        """Install Neovim from the official PPA (latest version)."""
        if not self._add_neovim_ppa():
            self.cli.print_warning("Impossible d'ajouter le PPA Neovim")
            return False

        return self.install_package("neovim", "Installation de Neovim depuis le PPA")

    def _install_neovim_appimage(self) -> bool:
        """Install Neovim using AppImage."""
        version = self.options.neovim_version if self.options else "stable"

        url = self.NEOVIM_APPIMAGE_URL_NIGHTLY if version == "nightly" else self.NEOVIM_APPIMAGE_URL_STABLE

        # Create installation directory
        install_dir = os.path.expanduser("~/.local/bin")
        self.runner.ensure_directory(install_dir)

        # Download path
        appimage_path = os.path.join(install_dir, "nvim.appimage")

        # Download the AppImage
        self.cli.print_info(f"Téléchargement de Neovim ({version})...")
        if not self.runner.download_file(url, appimage_path, "Téléchargement de l'AppImage"):
            return False

        # Make executable
        result = self.runner.run(
            ["chmod", "+x", appimage_path],
            description="Rendre l'AppImage exécutable",
            sudo=False
        )

        if not result.success:
            return False

        # Try to extract the AppImage (better integration)
        self.cli.print_info("Extraction de l'AppImage...")
        result = self.runner.run(
            [appimage_path, "--appimage-extract"],
            description="Extraction de l'AppImage",
            sudo=False,
            cwd=install_dir
        )

        if result.success:
            # Create symlink to extracted binary
            extracted_nvim = os.path.join(install_dir, "squashfs-root", "usr", "bin", "nvim")
            nvim_link = os.path.join(install_dir, "nvim")

            if os.path.exists(extracted_nvim):
                # Remove old symlink if exists
                if os.path.exists(nvim_link) or os.path.islink(nvim_link):
                    os.remove(nvim_link)

                os.symlink(extracted_nvim, nvim_link)
                self.cli.print_success("AppImage extraite et liée")

                # Remove the original AppImage
                os.remove(appimage_path)
            else:
                # Fallback: use AppImage directly
                nvim_link = os.path.join(install_dir, "nvim")
                if os.path.exists(nvim_link) or os.path.islink(nvim_link):
                    os.remove(nvim_link)
                os.symlink(appimage_path, nvim_link)
        else:
            # AppImage extraction failed, use directly
            nvim_link = os.path.join(install_dir, "nvim")
            if os.path.exists(nvim_link) or os.path.islink(nvim_link):
                os.remove(nvim_link)
            os.symlink(appimage_path, nvim_link)

        # Ensure ~/.local/bin is in PATH
        self._ensure_local_bin_in_path()

        return True

    def _install_neovim_snap(self) -> bool:
        """Install Neovim using snap."""
        if not self.runner.check_command_exists("snap"):
            return False

        version = self.options.neovim_version if self.options else "stable"
        channel = "edge" if version == "nightly" else "stable"

        result = self.runner.run(
            ["snap", "install", "nvim", "--classic", f"--channel={channel}"],
            description=f"Installation de Neovim via snap ({channel})",
            sudo=True,
            timeout=300
        )

        return result.success

    def _ensure_local_bin_in_path(self):
        """Ensure ~/.local/bin is in PATH."""
        local_bin = os.path.expanduser("~/.local/bin")
        current_path = os.environ.get("PATH", "")

        if local_bin not in current_path:
            os.environ["PATH"] = f"{local_bin}:{current_path}"

            # Also add to shell rc file
            shell = os.environ.get("SHELL", "/bin/bash")
            if "zsh" in shell:
                rc_file = os.path.expanduser("~/.zshrc")
            else:
                rc_file = os.path.expanduser("~/.bashrc")

            path_export = f'\nexport PATH="$HOME/.local/bin:$PATH"\n'

            try:
                # Check if already in rc file
                if os.path.exists(rc_file):
                    with open(rc_file, "r") as f:
                        content = f.read()
                    if ".local/bin" not in content:
                        with open(rc_file, "a") as f:
                            f.write(path_export)
                        self.cli.print_info(f"PATH mis à jour dans {rc_file}")
                else:
                    with open(rc_file, "w") as f:
                        f.write(path_export)
            except Exception as e:
                self.cli.print_warning(f"Impossible de mettre à jour {rc_file}: {e}")

    def _install_lazygit(self) -> bool:
        """Install lazygit from GitHub releases."""
        # Check architecture
        arch = self.system_info.architecture
        if arch == Architecture.ARM64:
            arch_str = "arm64"
        else:
            arch_str = "x86_64"

        # Get latest version
        self.cli.print_info("Récupération de la dernière version de lazygit...")

        with tempfile.TemporaryDirectory() as tmpdir:
            # Download latest release info
            result = self.runner.run(
                ["curl", "-s", "https://api.github.com/repos/jesseduffield/lazygit/releases/latest"],
                sudo=False
            )

            if not result.success:
                return False

            # Extract version from JSON (simple parsing)
            import json
            try:
                release_info = json.loads(result.stdout)
                version = release_info["tag_name"].lstrip("v")
            except (json.JSONDecodeError, KeyError):
                self.cli.print_warning("Impossible de déterminer la version de lazygit")
                return False

            # Download URL
            filename = f"lazygit_{version}_Linux_{arch_str}.tar.gz"
            url = f"https://github.com/jesseduffield/lazygit/releases/latest/download/{filename}"
            download_path = os.path.join(tmpdir, filename)

            if not self.runner.download_file(url, download_path, "Téléchargement de lazygit"):
                return False

            # Extract
            result = self.runner.run(
                ["tar", "xzf", download_path, "-C", tmpdir],
                sudo=False
            )

            if not result.success:
                return False

            # Move to ~/.local/bin
            install_dir = os.path.expanduser("~/.local/bin")
            self.runner.ensure_directory(install_dir)

            result = self.runner.run(
                ["mv", os.path.join(tmpdir, "lazygit"), install_dir],
                sudo=False
            )

            return result.success

    def install_neovim(self) -> bool:
        """Install Neovim using the best available method."""
        version = self.options.neovim_version if self.options else "stable"

        # Determine best installation method based on architecture
        arch = self.system_info.architecture

        if arch == Architecture.ARM64:
            # ARM64: PPA or build from source (AppImage doesn't support ARM)
            self.cli.print_info("Système ARM64 détecté")

            # Try PPA first
            self.cli.print_info("Tentative d'installation via PPA...")
            if self._install_neovim_ppa():
                self._install_method = "ppa"
                return True

            # Try snap
            self.cli.print_info("Tentative d'installation via snap...")
            if self._install_neovim_snap():
                self._install_method = "snap"
                return True

            # Fall back to apt (older version)
            self.cli.print_warning("Utilisation de apt (version potentiellement ancienne)")
            if self._install_neovim_apt():
                self._install_method = "apt"
                return True

        else:
            # x86_64: AppImage preferred for latest version
            self.cli.print_info("Système x86_64 détecté")

            # Try AppImage first (always has latest version)
            self.cli.print_info("Tentative d'installation via AppImage...")
            if self._install_neovim_appimage():
                self._install_method = "appimage"
                return True

            # Try PPA
            self.cli.print_info("Tentative d'installation via PPA...")
            if self._install_neovim_ppa():
                self._install_method = "ppa"
                return True

            # Try snap
            self.cli.print_info("Tentative d'installation via snap...")
            if self._install_neovim_snap():
                self._install_method = "snap"
                return True

            # Fall back to apt
            self.cli.print_warning("Utilisation de apt (version potentiellement ancienne)")
            if self._install_neovim_apt():
                self._install_method = "apt"
                return True

        return False

    def install_python_provider(self) -> bool:
        """Install Python provider for Neovim."""
        result = self.runner.run(
            ["pip3", "install", "--user", "--break-system-packages", "pynvim"],
            description="Installation du provider Python (pynvim)",
            sudo=False
        )

        if not result.success:
            # Try without --break-system-packages for older pip
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
            sudo=True
        )

        return result.success

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete installation process for Ubuntu."""
        # Ensure FUSE is available for AppImage
        if self.system_info.architecture != Architecture.ARM64:
            self.install_package("fuse", "Installation de FUSE (pour AppImage)")
            self.install_package("libfuse2", "Installation de libfuse2")

        # Run base installation
        result = super().install(options)

        if result.success:
            # Install additional tools
            self.cli.print_section("Installation des outils additionnels")

            # Install lazygit from GitHub (not in apt)
            if not self.runner.check_command_exists("lazygit"):
                self.cli.print_info("Installation de lazygit...")
                if self._install_lazygit():
                    self.cli.print_success("lazygit installé")
                else:
                    result.warnings.append("lazygit non installé")
                    self.cli.print_warning("lazygit non installé")

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

            # Add install method info
            if self._install_method:
                result.message += f" (méthode: {self._install_method})"

        return result
