#!/usr/bin/env bash
#
# DevBootstrap - Script de lancement unifie
# Usage: curl -fsSL <url>/install.sh | bash
#    or: ./install.sh
#    or: ./install.sh --no-interaction
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
INSTALL_DIR="${HOME}/.devbootstrap"
BINARY_NAME="devbootstrap"

# Global variables
NO_INTERACTION=false
BOOTSTRAP_ARGS=()

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                show_help
                exit 0
                ;;
            -n|--no-interaction)
                NO_INTERACTION=true
                shift
                ;;
            -*)
                # Pass remaining arguments to devbootstrap
                BOOTSTRAP_ARGS+=("$1")
                shift
                ;;
            *)
                # Pass remaining arguments to devbootstrap
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
    echo "  -n, --no-interaction  Mode non-interactif (installe tout sans confirmation)"
    echo "  -u, --uninstall       Mode desinstallation"
    echo "  --dry-run             Simuler sans effectuer de changements"
    echo ""
    echo -e "${BOLD}Exemples:${RESET}"
    echo "  ./install.sh                      # Mode interactif (par defaut)"
    echo "  ./install.sh --no-interaction     # Installe tout automatiquement"
    echo "  ./install.sh -n                   # Raccourci pour --no-interaction"
    echo ""
}

print_banner() {
    echo -e "${CYAN}${BOLD}+==============================================================+${RESET}"
    echo -e "${CYAN}${BOLD}|${RESET}           DevBootstrap - Installation automatique           ${CYAN}${BOLD}|${RESET}"
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

check_go_binary() {
    # Check if we have a pre-built Go binary
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        return 0
    fi
    return 1
}

download_or_build() {
    print_step "Synchronisation de DevBootstrap..."

    # Sync repository
    if command -v git &> /dev/null; then
        if [ -d "$INSTALL_DIR/.git" ]; then
            # Update existing repo silently
            (cd "$INSTALL_DIR" && git fetch origin main --quiet 2>/dev/null && git reset --hard origin/main --quiet 2>/dev/null) || true
        else
            # Clone repository
            rm -rf "$INSTALL_DIR" 2>/dev/null || true
            git clone --depth=1 --quiet "$REPO_URL" "$INSTALL_DIR" 2>/dev/null || {
                print_step "Telechargement de DevBootstrap..."
                git clone --depth=1 "$REPO_URL" "$INSTALL_DIR"
            }
        fi
    else
        print_error "git est requis pour installer DevBootstrap"
        exit 1
    fi

    # Check if Go binary exists
    if check_go_binary; then
        print_success "DevBootstrap pret (binaire Go)"
        return 0
    fi

    # Check if we can build from source
    if command -v go &> /dev/null && [ -f "$INSTALL_DIR/go.mod" ]; then
        print_step "Compilation de DevBootstrap..."
        (cd "$INSTALL_DIR" && go build -o "$BINARY_NAME" .) && {
            print_success "DevBootstrap compile"
            return 0
        }
    fi

    # Fallback to Python if available
    if command -v python3 &> /dev/null && [ -f "$INSTALL_DIR/bootstrap/__main__.py" ]; then
        print_warning "Binaire Go non disponible, utilisation de la version Python"
        return 1
    fi

    print_error "Impossible d'installer DevBootstrap (Go ou Python requis)"
    exit 1
}

run_installer() {
    cd "$INSTALL_DIR"

    # Build arguments
    local args=()
    if [ "$NO_INTERACTION" = true ]; then
        args+=("--no-interaction")
    fi
    args+=("${BOOTSTRAP_ARGS[@]}")

    # Prefer Go binary
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        print_step "Lancement de DevBootstrap..."
        echo ""
        "./$BINARY_NAME" "${args[@]}"
        return $?
    fi

    # Fallback to Python
    if command -v python3 &> /dev/null && [ -f "$INSTALL_DIR/bootstrap/__main__.py" ]; then
        print_step "Lancement de DevBootstrap (Python)..."
        echo ""
        python3 -m bootstrap "${args[@]}"
        return $?
    fi

    print_error "Aucune version executable trouvee"
    exit 1
}

main() {
    # Parse arguments first
    parse_args "$@"

    echo ""
    print_banner

    # Check for root on macOS
    if [[ "$OSTYPE" == "darwin"* ]] && [[ $EUID -eq 0 ]]; then
        print_error "Ne pas executer ce script avec sudo sur macOS."
        print_error "Homebrew ne fonctionne pas correctement en root."
        echo ""
        print_step "Utilisez simplement: ./install.sh"
        exit 1
    fi

    # Detect OS and architecture
    local os_type arch
    os_type=$(detect_os)
    arch=$(detect_arch)
    print_step "Systeme detecte: $os_type ($arch)"

    if [ "$os_type" = "unsupported" ]; then
        print_error "Systeme non supporte"
        print_warning "Systemes supportes: macOS, Ubuntu, Debian"
        exit 1
    fi

    # Download/update and build if needed
    download_or_build

    # Run the installer
    run_installer
}

# Check if script is being piped
if [ -t 0 ]; then
    main "$@"
else
    # When piped, save stdin and restore it
    exec 3<&0
    main "$@" <&3
fi
