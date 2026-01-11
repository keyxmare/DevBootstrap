"""Base installer class for Zsh installation."""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Optional
from enum import Enum
from ..utils.system import SystemInfo
from ..utils.cli import CLI
from ..utils.runner import CommandRunner


@dataclass
class Dependency:
    """Represents a dependency that needs to be installed."""
    name: str
    description: str
    check_command: str
    required: bool = True
    version_arg: str = "--version"


class ZshTheme(Enum):
    """Available Oh My Zsh themes."""
    ROBBYRUSSELL = "robbyrussell"  # Default
    AGNOSTER = "agnoster"
    POWERLEVEL10K = "powerlevel10k/powerlevel10k"
    SPACESHIP = "spaceship"


@dataclass
class InstallOptions:
    """Options for the installation process."""
    install_oh_my_zsh: bool = True
    install_autocompletion: bool = True
    install_syntax_highlighting: bool = True
    install_autosuggestions: bool = True
    set_as_default_shell: bool = True
    theme: str = "robbyrussell"
    backup_existing: bool = True


@dataclass
class InstallResult:
    """Result of an installation."""
    success: bool
    message: str
    zsh_path: Optional[str] = None
    zsh_version: Optional[str] = None
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class BaseInstaller(ABC):
    """Abstract base class for platform-specific installers."""

    # Common dependencies needed for Zsh and plugins
    COMMON_DEPENDENCIES: list[Dependency] = [
        Dependency(
            name="git",
            description="Systeme de controle de version (requis pour oh-my-zsh)",
            check_command="git"
        ),
        Dependency(
            name="curl",
            description="Outil de transfert de donnees",
            check_command="curl"
        ),
    ]

    # Oh My Zsh plugin repositories
    PLUGIN_REPOS = {
        "zsh-autosuggestions": "https://github.com/zsh-users/zsh-autosuggestions.git",
        "zsh-syntax-highlighting": "https://github.com/zsh-users/zsh-syntax-highlighting.git",
        "zsh-completions": "https://github.com/zsh-users/zsh-completions.git",
    }

    def __init__(self, system_info: SystemInfo, cli: CLI, runner: CommandRunner):
        """Initialize the installer."""
        self.system_info = system_info
        self.cli = cli
        self.runner = runner
        self.options: Optional[InstallOptions] = None

    @abstractmethod
    def get_dependencies(self) -> list[Dependency]:
        """Get the list of dependencies for this platform."""
        pass

    @abstractmethod
    def install_package_manager(self) -> bool:
        """Install the package manager if not present."""
        pass

    @abstractmethod
    def update_package_manager(self) -> bool:
        """Update the package manager."""
        pass

    @abstractmethod
    def install_package(self, package: str, description: Optional[str] = None) -> bool:
        """Install a package using the platform's package manager."""
        pass

    @abstractmethod
    def install_zsh(self) -> bool:
        """Install Zsh."""
        pass

    @abstractmethod
    def install_bash_completion(self) -> bool:
        """Install bash-completion package."""
        pass

    def check_dependency(self, dep: Dependency) -> bool:
        """Check if a dependency is installed."""
        return self.runner.check_command_exists(dep.check_command)

    def check_all_dependencies(self) -> dict[str, bool]:
        """Check all dependencies and return their status."""
        results = {}
        for dep in self.get_dependencies():
            results[dep.name] = self.check_dependency(dep)
        return results

    def install_missing_dependencies(self) -> bool:
        """Install all missing dependencies."""
        dependencies = self.get_dependencies()
        total = len(dependencies)
        success = True

        for i, dep in enumerate(dependencies, 1):
            if self.check_dependency(dep):
                self.cli.print_success(f"{dep.name} - deja installe")
                continue

            self.cli.print_step(i, total, f"Installation de {dep.name} ({dep.description})")

            if not self.install_package(dep.name, dep.description):
                if dep.required:
                    self.cli.print_error(f"Echec de l'installation de {dep.name}")
                    success = False
                else:
                    self.cli.print_warning(f"Impossible d'installer {dep.name} (optionnel)")
            else:
                self.cli.print_success(f"{dep.name} installe")

        return success

    def check_existing_installation(self) -> bool:
        """Check if Zsh is already installed."""
        return self.runner.check_command_exists("zsh")

    def verify_installation(self) -> Optional[str]:
        """Verify Zsh installation and return version."""
        version = self.runner.get_command_version("zsh")
        if version:
            return version
        return None

    def backup_existing_config(self) -> bool:
        """Backup existing Zsh configuration."""
        import os
        import datetime

        zshrc_path = self.system_info.get_zshrc_path()
        if not os.path.exists(zshrc_path):
            return True

        timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_path = f"{zshrc_path}.backup.{timestamp}"

        self.cli.print_info(f"Sauvegarde de .zshrc vers {backup_path}")
        return self.runner.copy_file(zshrc_path, backup_path)

    def install_oh_my_zsh(self) -> bool:
        """Install Oh My Zsh."""
        import os

        oh_my_zsh_dir = self.system_info.get_oh_my_zsh_dir()

        # Check if already installed
        if os.path.exists(oh_my_zsh_dir):
            self.cli.print_info("Oh My Zsh est deja installe")
            return True

        self.cli.print_info("Installation de Oh My Zsh...")

        # Clone oh-my-zsh repository
        return self.runner.clone_git_repo(
            url="https://github.com/ohmyzsh/ohmyzsh.git",
            destination=oh_my_zsh_dir,
            depth=1,
            description="Clonage de Oh My Zsh"
        )

    def install_zsh_plugin(self, plugin_name: str) -> bool:
        """Install a Zsh plugin from the plugin repos."""
        import os

        if plugin_name not in self.PLUGIN_REPOS:
            self.cli.print_warning(f"Plugin inconnu: {plugin_name}")
            return False

        custom_dir = self.system_info.get_zsh_custom_dir()
        plugins_dir = os.path.join(custom_dir, "plugins")
        plugin_dir = os.path.join(plugins_dir, plugin_name)

        # Ensure plugins directory exists
        self.runner.ensure_directory(plugins_dir)

        # Check if already installed
        if os.path.exists(plugin_dir):
            self.cli.print_info(f"Plugin {plugin_name} deja installe")
            return True

        # Clone the plugin
        return self.runner.clone_git_repo(
            url=self.PLUGIN_REPOS[plugin_name],
            destination=plugin_dir,
            depth=1,
            description=f"Installation du plugin {plugin_name}"
        )

    def set_default_shell(self) -> bool:
        """Set Zsh as the default shell."""
        zsh_path = self.runner.get_command_path("zsh")
        if not zsh_path:
            self.cli.print_error("Zsh non trouve dans le PATH")
            return False

        self.cli.print_info(f"Changement du shell par defaut vers {zsh_path}")

        # Use chsh to change shell
        result = self.runner.run_interactive(
            ["chsh", "-s", zsh_path],
            description="Changement du shell par defaut"
        )

        return result

    def create_zshrc(self) -> bool:
        """Create or update .zshrc with proper configuration."""
        import os

        zshrc_path = self.system_info.get_zshrc_path()
        oh_my_zsh_dir = self.system_info.get_oh_my_zsh_dir()

        # Determine plugins to enable
        plugins = ["git"]

        if self.options:
            if self.options.install_autosuggestions:
                plugins.append("zsh-autosuggestions")
            if self.options.install_syntax_highlighting:
                plugins.append("zsh-syntax-highlighting")
            if self.options.install_autocompletion:
                plugins.append("zsh-completions")

        plugins_str = " ".join(plugins)
        theme = self.options.theme if self.options else "robbyrussell"

        # Create .zshrc content
        zshrc_content = f'''# Path to your oh-my-zsh installation.
export ZSH="{oh_my_zsh_dir}"

# Theme
ZSH_THEME="{theme}"

# Plugins
plugins=({plugins_str})

# Load Oh My Zsh
source $ZSH/oh-my-zsh.sh

# User configuration

# Aliases
alias ll="ls -la"
alias la="ls -A"
alias l="ls -CF"

# History configuration
HISTSIZE=10000
SAVEHIST=10000
setopt HIST_IGNORE_DUPS
setopt HIST_IGNORE_SPACE
setopt SHARE_HISTORY

# Autocompletion settings
autoload -Uz compinit
compinit

# Case-insensitive completion
zstyle ':completion:*' matcher-list 'm:{{a-z}}={{A-Z}}'

# Colored completion
zstyle ':completion:*' list-colors "${{(s.:.)LS_COLORS}}"

# Menu selection
zstyle ':completion:*' menu select

# zsh-completions (if installed)
if [[ -d "${{ZSH_CUSTOM:-$ZSH/custom}}/plugins/zsh-completions/src" ]]; then
    fpath+="${{ZSH_CUSTOM:-$ZSH/custom}}/plugins/zsh-completions/src"
fi
'''

        # Write .zshrc
        try:
            if self.runner.dry_run:
                self.cli.print_info(f"[DRY RUN] Ecriture de {zshrc_path}")
                return True

            with open(zshrc_path, "w") as f:
                f.write(zshrc_content)

            self.cli.print_success(f".zshrc cree: {zshrc_path}")
            return True

        except Exception as e:
            self.cli.print_error(f"Impossible de creer .zshrc: {e}")
            return False

    def install(self, options: InstallOptions) -> InstallResult:
        """Run the complete installation process."""
        self.options = options
        errors = []
        warnings = []

        # Step 0: Check existing installation
        self.cli.print_section("Verification de l'installation existante")
        if self.check_existing_installation():
            version = self.verify_installation()
            self.cli.print_info(f"Zsh est deja installe: {version}")

            if not self.cli.ask_yes_no("Voulez-vous continuer la configuration?", default=True):
                return InstallResult(
                    success=True,
                    message="Zsh est deja installe",
                    zsh_path=self.runner.get_command_path("zsh"),
                    zsh_version=version
                )

        # Step 1: Check/install package manager
        self.cli.print_section("Verification du gestionnaire de paquets")
        if not self.install_package_manager():
            return InstallResult(
                success=False,
                message="Impossible d'installer le gestionnaire de paquets",
                errors=["Package manager installation failed"]
            )

        # Step 2: Update package manager
        self.cli.print_section("Mise a jour du gestionnaire de paquets")
        if not self.update_package_manager():
            warnings.append("Impossible de mettre a jour le gestionnaire de paquets")
            self.cli.print_warning("Impossible de mettre a jour le gestionnaire de paquets")

        # Step 3: Install dependencies
        self.cli.print_section("Installation des dependances")
        if not self.install_missing_dependencies():
            warnings.append("Certaines dependances n'ont pas pu etre installees")

        # Step 4: Install Zsh
        self.cli.print_section("Installation de Zsh")
        if not self.check_existing_installation():
            if not self.install_zsh():
                return InstallResult(
                    success=False,
                    message="Echec de l'installation de Zsh",
                    errors=["Zsh installation failed"]
                )

        # Step 5: Backup existing config
        if options.backup_existing:
            self.cli.print_section("Sauvegarde de la configuration existante")
            if not self.backup_existing_config():
                warnings.append("Impossible de sauvegarder la configuration existante")

        # Step 6: Install Oh My Zsh
        if options.install_oh_my_zsh:
            self.cli.print_section("Installation de Oh My Zsh")
            if not self.install_oh_my_zsh():
                warnings.append("Oh My Zsh non installe")
                self.cli.print_warning("Oh My Zsh non installe")

        # Step 7: Install plugins
        if options.install_oh_my_zsh:
            self.cli.print_section("Installation des plugins")

            if options.install_autosuggestions:
                if not self.install_zsh_plugin("zsh-autosuggestions"):
                    warnings.append("Plugin zsh-autosuggestions non installe")
                else:
                    self.cli.print_success("zsh-autosuggestions installe")

            if options.install_syntax_highlighting:
                if not self.install_zsh_plugin("zsh-syntax-highlighting"):
                    warnings.append("Plugin zsh-syntax-highlighting non installe")
                else:
                    self.cli.print_success("zsh-syntax-highlighting installe")

            if options.install_autocompletion:
                if not self.install_zsh_plugin("zsh-completions"):
                    warnings.append("Plugin zsh-completions non installe")
                else:
                    self.cli.print_success("zsh-completions installe")

        # Step 8: Install bash-completion (for bash users)
        self.cli.print_section("Installation de bash-completion")
        if not self.install_bash_completion():
            warnings.append("bash-completion non installe")
            self.cli.print_warning("bash-completion non installe (optionnel)")
        else:
            self.cli.print_success("bash-completion installe")

        # Step 9: Create .zshrc
        if options.install_oh_my_zsh:
            self.cli.print_section("Configuration de .zshrc")
            if not self.create_zshrc():
                warnings.append("Configuration .zshrc non creee")

        # Step 10: Set default shell
        if options.set_as_default_shell:
            self.cli.print_section("Configuration du shell par defaut")
            if not self.set_default_shell():
                warnings.append("Zsh n'a pas ete defini comme shell par defaut")
                self.cli.print_warning("Zsh n'a pas ete defini comme shell par defaut")

        # Step 11: Verify installation
        self.cli.print_section("Verification de l'installation")
        version = self.verify_installation()
        if not version:
            return InstallResult(
                success=False,
                message="Zsh installe mais non trouve dans le PATH",
                errors=["Zsh not found in PATH after installation"]
            )

        zsh_path = self.runner.get_command_path("zsh")

        return InstallResult(
            success=True,
            message="Installation terminee avec succes!",
            zsh_path=zsh_path,
            zsh_version=version,
            warnings=warnings
        )
