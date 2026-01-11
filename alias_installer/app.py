"""Main application class for Alias Installer."""

import sys
import os
import platform
from typing import Optional
from dataclasses import dataclass


class Colors:
    """ANSI color codes for terminal output."""
    RESET = "\033[0m"
    BOLD = "\033[1m"
    DIM = "\033[2m"
    RED = "\033[31m"
    GREEN = "\033[32m"
    YELLOW = "\033[33m"
    BLUE = "\033[34m"
    MAGENTA = "\033[35m"
    CYAN = "\033[36m"

    @classmethod
    def disable(cls):
        """Disable all colors."""
        for attr in dir(cls):
            if not attr.startswith('_') and attr.isupper():
                setattr(cls, attr, "")


@dataclass
class InstallResult:
    """Result of an installation."""
    success: bool
    message: str
    warnings: list[str] = None

    def __post_init__(self):
        if self.warnings is None:
            self.warnings = []


class AliasInstallerApp:
    """Application for installing the devbootstrap command alias."""

    VERSION = "1.0.0"

    # GitHub repository URL (à personnaliser)
    GITHUB_REPO = "keyxmare/DevBootstrap"

    # Command that will be installed
    COMMAND_NAME = "devbootstrap"

    # The curl command that runs the installer
    CURL_COMMAND = 'bash -c "$(curl -fsSL https://raw.githubusercontent.com/{repo}/main/install.sh)"'

    def __init__(self, dry_run: bool = False):
        """Initialize the application."""
        self.dry_run = dry_run
        self.use_colors = sys.stdout.isatty()
        if not self.use_colors:
            Colors.disable()
        self.home_dir = os.path.expanduser("~")
        self.os_type = self._detect_os()

    def _detect_os(self) -> str:
        """Detect the operating system."""
        system = platform.system().lower()
        if system == "darwin":
            return "macos"
        elif system == "linux":
            return "linux"
        return "unknown"

    def print(self, message: str = "", end: str = "\n"):
        """Print a message."""
        print(message, end=end, flush=True)

    def print_header(self, title: str):
        """Print a styled header."""
        width = max(60, len(title) + 4)
        border = "=" * width
        self.print()
        self.print(f"{Colors.CYAN}{Colors.BOLD}+{border}+{Colors.RESET}")
        self.print(f"{Colors.CYAN}{Colors.BOLD}|{Colors.RESET} {title.center(width - 2)} {Colors.CYAN}{Colors.BOLD}|{Colors.RESET}")
        self.print(f"{Colors.CYAN}{Colors.BOLD}+{border}+{Colors.RESET}")
        self.print()

    def print_section(self, title: str):
        """Print a section header."""
        self.print()
        self.print(f"{Colors.BLUE}{Colors.BOLD}> {title}{Colors.RESET}")
        self.print(f"{Colors.DIM}{'-' * 50}{Colors.RESET}")

    def print_success(self, message: str):
        """Print a success message."""
        self.print(f"{Colors.GREEN}[OK]{Colors.RESET} {message}")

    def print_error(self, message: str):
        """Print an error message."""
        self.print(f"{Colors.RED}[X]{Colors.RESET} {message}")

    def print_warning(self, message: str):
        """Print a warning message."""
        self.print(f"{Colors.YELLOW}[!]{Colors.RESET} {message}")

    def print_info(self, message: str):
        """Print an info message."""
        self.print(f"{Colors.CYAN}[i]{Colors.RESET} {message}")

    def ask_yes_no(self, question: str, default: bool = True) -> bool:
        """Ask a yes/no question."""
        default_str = "O/n" if default else "o/N"
        prompt = f"{Colors.YELLOW}?{Colors.RESET} {question} [{default_str}]: "

        while True:
            try:
                response = input(prompt).strip().lower()
                if not response:
                    return default
                if response in ("o", "oui", "y", "yes"):
                    return True
                elif response in ("n", "non", "no"):
                    return False
                else:
                    self.print_warning("Repondez par 'o' (oui) ou 'n' (non)")
            except EOFError:
                return default
            except KeyboardInterrupt:
                self.print()
                raise

    def _get_shell_rc_files(self) -> list[str]:
        """Get the list of shell RC files to update."""
        rc_files = []

        # Bash
        bashrc = os.path.join(self.home_dir, ".bashrc")
        bash_profile = os.path.join(self.home_dir, ".bash_profile")

        # Zsh
        zshrc = os.path.join(self.home_dir, ".zshrc")

        # Check which files exist or should be created
        if os.path.exists(zshrc):
            rc_files.append(zshrc)

        if os.path.exists(bashrc):
            rc_files.append(bashrc)
        elif os.path.exists(bash_profile):
            rc_files.append(bash_profile)

        # If no RC file exists, create based on current shell
        if not rc_files:
            current_shell = os.environ.get("SHELL", "/bin/bash")
            if "zsh" in current_shell:
                rc_files.append(zshrc)
            else:
                rc_files.append(bashrc)

        return rc_files

    def _get_alias_line(self) -> str:
        """Get the alias line to add to shell RC files."""
        curl_cmd = self.CURL_COMMAND.format(repo=self.GITHUB_REPO)
        return f'\n# DevBootstrap command\nalias {self.COMMAND_NAME}=\'{curl_cmd}\'\n'

    def _get_function_line(self) -> str:
        """Get a shell function that syncs silently then runs locally."""
        return f'''
# DevBootstrap command - syncs silently then runs locally
{self.COMMAND_NAME}() {{
    local INSTALL_DIR="${{HOME}}/.devbootstrap"
    local REPO_URL="https://github.com/{self.GITHUB_REPO}"

    # Sync silently (ignore errors)
    if command -v git &> /dev/null; then
        if [ -d "$INSTALL_DIR/.git" ]; then
            (cd "$INSTALL_DIR" && git fetch origin main --quiet && git reset --hard origin/main --quiet) &>/dev/null
        else
            (rm -rf "$INSTALL_DIR" && git clone --depth=1 --quiet "$REPO_URL" "$INSTALL_DIR") &>/dev/null
        fi
    fi

    # Run bootstrap (or show error if not installed)
    if [ -d "$INSTALL_DIR" ] && [ -f "$INSTALL_DIR/bootstrap/__main__.py" ]; then
        (cd "$INSTALL_DIR" && python3 -m bootstrap "$@")
    else
        echo "DevBootstrap n'est pas installe. Executez: curl -fsSL https://raw.githubusercontent.com/{self.GITHUB_REPO}/main/install.sh | bash"
        return 1
    fi
}}
'''

    def _check_existing_alias(self, rc_file: str) -> bool:
        """Check if the alias already exists in an RC file."""
        try:
            if os.path.exists(rc_file):
                with open(rc_file, "r") as f:
                    content = f.read()
                return self.COMMAND_NAME in content and "DevBootstrap" in content
        except Exception:
            pass
        return False

    def _add_alias_to_file(self, rc_file: str) -> bool:
        """Add the alias to a shell RC file."""
        try:
            if self.dry_run:
                self.print_info(f"[DRY RUN] Ajout de l'alias dans {rc_file}")
                return True

            # Create file if it doesn't exist
            if not os.path.exists(rc_file):
                with open(rc_file, "w") as f:
                    f.write(f"# Shell configuration\n")

            # Add the function (more flexible than alias)
            with open(rc_file, "a") as f:
                f.write(self._get_function_line())

            return True

        except Exception as e:
            self.print_error(f"Erreur lors de l'ecriture dans {rc_file}: {e}")
            return False

    def _create_bin_script(self) -> bool:
        """Create an executable script in ~/.local/bin."""
        bin_dir = os.path.join(self.home_dir, ".local", "bin")
        script_path = os.path.join(bin_dir, self.COMMAND_NAME)

        try:
            if self.dry_run:
                self.print_info(f"[DRY RUN] Creation du script {script_path}")
                return True

            # Create bin directory if it doesn't exist
            os.makedirs(bin_dir, exist_ok=True)

            # Create the script that syncs silently then runs locally
            script_content = f'''#!/bin/bash
# DevBootstrap - Installation automatique de l'environnement de developpement
# https://github.com/{self.GITHUB_REPO}

INSTALL_DIR="${{HOME}}/.devbootstrap"
REPO_URL="https://github.com/{self.GITHUB_REPO}"

# Sync repository silently (ignore errors, continue with local version)
sync_repo() {{
    if command -v git &> /dev/null; then
        if [ -d "$INSTALL_DIR/.git" ]; then
            # Update existing repo
            cd "$INSTALL_DIR" && git fetch origin main --quiet 2>/dev/null && git reset --hard origin/main --quiet 2>/dev/null
        else
            # Clone if not exists
            rm -rf "$INSTALL_DIR" 2>/dev/null
            git clone --depth=1 --quiet "$REPO_URL" "$INSTALL_DIR" 2>/dev/null
        fi
    fi
}}

# Sync in background (silent)
sync_repo &>/dev/null

# Wait for sync to complete (max 5 seconds)
wait

# Check if we have a local installation
if [ ! -d "$INSTALL_DIR" ] || [ ! -f "$INSTALL_DIR/bootstrap/__main__.py" ]; then
    echo "DevBootstrap n'est pas installe. Telechargement..."
    if command -v git &> /dev/null; then
        git clone --depth=1 "$REPO_URL" "$INSTALL_DIR"
    else
        echo "Erreur: git est requis pour installer DevBootstrap"
        exit 1
    fi
fi

# Run bootstrap
cd "$INSTALL_DIR"
python3 -m bootstrap "$@"
'''

            with open(script_path, "w") as f:
                f.write(script_content)

            # Make executable
            os.chmod(script_path, 0o755)

            return True

        except Exception as e:
            self.print_error(f"Erreur lors de la creation du script: {e}")
            return False

    def _ensure_local_bin_in_path(self, rc_files: list[str]) -> bool:
        """Ensure ~/.local/bin is in PATH."""
        local_bin = os.path.join(self.home_dir, ".local", "bin")
        path_line = f'\n# Add ~/.local/bin to PATH\nexport PATH="$HOME/.local/bin:$PATH"\n'

        for rc_file in rc_files:
            try:
                if os.path.exists(rc_file):
                    with open(rc_file, "r") as f:
                        content = f.read()
                    if ".local/bin" in content:
                        continue  # Already in PATH

                if self.dry_run:
                    self.print_info(f"[DRY RUN] Ajout de ~/.local/bin au PATH dans {rc_file}")
                    continue

                with open(rc_file, "a") as f:
                    f.write(path_line)

            except Exception as e:
                self.print_warning(f"Impossible de modifier {rc_file}: {e}")

        return True

    def check_existing_installation(self) -> bool:
        """Check if devbootstrap command is already installed."""
        # Check in PATH
        import shutil
        if shutil.which(self.COMMAND_NAME):
            return True

        # Check in RC files
        for rc_file in self._get_shell_rc_files():
            if self._check_existing_alias(rc_file):
                return True

        return False

    def install(self) -> InstallResult:
        """Install the devbootstrap command."""
        warnings = []

        # Step 1: Create executable script in ~/.local/bin
        self.print_section("Creation du script executable")
        if not self._create_bin_script():
            return InstallResult(
                success=False,
                message="Echec de la creation du script"
            )
        self.print_success(f"Script cree dans ~/.local/bin/{self.COMMAND_NAME}")

        # Step 2: Ensure ~/.local/bin is in PATH
        self.print_section("Configuration du PATH")
        rc_files = self._get_shell_rc_files()
        self._ensure_local_bin_in_path(rc_files)
        self.print_success("PATH configure")

        # Step 3: Add alias/function to shell RC files (as backup)
        self.print_section("Configuration des alias shell")
        for rc_file in rc_files:
            if self._check_existing_alias(rc_file):
                self.print_info(f"Alias deja present dans {rc_file}")
            else:
                if self._add_alias_to_file(rc_file):
                    self.print_success(f"Fonction ajoutee dans {rc_file}")
                else:
                    warnings.append(f"Impossible d'ajouter l'alias dans {rc_file}")

        return InstallResult(
            success=True,
            message="Commande devbootstrap installee avec succes!",
            warnings=warnings
        )

    def show_final_instructions(self):
        """Show final instructions after installation."""
        self.print_section("Prochaines etapes")

        self.print()
        self.print(f"{Colors.BOLD}1. Redemarrer le terminal ou executer:{Colors.RESET}")
        self.print(f"   $ source ~/.bashrc  # ou ~/.zshrc")
        self.print()

        self.print(f"{Colors.BOLD}2. Utiliser la commande:{Colors.RESET}")
        self.print(f"   $ {self.COMMAND_NAME}")
        self.print()

        self.print(f"{Colors.BOLD}Comportement:{Colors.RESET}")
        self.print(f"   - Synchronise automatiquement avec GitHub (silencieux)")
        self.print(f"   - En cas d'echec, utilise la version locale")
        self.print(f"   - Lance le menu DevBootstrap")
        self.print()

    def run(self) -> int:
        """Run the installation process."""
        try:
            self.print_header(f"Alias Installer v{self.VERSION}")

            # Show info
            self.print_info(f"Systeme detecte: {self.os_type}")
            self.print_info(f"Commande a installer: {self.COMMAND_NAME}")
            self.print()

            # Check existing installation
            self.print_section("Verification de l'installation existante")
            if self.check_existing_installation():
                self.print_info(f"La commande '{self.COMMAND_NAME}' semble deja installee")
                if not self.ask_yes_no("Voulez-vous reinstaller?", default=False):
                    self.print_info("Installation annulee")
                    return 0

            # Confirm installation
            if not self.ask_yes_no(f"Installer la commande '{self.COMMAND_NAME}'?", default=True):
                self.print_info("Installation annulee")
                return 0

            # Run installation
            result = self.install()

            if result.success:
                self.print_section("Installation terminee")
                self.print_success(result.message)

                if result.warnings:
                    self.print()
                    self.print_warning("Avertissements:")
                    for warning in result.warnings:
                        self.print(f"  - {warning}")

                self.show_final_instructions()
                return 0
            else:
                self.print_section("Echec de l'installation")
                self.print_error(result.message)
                return 1

        except KeyboardInterrupt:
            self.print()
            self.print_warning("Installation interrompue par l'utilisateur")
            return 130

        except Exception as e:
            self.print_error(f"Erreur inattendue: {e}")
            if self.dry_run:
                import traceback
                traceback.print_exc()
            return 1


def main():
    """Main entry point."""
    import argparse

    parser = argparse.ArgumentParser(
        description="Alias Installer - Installation de la commande devbootstrap"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simuler l'installation sans effectuer de changements"
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"Alias Installer {AliasInstallerApp.VERSION}"
    )

    args = parser.parse_args()

    app = AliasInstallerApp(dry_run=args.dry_run)
    sys.exit(app.run())


if __name__ == "__main__":
    main()
