"""Command execution utilities."""

import subprocess
import os
import shutil
from dataclasses import dataclass
from typing import Optional
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

    def check_command_exists(self, command: str) -> bool:
        """Check if a command exists in PATH."""
        return shutil.which(command) is not None

    def get_command_path(self, command: str) -> Optional[str]:
        """Get the full path to a command."""
        return shutil.which(command)

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
