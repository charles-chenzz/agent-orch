#!/bin/bash

# Agent Orchestrator Development Environment Setup Script
# For users without Nix

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RESET='\033[0m'

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════╗${RESET}"
echo -e "${CYAN}║   Agent Orchestrator Environment Setup   ║${RESET}"
echo -e "${CYAN}╚══════════════════════════════════════════╝${RESET}"
echo ""

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Darwin) OS="macos" ;;
    Linux)  OS="linux" ;;
    *)      echo -e "${RED}Unsupported OS: $OS${RESET}"; exit 1 ;;
esac
echo -e "${GREEN}Detected OS: $OS${RESET}"

# Check if running in WSL
if [[ "$OS" == "linux" ]] && grep -qi microsoft /proc/version 2>/dev/null; then
    IS_WSL=true
    echo -e "${YELLOW}Running in WSL${RESET}"
else
    IS_WSL=false
fi

# =============================================================================
# Helper Functions
# =============================================================================

check_command() {
    if command -v "$1" >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

install_go() {
    echo -e "${CYAN}Installing Go...${RESET}"
    
    GO_VERSION="1.22.0"
    
    if [[ "$OS" == "macos" ]]; then
        if check_command brew; then
            brew install go
        else
            echo -e "${YELLOW}Homebrew not found. Please install Go manually from https://go.dev/dl/${RESET}"
            exit 1
        fi
    else
        # Linux
        GO_ARCH=$(uname -m)
        case "$GO_ARCH" in
            x86_64) GO_ARCH="amd64" ;;
            aarch64) GO_ARCH="arm64" ;;
        esac
        
        GO_URL="https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
        wget -q "$GO_URL" -O /tmp/go.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf /tmp/go.tar.gz
        rm /tmp/go.tar.gz
        
        # Add to PATH if not already
        if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        fi
        export PATH=$PATH:/usr/local/go/bin
    fi
    
    echo -e "${GREEN}Go installed: $(go version)${RESET}"
}

install_node() {
    echo -e "${CYAN}Installing Node.js via nvm...${RESET}"
    
    # Install nvm if not present
    if [[ ! -d "$HOME/.nvm" ]]; then
        curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
    fi
    
    # Load nvm
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    
    # Install Node.js 20
    nvm install 20
    nvm use 20
    
    echo -e "${GREEN}Node.js installed: $(node --version)${RESET}"
}

install_wails() {
    echo -e "${CYAN}Installing Wails CLI...${RESET}"
    
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    
    # Add Go bin to PATH if not already
    if ! grep -q '$HOME/go/bin' ~/.bashrc; then
        echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
    fi
    export PATH=$PATH:$HOME/go/bin
    
    echo -e "${GREEN}Wails installed: $(wails version)${RESET}"
}

install_linux_deps() {
    echo -e "${CYAN}Installing Linux GUI dependencies...${RESET}"
    
    if check_command apt-get; then
        sudo apt-get update
        sudo apt-get install -y \
            libgtk-3-dev \
            libwebkit2gtk-4.0-dev \
            build-essential \
            pkg-config
    elif check_command dnf; then
        sudo dnf install -y \
            gtk3-devel \
            webkit2gtk4.0-devel \
            gcc \
            pkg-config
    elif check_command pacman; then
        sudo pacman -S --noconfirm \
            gtk3 \
            webkit2gtk \
            base-devel \
            pkg-config
    else
        echo -e "${YELLOW}Package manager not detected. Please install manually:${RESET}"
        echo "  - GTK 3 development libraries"
        echo "  - WebKitGTK development libraries"
        echo "  - Build tools (gcc, make, pkg-config)"
    fi
    
    echo -e "${GREEN}Linux dependencies installed${RESET}"
}

# =============================================================================
# Main Installation
# =============================================================================

# Check and install Go
if check_command go; then
    GO_VERSION=$(go version | grep -oP 'go\d+\.\d+' | head -1)
    echo -e "${GREEN}Go found: $(go version)${RESET}"
else
    install_go
fi

# Check and install Node.js
if check_command node; then
    echo -e "${GREEN}Node.js found: $(node --version)${RESET}"
else
    install_node
fi

# Check and install Wails
if check_command wails; then
    echo -e "${GREEN}Wails found: $(wails version 2>&1 | head -1)${RESET}"
else
    install_wails
fi

# Install Linux-specific dependencies
if [[ "$OS" == "linux" ]]; then
    install_linux_deps
fi

# =============================================================================
# Project Dependencies
# =============================================================================

echo ""
echo -e "${CYAN}Installing project dependencies...${RESET}"

# Go modules
go mod download

# Frontend dependencies
cd frontend
if [[ -f "package.json" ]]; then
    npm install
fi
cd ..

# =============================================================================
# Done
# =============================================================================

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════╗${RESET}"
echo -e "${GREEN}║       Setup Complete!                    ║${RESET}"
echo -e "${GREEN}╚══════════════════════════════════════════╝${RESET}"
echo ""
echo -e "To start developing:"
echo -e "  ${CYAN}source ~/.bashrc${RESET}  # Reload shell config"
echo -e "  ${CYAN}make dev${RESET}          # Start development server"
echo ""
echo -e "Or use Nix for a better experience:"
echo -e "  ${CYAN}nix develop${RESET}       # Enter Nix dev shell"
echo ""
