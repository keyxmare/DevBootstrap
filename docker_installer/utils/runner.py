"""Command execution utilities."""

import subprocess
import os
import shutil
from dataclasses import dataclass
from typing import Optional, Callable
from .cli import CLI


@dataclass
class CommandResult:
    """Result of a command execution."""
    success: bool
    stdout: str
    stderr: str
    return_code: int


class CommandRunner:
    """Execute shell commands with logging and error handling."""

    def __init__(self, cli: CLI, use_sudo: bool = False, dry_run: bool = False):
        """Initialize the command runner."""
        self.cli = cli
        self.use_sudo = use_sudo
        self.dry_run = dry_run

    def run(
        self,
        command: list[str],
        description: Optional[str] = None,
        sudo: Optional[bool] = None,
        capture_output: bool = True,
        env: Optional[dict] = None,
        cwd: Optional[str] = None,
        timeout: Optional[int] = None
    ) -> CommandResult:
        """Run a command and return the result."""
        use_sudo = sudo if sudo is not None else self.use_sudo

        # Prepend sudo if needed
        if use_sudo and os.geteuid() != 0:
            command = ["sudo"] + command

        cmd_str = " ".join(command)

        if description:
            self.cli.print_progress(description)

        if self.dry_run:
            self.cli.clear_progress()
            self.cli.print_info(f"[DRY RUN] {cmd_str}")
            return CommandResult(
                success=True,
                stdout="",
                stderr="",
                return_code=0
            )

        try:
            # Merge environment
            run_env = os.environ.copy()
            if env:
                run_env.update(env)

            result = subprocess.run(
                command,
                capture_output=capture_output,
                text=True,
                env=run_env,
                cwd=cwd,
                timeout=timeout
            )

            self.cli.clear_progress()

            return CommandResult(
                success=result.returncode == 0,
                stdout=result.stdout or "",
                stderr=result.stderr or "",
                return_code=result.returncode
            )

        except subprocess.TimeoutExpired:
            self.cli.clear_progress()
            return CommandResult(
                success=False,
                stdout="",
                stderr="Command timed out",
                return_code=-1
            )
        except FileNotFoundError:
            self.cli.clear_progress()
            return CommandResult(
                success=False,
                stdout="",
                stderr=f"Command not found: {command[0]}",
                return_code=-1
            )
        except Exception as e:
            self.cli.clear_progress()
            return CommandResult(
                success=False,
                stdout="",
                stderr=str(e),
                return_code=-1
            )

    def run_interactive(
        self,
        command: list[str],
        description: Optional[str] = None,
        sudo: Optional[bool] = None,
        env: Optional[dict] = None,
        cwd: Optional[str] = None
    ) -> bool:
        """Run a command with interactive output (shows stdout/stderr in real-time)."""
        use_sudo = sudo if sudo is not None else self.use_sudo

        if use_sudo and os.geteuid() != 0:
            command = ["sudo"] + command

        if description:
            self.cli.print_info(description)

        if self.dry_run:
            self.cli.print_info(f"[DRY RUN] {' '.join(command)}")
            return True

        try:
            run_env = os.environ.copy()
            if env:
                run_env.update(env)

            result = subprocess.run(
                command,
                env=run_env,
                cwd=cwd
            )
            return result.returncode == 0

        except Exception as e:
            self.cli.print_error(f"Erreur: {e}")
            return False

    def check_command_exists(self, command: str) -> bool:
        """Check if a command exists in PATH."""
        return shutil.which(command) is not None

    def get_command_path(self, command: str) -> Optional[str]:
        """Get the full path to a command."""
        return shutil.which(command)

    def get_command_version(self, command: str, version_arg: str = "--version") -> Optional[str]:
        """Get the version of a command."""
        result = self.run([command, version_arg], capture_output=True)
        if result.success:
            # Return first line of output
            output = result.stdout.strip() or result.stderr.strip()
            return output.split("\n")[0] if output else None
        return None

    def ensure_directory(self, path: str, mode: int = 0o755) -> bool:
        """Ensure a directory exists, creating it if necessary."""
        try:
            if self.dry_run:
                self.cli.print_info(f"[DRY RUN] mkdir -p {path}")
                return True

            os.makedirs(path, mode=mode, exist_ok=True)
            return True
        except Exception as e:
            self.cli.print_error(f"Impossible de creer {path}: {e}")
            return False

    def download_file(
        self,
        url: str,
        destination: str,
        description: Optional[str] = None
    ) -> bool:
        """Download a file using curl or wget."""
        if description:
            self.cli.print_progress(description)

        # Try curl first
        if self.check_command_exists("curl"):
            result = self.run(
                ["curl", "-fsSL", "-o", destination, url],
                sudo=False
            )
            if result.success:
                self.cli.clear_progress()
                return True

        # Fall back to wget
        if self.check_command_exists("wget"):
            result = self.run(
                ["wget", "-q", "-O", destination, url],
                sudo=False
            )
            if result.success:
                self.cli.clear_progress()
                return True

        self.cli.clear_progress()
        self.cli.print_error("Ni curl ni wget n'est disponible")
        return False
