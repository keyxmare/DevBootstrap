#!/usr/bin/env bash
#
# DevBootstrap - Script de lancement unifié
# Usage: curl -fsSL <url>/install.sh | bash
#    or: ./install.sh
#
# Ce script affiche un menu avec toutes les applications disponibles
# et permet de sélectionner ce que vous souhaitez installer.
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

print_banner() {
    echo ""
    echo -e "${CYAN}${BOLD}╔════════════════════════════════════════════════════════════════╗${RESET}"
    echo -e "${CYAN}${BOLD}║${RESET}              ${BOLD}DevBootstrap${RESET} - Installation automatique              ${CYAN}${BOLD}║${RESET}"
    echo -e "${CYAN}${BOLD}╚════════════════════════════════════════════════════════════════╝${RESET}"
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
    # Check for Python 3.9+
    if command -v python3 &> /dev/null; then
        PYTHON_VERSION=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
        PYTHON_MAJOR=$(echo "$PYTHON_VERSION" | cut -d. -f1)
        PYTHON_MINOR=$(echo "$PYTHON_VERSION" | cut -d. -f2)

        if [ "$PYTHON_MAJOR" -ge 3 ] && [ "$PYTHON_MINOR" -ge 9 ]; then
            echo "python3"
            return 0
        fi
    fi

    echo ""
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
    PYTHON_CMD=$(check_python)

    if [ -z "$PYTHON_CMD" ]; then
        print_warning "Python 3.9+ non trouvé"

        OS_TYPE=$(detect_os)
        case "$OS_TYPE" in
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

        PYTHON_CMD=$(check_python)
        if [ -z "$PYTHON_CMD" ]; then
            print_error "Impossible d'installer Python"
            exit 1
        fi
    fi

    print_success "Python trouvé: $($PYTHON_CMD --version)"
    echo "$PYTHON_CMD"
}

download_installer() {
    print_step "Téléchargement de DevBootstrap..."

    # Create install directory
    mkdir -p "$INSTALL_DIR"

    # Check if git is available
    if command -v git &> /dev/null; then
        if [ -d "$INSTALL_DIR/.git" ]; then
            print_step "Mise à jour de DevBootstrap..."
            cd "$INSTALL_DIR"
            git pull --quiet
        else
            rm -rf "$INSTALL_DIR"
            git clone --depth=1 "$REPO_URL" "$INSTALL_DIR"
        fi
    else
        # Download as zip if git not available
        TEMP_ZIP="/tmp/devbootstrap.zip"
        curl -fsSL "${REPO_URL}/archive/refs/heads/main.zip" -o "$TEMP_ZIP"

        rm -rf "$INSTALL_DIR"
        unzip -q "$TEMP_ZIP" -d /tmp
        mv /tmp/DevBootstrap-main "$INSTALL_DIR"
        rm "$TEMP_ZIP"
    fi

    print_success "DevBootstrap téléchargé"
}

run_installer() {
    PYTHON_CMD="$1"

    print_step "Lancement du menu d'installation..."
    echo ""

    cd "$INSTALL_DIR"

    # Run the unified Python installer (bootstrap module)
    "$PYTHON_CMD" -m bootstrap.app "$@"
}

main() {
    print_banner

    # Detect OS
    OS_TYPE=$(detect_os)
    print_step "Système détecté: $OS_TYPE"

    if [ "$OS_TYPE" = "unsupported" ]; then
        print_error "Système non supporté"
        print_warning "Systèmes supportés: macOS, Ubuntu, Debian"
        exit 1
    fi

    # Ensure Python is available
    PYTHON_CMD=$(ensure_python)

    # Download/update installer
    download_installer

    # Run the installer
    run_installer "$PYTHON_CMD" "${@:2}"
}

# Check if script is being piped
if [ -t 0 ]; then
    main "$@"
else
    # When piped, save stdin and restore it
    exec 3<&0
    main "$@" <&3
fi
