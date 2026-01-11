"""Ubuntu/Debian-specific font installer."""

import os
import zipfile
import tempfile
from .base import BaseInstaller, FontInfo, InstallResult
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


# Nerd Fonts download URLs (from GitHub releases)
NERD_FONTS_BASE_URL = "https://github.com/ryanoasis/nerd-fonts/releases/latest/download"

FONT_DOWNLOAD_NAMES = {
    "meslo": "Meslo.zip",
    "fira-code": "FiraCode.zip",
    "jetbrains-mono": "JetBrainsMono.zip",
    "hack": "Hack.zip",
}


class UbuntuInstaller(BaseInstaller):
    """Font installer for Ubuntu/Debian using manual download."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the Ubuntu installer."""
        super().__init__(system_info, cli, runner)

    def _get_fonts_dir(self) -> str:
        """Get the user fonts directory."""
        return self.system_info.get_fonts_dir()

    def _ensure_fonts_dir(self) -> bool:
        """Ensure the fonts directory exists."""
        fonts_dir = self._get_fonts_dir()
        return self.runner.ensure_directory(fonts_dir)

    def check_font_installed(self, font: FontInfo) -> bool:
        """Check if a font is installed by looking for font files."""
        fonts_dir = self._get_fonts_dir()

        # Check if any .ttf or .otf files with the font name exist
        if not os.path.exists(fonts_dir):
            return False

        font_name_lower = font.name.lower().replace(" ", "")
        for filename in os.listdir(fonts_dir):
            filename_lower = filename.lower().replace(" ", "")
            if font_name_lower[:5] in filename_lower and (filename.endswith(".ttf") or filename.endswith(".otf")):
                return True

        return False

    def _download_and_extract_font(self, font: FontInfo) -> bool:
        """Download and extract a Nerd Font."""
        download_name = FONT_DOWNLOAD_NAMES.get(font.id)
        if not download_name:
            self.cli.print_error(f"Police {font.name} non supportee pour le telechargement")
            return False

        url = f"{NERD_FONTS_BASE_URL}/{download_name}"
        fonts_dir = self._get_fonts_dir()

        # Create temp directory for download
        with tempfile.TemporaryDirectory() as temp_dir:
            zip_path = os.path.join(temp_dir, download_name)

            # Download the zip file
            self.cli.print_info(f"Telechargement de {font.name}...")

            if self.runner.dry_run:
                self.cli.print_info(f"[DRY RUN] curl -fsSL -o {zip_path} {url}")
                self.cli.print_info(f"[DRY RUN] unzip {zip_path} -d {fonts_dir}")
                return True

            # Use curl to download
            result = self.runner.run(
                ["curl", "-fsSL", "-o", zip_path, url],
                description=f"Telechargement de {download_name}",
                sudo=False,
                timeout=300
            )

            if not result.success:
                self.cli.print_error(f"Echec du telechargement: {result.stderr}")
                return False

            # Extract the zip file
            try:
                self.cli.print_info(f"Extraction de {font.name}...")
                with zipfile.ZipFile(zip_path, 'r') as zip_ref:
                    # Only extract .ttf and .otf files
                    for file_info in zip_ref.infolist():
                        if file_info.filename.endswith(('.ttf', '.otf')) and not file_info.filename.startswith('__MACOSX'):
                            # Extract to fonts directory with just the filename
                            filename = os.path.basename(file_info.filename)
                            if filename:
                                target_path = os.path.join(fonts_dir, filename)
                                with zip_ref.open(file_info) as source, open(target_path, 'wb') as target:
                                    target.write(source.read())

                return True

            except zipfile.BadZipFile:
                self.cli.print_error("Fichier zip corrompu")
                return False
            except Exception as e:
                self.cli.print_error(f"Erreur lors de l'extraction: {e}")
                return False

    def _update_font_cache(self) -> bool:
        """Update the font cache."""
        self.cli.print_info("Mise a jour du cache des polices...")
        result = self.runner.run(
            ["fc-cache", "-fv"],
            description="Mise a jour du cache",
            sudo=False,
            timeout=60
        )
        return result.success

    def install_font(self, font: FontInfo) -> bool:
        """Install a font by downloading from Nerd Fonts releases."""
        # Ensure fonts directory exists
        if not self._ensure_fonts_dir():
            return False

        # Download and extract
        if not self._download_and_extract_font(font):
            return False

        # Update font cache
        self._update_font_cache()

        return True

    def install(self, options) -> InstallResult:
        """Run the complete installation process for Ubuntu."""
        # Check curl is available
        if not self.runner.check_command_exists("curl"):
            return InstallResult(
                success=False,
                message="curl n'est pas installe. Veuillez installer curl.",
                errors=["curl requis pour telecharger les polices"]
            )

        return super().install(options)
