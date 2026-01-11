#!/usr/bin/env bash
#
# DevBootstrap - Script d'installation
# Usage: curl -fsSL <url>/install.sh | bash
#    or: ./install.sh
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# Configuration
REPO_URL="https://github.com/keyxmare/DevBootstrap"
CACHE_DIR="${HOME}/.devbootstrap"
BIN_DIR="/usr/local/bin"
BINARY_NAME="devbootstrap"

# Global variables
BOOTSTRAP_ARGS=()

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                BOOTSTRAP_ARGS+=("$1")
                shift
                ;;
        esac
    done
}

show_help() {
    echo ""
    echo -e "${BOLD}DevBootstrap${RESET} - Installation automatique de l'environnement de developpement"
    echo ""
    echo -e "${BOLD}Usage:${RESET}"
    echo "  curl -fsSL https://raw.githubusercontent.com/keyxmare/DevBootstrap/main/install.sh | bash"
    echo "  ./install.sh [options]"
    echo ""
    echo -e "${BOLD}Options:${RESET}"
    echo "  -h, --help            Affiche ce message d'aide"
    echo ""
    echo -e "${BOLD}Apres installation:${RESET}"
    echo "  devbootstrap                      # Mode interactif (par defaut)"
    echo "  devbootstrap --no-interaction     # Installe tout automatiquement"
    echo "  devbootstrap --uninstall          # Mode desinstallation"
    echo "  devbootstrap --help               # Plus d'options"
    echo ""
}

print_banner() {
    echo -e "${CYAN}${BOLD}+==============================================================+${RESET}"
    echo -e "${CYAN}${BOLD}|${RESET}           DevBootstrap - Installation                        ${CYAN}${BOLD}|${RESET}"
    echo -e "${CYAN}${BOLD}+==============================================================+${RESET}"
    echo ""
}

print_step() {
    echo -e "${BLUE}>${RESET} $1"
}

print_success() {
    echo -e "${GREEN}[OK]${RESET} $1"
}

print_error() {
    echo -e "${RED}[X]${RESET} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${RESET} $1"
}

detect_os() {
    case "$(uname -s)" in
        Darwin*)
            echo "macos"
            ;;
        Linux*)
            if [ -f /etc/os-release ]; then
                . /etc/os-release
                case "$ID" in
                    ubuntu|debian|linuxmint|pop)
                        echo "ubuntu"
                        ;;
                    *)
                        echo "linux_other"
                        ;;
                esac
            else
                echo "linux_other"
            fi
            ;;
        *)
            echo "unsupported"
            ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        arm64|aarch64)
            echo "arm64"
            ;;
        x86_64|amd64)
            echo "amd64"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

install_binary() {
    local source_path="$1"
    local dest_path="$BIN_DIR/$BINARY_NAME"

    chmod +x "$source_path"

    # Try without sudo first, then with sudo if needed
    if cp -f "$source_path" "$dest_path" 2>/dev/null; then
        return 0
    elif sudo cp -f "$source_path" "$dest_path" 2>/dev/null; then
        sudo chmod +x "$dest_path"
        return 0
    else
        print_error "Impossible de copier le binaire dans $BIN_DIR"
        print_warning "Essayez: sudo cp $source_path $dest_path"
        return 1
    fi
}

download_or_build() {
    local is_update=false

    # Check if already installed
    if [ -f "$BIN_DIR/$BINARY_NAME" ]; then
        is_update=true
        print_step "Mise a jour de DevBootstrap..."
    else
        print_step "Installation de DevBootstrap..."
    fi

    # Create cache directory
    mkdir -p "$CACHE_DIR"

    # Ensure /usr/local/bin exists
    if [ ! -d "$BIN_DIR" ]; then
        sudo mkdir -p "$BIN_DIR" 2>/dev/null || mkdir -p "$BIN_DIR"
    fi

    # Sync repository (always fetch latest)
    if command -v git &> /dev/null; then
        if [ -d "$CACHE_DIR/.git" ]; then
            print_step "Synchronisation avec GitHub..."
            (cd "$CACHE_DIR" && git fetch origin main --quiet 2>/dev/null && git reset --hard origin/main --quiet 2>/dev/null) || {
                print_warning "Echec de la synchronisation, utilisation du cache local"
            }
        else
            rm -rf "$CACHE_DIR" 2>/dev/null || true
            git clone --depth=1 --quiet "$REPO_URL" "$CACHE_DIR" 2>/dev/null || {
                git clone --depth=1 "$REPO_URL" "$CACHE_DIR"
            }
        fi
    else
        print_error "git est requis pour installer DevBootstrap"
        exit 1
    fi

    # Check if Go binary exists in repo
    if [ -f "$CACHE_DIR/$BINARY_NAME" ]; then
        if install_binary "$CACHE_DIR/$BINARY_NAME"; then
            if [ "$is_update" = true ]; then
                print_success "DevBootstrap mis a jour"
            else
                print_success "DevBootstrap installe"
            fi
            return 0
        fi
        return 1
    fi

    # Build from source if Go is available
    if command -v go &> /dev/null && [ -f "$CACHE_DIR/go.mod" ]; then
        print_step "Compilation de DevBootstrap..."
        (cd "$CACHE_DIR" && go build -o "$BINARY_NAME" .) && {
            if install_binary "$CACHE_DIR/$BINARY_NAME"; then
                if [ "$is_update" = true ]; then
                    print_success "DevBootstrap compile et mis a jour"
                else
                    print_success "DevBootstrap compile et installe"
                fi
                return 0
            fi
            return 1
        }
    fi

    print_error "Impossible d'installer DevBootstrap (binaire non disponible et Go non installe)"
    exit 1
}

configure_path() {
    # /usr/local/bin is typically already in PATH on macOS and most Linux systems
    # Only configure if needed
    if echo "$PATH" | grep -q "/usr/local/bin"; then
        return 0
    fi

    print_step "Configuration du PATH..."

    local path_line='export PATH="/usr/local/bin:$PATH"'
    local configured=false

    # Add to .zshrc if using zsh
    if [ -n "$ZSH_VERSION" ] || [ "$SHELL" = "/bin/zsh" ] || [ -f "$HOME/.zshrc" ]; then
        if [ -f "$HOME/.zshrc" ]; then
            if ! grep -q "/usr/local/bin" "$HOME/.zshrc" 2>/dev/null; then
                echo "" >> "$HOME/.zshrc"
                echo "# Added by DevBootstrap" >> "$HOME/.zshrc"
                echo "$path_line" >> "$HOME/.zshrc"
                configured=true
            fi
        fi
    fi

    # Add to .bashrc or .bash_profile
    if [ -f "$HOME/.bashrc" ]; then
        if ! grep -q "/usr/local/bin" "$HOME/.bashrc" 2>/dev/null; then
            echo "" >> "$HOME/.bashrc"
            echo "# Added by DevBootstrap" >> "$HOME/.bashrc"
            echo "$path_line" >> "$HOME/.bashrc"
            configured=true
        fi
    elif [ -f "$HOME/.bash_profile" ]; then
        if ! grep -q "/usr/local/bin" "$HOME/.bash_profile" 2>/dev/null; then
            echo "" >> "$HOME/.bash_profile"
            echo "# Added by DevBootstrap" >> "$HOME/.bash_profile"
            echo "$path_line" >> "$HOME/.bash_profile"
            configured=true
        fi
    fi

    if [ "$configured" = true ]; then
        print_success "PATH configure"
    fi
}

show_next_steps() {
    local is_update=$1
    echo ""
    if [ "$is_update" = true ]; then
        echo -e "${CYAN}${BOLD}Mise a jour terminee!${RESET}"
    else
        echo -e "${CYAN}${BOLD}Installation terminee!${RESET}"
    fi
    echo ""
    echo -e "Lancez DevBootstrap:"
    echo -e "     ${GREEN}devbootstrap${RESET}"
    echo ""
}

main() {
    parse_args "$@"

    echo ""
    print_banner

    # Check for root on macOS
    if [[ "$OSTYPE" == "darwin"* ]] && [[ $EUID -eq 0 ]]; then
        print_error "Ne pas executer ce script avec sudo sur macOS."
        print_error "Homebrew ne fonctionne pas correctement en root."
        echo ""
        exit 1
    fi

    # Detect OS and architecture
    local os_type arch is_update
    os_type=$(detect_os)
    arch=$(detect_arch)
    print_step "Systeme detecte: $os_type ($arch)"

    if [ "$os_type" = "unsupported" ]; then
        print_error "Systeme non supporte"
        print_warning "Systemes supportes: macOS, Ubuntu, Debian"
        exit 1
    fi

    # Check if this is an update
    is_update=false
    if [ -f "$BIN_DIR/$BINARY_NAME" ]; then
        is_update=true
    fi

    # Download/build and install binary
    download_or_build

    # Configure PATH (only on fresh install)
    if [ "$is_update" = false ]; then
        configure_path
    fi

    # Show next steps
    show_next_steps "$is_update"

    # Run devbootstrap if arguments were passed
    if [ ${#BOOTSTRAP_ARGS[@]} -gt 0 ]; then
        echo ""
        print_step "Lancement de DevBootstrap..."
        echo ""
        "$BIN_DIR/$BINARY_NAME" "${BOOTSTRAP_ARGS[@]}"
    fi
}

# Check if script is being piped
if [ -t 0 ]; then
    main "$@"
else
    exec 3<&0
    main "$@" <&3
fi
