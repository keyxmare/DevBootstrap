"""Command-line interface utilities."""

import sys
from typing import Optional, Callable, Any
from enum import Enum


class Colors:
    """ANSI color codes for terminal output."""
    RESET = "\033[0m"
    BOLD = "\033[1m"
    DIM = "\033[2m"

    # Colors
    RED = "\033[31m"
    GREEN = "\033[32m"
    YELLOW = "\033[33m"
    BLUE = "\033[34m"
    MAGENTA = "\033[35m"
    CYAN = "\033[36m"
    WHITE = "\033[37m"

    # Background
    BG_RED = "\033[41m"
    BG_GREEN = "\033[42m"
    BG_YELLOW = "\033[43m"
    BG_BLUE = "\033[44m"

    @classmethod
    def disable(cls):
        """Disable all colors (for non-TTY output)."""
        for attr in dir(cls):
            if not attr.startswith('_') and attr.isupper():
                setattr(cls, attr, "")


class CLI:
    """Interactive CLI utilities."""

    def __init__(self, use_colors: bool = True, no_interaction: bool = False):
        """Initialize CLI with optional color support."""
        self.use_colors = use_colors and sys.stdout.isatty()
        self.no_interaction = no_interaction
        if not self.use_colors:
            Colors.disable()

    def print(self, message: str = "", end: str = "\n"):
        """Print a message."""
        print(message, end=end, flush=True)

    def print_header(self, title: str):
        """Print a styled header."""
        width = max(60, len(title) + 4)
        border = "═" * width

        self.print()
        self.print(f"{Colors.CYAN}{Colors.BOLD}╔{border}╗{Colors.RESET}")
        self.print(f"{Colors.CYAN}{Colors.BOLD}║{Colors.RESET} {title.center(width - 2)} {Colors.CYAN}{Colors.BOLD}║{Colors.RESET}")
        self.print(f"{Colors.CYAN}{Colors.BOLD}╚{border}╝{Colors.RESET}")
        self.print()

    def print_section(self, title: str):
        """Print a section header."""
        self.print()
        self.print(f"{Colors.BLUE}{Colors.BOLD}▶ {title}{Colors.RESET}")
        self.print(f"{Colors.DIM}{'─' * 50}{Colors.RESET}")

    def print_success(self, message: str):
        """Print a success message."""
        self.print(f"{Colors.GREEN}✓{Colors.RESET} {message}")

    def print_error(self, message: str):
        """Print an error message."""
        self.print(f"{Colors.RED}✗{Colors.RESET} {message}")

    def print_warning(self, message: str):
        """Print a warning message."""
        self.print(f"{Colors.YELLOW}⚠{Colors.RESET} {message}")

    def print_info(self, message: str):
        """Print an info message."""
        self.print(f"{Colors.CYAN}ℹ{Colors.RESET} {message}")

    def print_step(self, step: int, total: int, message: str):
        """Print a step progress message."""
        self.print(f"{Colors.MAGENTA}[{step}/{total}]{Colors.RESET} {message}")

    def print_progress(self, message: str):
        """Print a progress message (in-place update)."""
        self.print(f"\r{Colors.DIM}  → {message}...{Colors.RESET}", end="")

    def clear_progress(self):
        """Clear the progress line."""
        self.print("\r" + " " * 80 + "\r", end="")

    def ask_yes_no(
        self,
        question: str,
        default: bool = True
    ) -> bool:
        """Ask a yes/no question and return the answer."""
        # In no-interaction mode, always return default
        if self.no_interaction:
            self.print_info(f"{question} → {'oui' if default else 'non'} (auto)")
            return default

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
                    self.print_warning("Répondez par 'o' (oui) ou 'n' (non)")

            except EOFError:
                return default
            except KeyboardInterrupt:
                self.print()
                raise

    def ask_path(
        self,
        question: str,
        default: str,
        validate: Optional[Callable[[str], bool]] = None
    ) -> str:
        """Ask for a path with a default value."""
        # In no-interaction mode, always return default
        if self.no_interaction:
            self.print_info(f"{question} → {default} (auto)")
            return default

        prompt = f"{Colors.YELLOW}?{Colors.RESET} {question}\n  [{Colors.DIM}{default}{Colors.RESET}]: "

        while True:
            try:
                response = input(prompt).strip()

                if not response:
                    return default

                if validate is None or validate(response):
                    return response
                else:
                    self.print_warning("Chemin invalide, réessayez")

            except EOFError:
                return default
            except KeyboardInterrupt:
                self.print()
                raise

    def ask_choice(
        self,
        question: str,
        choices: list[str],
        default: int = 0
    ) -> int:
        """Ask user to choose from a list of options."""
        # In no-interaction mode, always return default
        if self.no_interaction:
            self.print_info(f"{question} → {choices[default]} (auto)")
            return default

        self.print(f"{Colors.YELLOW}?{Colors.RESET} {question}")

        for i, choice in enumerate(choices):
            marker = f"{Colors.GREEN}→{Colors.RESET}" if i == default else " "
            self.print(f"  {marker} [{i + 1}] {choice}")

        prompt = f"  Choix [{default + 1}]: "

        while True:
            try:
                response = input(prompt).strip()

                if not response:
                    return default

                try:
                    idx = int(response) - 1
                    if 0 <= idx < len(choices):
                        return idx
                except ValueError:
                    pass

                self.print_warning(f"Choisissez un nombre entre 1 et {len(choices)}")

            except EOFError:
                return default
            except KeyboardInterrupt:
                self.print()
                raise

    def confirm_action(
        self,
        action: str,
        details: Optional[list[str]] = None
    ) -> bool:
        """Confirm an action with optional details."""
        self.print()
        self.print(f"{Colors.BOLD}Action: {action}{Colors.RESET}")

        if details:
            for detail in details:
                self.print(f"  {Colors.DIM}• {detail}{Colors.RESET}")

        self.print()
        return self.ask_yes_no("Confirmer cette action?")

    def show_summary(self, title: str, items: dict[str, str]):
        """Show a summary of key-value pairs."""
        self.print()
        self.print(f"{Colors.BOLD}{title}{Colors.RESET}")

        max_key_len = max(len(k) for k in items.keys())
        for key, value in items.items():
            self.print(f"  {Colors.CYAN}{key.ljust(max_key_len)}{Colors.RESET} : {value}")

        self.print()
