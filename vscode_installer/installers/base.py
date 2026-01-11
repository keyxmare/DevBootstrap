"""Base installer class for VS Code installation."""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Optional
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


@dataclass
class InstallOptions:
    """Options for the VS Code installation process."""
    install_extensions: bool = True
    default_extensions: list[str] = field(default_factory=lambda: [
        "ms-python.python",
        "esbenp.prettier-vscode",
        "dbaeumer.vscode-eslint",
    ])


@dataclass
class InstallResult:
    """Result of a VS Code installation."""
    success: bool
    message: str
    vscode_path: Optional[str] = None
    vscode_version: Optional[str] = None
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class BaseInstaller(ABC):
    """Abstract base class for platform-specific VS Code installers."""

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the installer."""
        self.system_info = system_info
        self.cli = cli
        self.runner = runner
        self.options: Optional[InstallOptions] = None

    @abstractmethod
    def check_existing_installation(self) -> bool:
        """Check if VS Code is already installed."""
        pass

    @abstractmethod
    def install_vscode(self) -> bool:
        """Install VS Code."""
        pass

    def install_extensions(self, extensions: list[str]) -> bool:
        """Install VS Code extensions."""
        if not extensions:
            return True

        self.cli.print_info("Installation des extensions...")

        code_cmd = self._get_code_command()
        if not code_cmd:
            self.cli.print_warning("Commande 'code' non trouvee")
            return False

        success = True
        for ext in extensions:
            self.cli.print_progress(f"Installation de {ext}")
            result = self.runner.run(
                [code_cmd, "--install-extension", ext, "--force"],
                sudo=False,
                timeout=120
            )
            self.cli.clear_progress()

            if result.success:
                self.cli.print_success(f"{ext}")
            else:
                self.cli.print_warning(f"{ext} - echec")
                success = False

        return success

    def _get_code_command(self) -> Optional[str]:
        """Get the VS Code CLI command."""
        # Check for 'code' in PATH
        if self.runner.check_command_exists("code"):
            return "code"
        return None

    def verify_installation(self) -> Optional[str]:
        """Verify VS Code installation and return version."""
        code_cmd = self._get_code_command()
        if not code_cmd:
            return None

        result = self.runner.run([code_cmd, "--version"], sudo=False, timeout=10)
        if result.success:
            return result.stdout.strip().split("\n")[0]
        return None

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete VS Code installation process."""
        self.options = options
        errors = []
        warnings = []

        # Step 1: Check existing installation
        self.cli.print_section("Verification de l'installation existante")
        if self.check_existing_installation():
            version = self.verify_installation()
            self.cli.print_info(f"VS Code est deja installe: {version}")

            if not self.cli.ask_yes_no("Voulez-vous reinstaller/mettre a jour VS Code?", default=False):
                return InstallResult(
                    success=True,
                    message="VS Code est deja installe",
                    vscode_path=self.runner.get_command_path("code"),
                    vscode_version=version
                )

        # Step 2: Install VS Code
        self.cli.print_section("Installation de VS Code")
        if not self.install_vscode():
            return InstallResult(
                success=False,
                message="Echec de l'installation de VS Code",
                errors=["VS Code installation failed"]
            )

        # Step 3: Install extensions if requested
        if options.install_extensions and options.default_extensions:
            self.cli.print_section("Installation des extensions")
            if not self.install_extensions(options.default_extensions):
                warnings.append("Certaines extensions n'ont pas pu etre installees")

        # Step 4: Verify installation
        self.cli.print_section("Verification de l'installation")
        version = self.verify_installation()

        if not version:
            return InstallResult(
                success=False,
                message="VS Code installe mais 'code' non trouve dans le PATH",
                errors=["VS Code CLI not found in PATH after installation"],
                warnings=warnings
            )

        vscode_path = self.runner.get_command_path("code")

        return InstallResult(
            success=True,
            message="Installation de VS Code terminee avec succes!",
            vscode_path=vscode_path,
            vscode_version=version,
            warnings=warnings
        )
