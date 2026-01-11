#!/usr/bin/env bash
#
# DevBootstrap - Script de lancement unifié
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

# Global variables
PYTHON_CMD=""
NO_INTERACTION=false

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
                print_error "Option inconnue: $1"
                show_help
                exit 1
                ;;
            *)
                # Pass remaining arguments to bootstrap
                break
                ;;
        esac
    done
    # Store remaining args for bootstrap
    BOOTSTRAP_ARGS=("$@")
}

show_help() {
    echo ""
    echo -e "${BOLD}DevBootstrap${RESET} - Installation automatique de l'environnement de développement"
    echo ""
    echo -e "${BOLD}Usage:${RESET}"
    echo "  curl -fsSL https://raw.githubusercontent.com/keyxmare/DevBootstrap/main/install.sh | bash"
    echo "  ./install.sh [options]"
    echo ""
    echo -e "${BOLD}Options:${RESET}"
    echo "  -h, --help            Affiche ce message d'aide"
    echo "  -n, --no-interaction  Mode non-interactif (installe tout sans confirmation)"
    echo ""
    echo -e "${BOLD}Exemples:${RESET}"
    echo "  ./install.sh                      # Mode interactif (par défaut)"
    echo "  ./install.sh --no-interaction     # Installe tout automatiquement"
    echo "  ./install.sh -n                   # Raccourci pour --no-interaction"
    echo ""
}

print_banner() {
    echo -e "${CYAN}${BOLD}╔══════════════════════════════════════════════════════════════╗${RESET}"
    echo -e "${CYAN}${BOLD}║${RESET}           DevBootstrap - Installation automatique           ${CYAN}${BOLD}║${RESET}"
    echo -e "${CYAN}${BOLD}╚══════════════════════════════════════════════════════════════╝${RESET}"
    echo ""
}

print_step() {
    echo -e "${BLUE}▶${RESET} $1"
}

print_success() {
    echo -e "${GREEN}✓${RESET} $1"
}

print_error() {
    echo -e "${RED}✗${RESET} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${RESET} $1"
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

check_python() {
    # Check for Python 3.9+ and set PYTHON_CMD global variable
    if command -v python3 &> /dev/null; then
        local version
        version=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
        local major minor
        major=$(echo "$version" | cut -d. -f1)
        minor=$(echo "$version" | cut -d. -f2)

        if [ "$major" -ge 3 ] && [ "$minor" -ge 9 ]; then
            PYTHON_CMD="python3"
            return 0
        fi
    fi
    return 1
}

install_python_macos() {
    print_step "Installation de Python via Homebrew..."

    # Check if Homebrew is installed
    if ! command -v brew &> /dev/null; then
        print_step "Installation de Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

        # Add Homebrew to PATH for Apple Silicon
        if [ -f "/opt/homebrew/bin/brew" ]; then
            eval "$(/opt/homebrew/bin/brew shellenv)"
        fi
    fi

    brew install python@3.12
}

install_python_ubuntu() {
    print_step "Installation de Python..."
    sudo apt-get update
    sudo apt-get install -y python3 python3-pip python3-venv
}

ensure_python() {
    if check_python; then
        print_success "Python trouvé: $($PYTHON_CMD --version 2>&1)"
        return 0
    fi

    print_warning "Python 3.9+ non trouvé"

    local os_type
    os_type=$(detect_os)
    case "$os_type" in
        macos)
            install_python_macos
            ;;
        ubuntu)
            install_python_ubuntu
            ;;
        *)
            print_error "Veuillez installer Python 3.9+ manuellement"
            exit 1
            ;;
    esac

    if ! check_python; then
        print_error "Impossible d'installer Python"
        exit 1
    fi

    print_success "Python installé: $($PYTHON_CMD --version 2>&1)"
}

download_installer() {
    # Sync silently - don't show git output
    if command -v git &> /dev/null; then
        if [ -d "$INSTALL_DIR/.git" ]; then
            # Update existing repo silently
            (cd "$INSTALL_DIR" && git fetch origin main --quiet 2>/dev/null && git reset --hard origin/main --quiet 2>/dev/null) || true
        else
            # Clone silently
            rm -rf "$INSTALL_DIR" 2>/dev/null || true
            git clone --depth=1 --quiet "$REPO_URL" "$INSTALL_DIR" 2>/dev/null || true
        fi
    fi

    # Check if we have a valid installation, if not try again with output
    if [ ! -d "$INSTALL_DIR" ] || [ ! -f "$INSTALL_DIR/bootstrap/__main__.py" ]; then
        print_step "Téléchargement de DevBootstrap..."

        if command -v git &> /dev/null; then
            rm -rf "$INSTALL_DIR" 2>/dev/null || true
            git clone --depth=1 --quiet "$REPO_URL" "$INSTALL_DIR"
        else
            # Download as zip if git not available
            local temp_zip="/tmp/devbootstrap.zip"
            curl -fsSL "${REPO_URL}/archive/refs/heads/main.zip" -o "$temp_zip"
            rm -rf "$INSTALL_DIR"
            unzip -q "$temp_zip" -d /tmp
            mv /tmp/DevBootstrap-main "$INSTALL_DIR"
            rm "$temp_zip"
        fi

        print_success "DevBootstrap téléchargé"
    fi
}

run_installer() {
    print_step "Lancement du menu d'installation..."
    echo ""

    cd "$INSTALL_DIR"

    # Build arguments for bootstrap
    local args=()
    if [ "$NO_INTERACTION" = true ]; then
        args+=("--no-interaction")
    fi
    args+=("${BOOTSTRAP_ARGS[@]}")

    # Run the unified Python installer (bootstrap module)
    "$PYTHON_CMD" -m bootstrap "${args[@]}"
}

main() {
    # Parse arguments first
    parse_args "$@"

    echo ""
    print_banner

    # Detect OS
    local os_type
    os_type=$(detect_os)
    print_step "Système détecté: $os_type"

    if [ "$os_type" = "unsupported" ]; then
        print_error "Système non supporté"
        print_warning "Systèmes supportés: macOS, Ubuntu, Debian"
        exit 1
    fi

    # Ensure Python is available (sets PYTHON_CMD global)
    ensure_python

    # Download/update installer (silent sync)
    download_installer

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
