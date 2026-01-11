"""Uninstaller for the devbootstrap command alias."""

import os
import re
import shutil
from dataclasses import dataclass, field
from typing import Optional


@dataclass
class UninstallResult:
    """Result of an uninstallation."""
    success: bool
    message: str
    removed_items: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class AliasUninstallerApp:
    """Application for uninstalling the devbootstrap command alias."""

    VERSION = "1.0.0"
    COMMAND_NAME = "devbootstrap"
    GITHUB_REPO = "keyxmare/DevBootstrap"

    def __init__(self, dry_run: bool = False, no_interaction: bool = False):
        """Initialize the uninstaller."""
        self.dry_run = dry_run
        self.no_interaction = no_interaction
        self.home_dir = os.path.expanduser("~")

    def _get_shell_rc_files(self) -> list[str]:
        """Get the list of shell RC files to check."""
        rc_files = []

        # Bash
        bashrc = os.path.join(self.home_dir, ".bashrc")
        bash_profile = os.path.join(self.home_dir, ".bash_profile")

        # Zsh
        zshrc = os.path.join(self.home_dir, ".zshrc")

        for rc_file in [zshrc, bashrc, bash_profile]:
            if os.path.exists(rc_file):
                rc_files.append(rc_file)

        return rc_files

    def _remove_alias_from_file(self, rc_file: str) -> bool:
        """Remove the devbootstrap alias/function from a shell RC file."""
        try:
            with open(rc_file, "r") as f:
                content = f.read()

            if self.COMMAND_NAME not in content or "DevBootstrap" not in content:
                return False  # Nothing to remove

            # Remove the function block
            # Pattern to match the devbootstrap function
            function_pattern = r'\n?# DevBootstrap command.*?(?=\n[^#\s]|\n\n|\Z)'
            new_content = re.sub(function_pattern, '', content, flags=re.DOTALL)

            # Remove alias pattern
            alias_pattern = r'\n?# DevBootstrap command\nalias devbootstrap=.*?\n'
            new_content = re.sub(alias_pattern, '\n', new_content)

            # Remove any remaining DevBootstrap comments
            new_content = re.sub(r'\n# DevBootstrap.*\n', '\n', new_content)

            if self.dry_run:
                print(f"[DRY RUN] Would update {rc_file}")
                return True

            if content != new_content:
                with open(rc_file, "w") as f:
                    f.write(new_content)
                return True

            return False

        except Exception as e:
            print(f"Erreur lors de la modification de {rc_file}: {e}")
            return False

    def _remove_bin_script(self) -> bool:
        """Remove the executable script from ~/.local/bin."""
        script_path = os.path.join(self.home_dir, ".local", "bin", self.COMMAND_NAME)

        if not os.path.exists(script_path):
            return False

        if self.dry_run:
            print(f"[DRY RUN] Would remove {script_path}")
            return True

        try:
            os.remove(script_path)
            return True
        except Exception as e:
            print(f"Erreur lors de la suppression de {script_path}: {e}")
            return False

    def _remove_devbootstrap_directory(self) -> bool:
        """Remove the ~/.devbootstrap directory."""
        devbootstrap_dir = os.path.join(self.home_dir, ".devbootstrap")

        if not os.path.exists(devbootstrap_dir):
            return False

        if self.dry_run:
            print(f"[DRY RUN] Would remove {devbootstrap_dir}")
            return True

        try:
            shutil.rmtree(devbootstrap_dir)
            return True
        except Exception as e:
            print(f"Erreur lors de la suppression de {devbootstrap_dir}: {e}")
            return False

    def check_installed(self) -> bool:
        """Check if devbootstrap command is installed."""
        # Check in PATH
        script_path = os.path.join(self.home_dir, ".local", "bin", self.COMMAND_NAME)
        if os.path.exists(script_path):
            return True

        # Check in RC files
        for rc_file in self._get_shell_rc_files():
            try:
                with open(rc_file, "r") as f:
                    content = f.read()
                if self.COMMAND_NAME in content and "DevBootstrap" in content:
                    return True
            except Exception:
                pass

        # Check for .devbootstrap directory
        devbootstrap_dir = os.path.join(self.home_dir, ".devbootstrap")
        if os.path.exists(devbootstrap_dir):
            return True

        return False

    def uninstall(self) -> UninstallResult:
        """Uninstall the devbootstrap command."""
        removed_items = []
        warnings = []

        # Step 1: Remove bin script
        print("\n> Suppression du script executable")
        print("-" * 50)
        if self._remove_bin_script():
            print(f"[OK] Script ~/.local/bin/{self.COMMAND_NAME} supprime")
            removed_items.append("bin_script")
        else:
            print(f"[i] Script ~/.local/bin/{self.COMMAND_NAME} non trouve")

        # Step 2: Remove alias/function from RC files
        print("\n> Suppression des alias shell")
        print("-" * 50)
        for rc_file in self._get_shell_rc_files():
            if self._remove_alias_from_file(rc_file):
                print(f"[OK] Alias supprime de {rc_file}")
                removed_items.append(f"alias_{os.path.basename(rc_file)}")
            else:
                print(f"[i] Pas d'alias dans {rc_file}")

        # Step 3: Remove .devbootstrap directory
        print("\n> Suppression du repertoire .devbootstrap")
        print("-" * 50)
        if self._remove_devbootstrap_directory():
            print("[OK] Repertoire ~/.devbootstrap supprime")
            removed_items.append("devbootstrap_dir")
        else:
            print("[i] Repertoire ~/.devbootstrap non trouve")

        if not removed_items:
            return UninstallResult(
                success=True,
                message="Rien a desinstaller",
                warnings=["DevBootstrap n'etait pas installe"]
            )

        return UninstallResult(
            success=True,
            message="Desinstallation de devbootstrap terminee!",
            removed_items=removed_items,
            warnings=warnings
        )

    def run(self) -> int:
        """Run the uninstallation process."""
        try:
            print()
            print("=" * 60)
            print(f"  Alias Uninstaller v{self.VERSION}")
            print("=" * 60)
            print()

            # Check if installed
            print("> Verification de l'installation")
            print("-" * 50)
            if not self.check_installed():
                print("[i] La commande devbootstrap n'est pas installee")
                return 0

            print("[i] Installation detectee")

            # Confirm uninstallation
            if not self.no_interaction:
                try:
                    response = input("\n? Desinstaller la commande devbootstrap? [O/n]: ").strip().lower()
                    if response and response not in ("o", "oui", "y", "yes", ""):
                        print("[i] Desinstallation annulee")
                        return 0
                except (EOFError, KeyboardInterrupt):
                    print()
                    return 130

            # Run uninstallation
            result = self.uninstall()

            print()
            print("=" * 60)
            if result.success:
                print(f"[OK] {result.message}")
            else:
                print(f"[X] {result.message}")

            if result.warnings:
                print()
                print("[!] Avertissements:")
                for warning in result.warnings:
                    print(f"    - {warning}")

            print()
            print("> Prochaines etapes")
            print("-" * 50)
            print("1. Redemarrer le terminal ou executer:")
            print("   $ source ~/.bashrc  # ou ~/.zshrc")
            print()

            return 0 if result.success else 1

        except KeyboardInterrupt:
            print()
            print("[!] Desinstallation interrompue par l'utilisateur")
            return 130

        except Exception as e:
            print(f"[X] Erreur inattendue: {e}")
            return 1


def main():
    """Main entry point."""
    import argparse

    parser = argparse.ArgumentParser(
        description="Alias Uninstaller - Desinstallation de la commande devbootstrap"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simuler la desinstallation sans effectuer de changements"
    )
    parser.add_argument(
        "-y", "--yes",
        action="store_true",
        help="Mode non-interactif (confirmer automatiquement)"
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"Alias Uninstaller {AliasUninstallerApp.VERSION}"
    )

    args = parser.parse_args()

    app = AliasUninstallerApp(dry_run=args.dry_run, no_interaction=args.yes)
    exit(app.run())


if __name__ == "__main__":
    main()
