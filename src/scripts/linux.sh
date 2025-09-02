#!/bin/bash

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' 

CMD_NAME="huh"

echo -e "${GREEN}Starting '${CMD_NAME}' command line tool installation for Linux...${NC}"

install_go() {
    echo -e "${YELLOW}Go compiler not found. Attempting to install...${NC}"
    
    if command -v apt &> /dev/null; then
        echo "Detected APT package manager (Debian/Ubuntu)"
        sudo apt update && sudo apt install -y golang-go
    elif command -v yum &> /dev/null; then
        echo "Detected YUM package manager (RHEL/CentOS)"
        sudo yum install -y golang
    elif command -v dnf &> /dev/null; then
        echo "Detected DNF package manager (Fedora)"
        sudo dnf install -y golang
    elif command -v pacman &> /dev/null; then
        echo "Detected Pacman package manager (Arch Linux)"
        sudo pacman -S --noconfirm go
    elif command -v zypper &> /dev/null; then
        echo "Detected Zypper package manager (openSUSE)"
        sudo zypper install -y go
    elif command -v apk &> /dev/null; then
        echo "Detected APK package manager (Alpine Linux)"
        sudo apk add --no-cache go
    elif command -v eopkg &> /dev/null; then
        echo "Detected Eopkg package manager (Solus)"
        sudo eopkg install -y golang
    elif command -v pisi &> /dev/null; then
        echo "Detected PiSi package manager (Pardus)"
        sudo pisi install golang
    elif command -v xbps-install &> /dev/null; then
        echo "Detected XBPS package manager (Void Linux)"
        sudo xbps-install -S go
    elif command -v emerge &> /dev/null; then
        echo "Detected Portage package manager (Gentoo)"
        sudo emerge --ask=n dev-lang/go
    elif command -v nix-env &> /dev/null; then
        echo "Detected Nix package manager (NixOS)"
        nix-env -iA nixpkgs.go
    elif command -v swupd &> /dev/null; then
        echo "Detected swupd package manager (Clear Linux)"
        sudo swupd bundle-add go-basic
    else
        echo -e "${RED}Could not detect package manager.${NC}"
        echo "Please install Go manually from: https://golang.org/dl/"
        echo "Supported package managers: apt, yum, dnf, pacman, zypper, apk, eopkg, pisi, xbps, emerge, nix, swupd"
        exit 1
    fi
}

if ! command -v go &> /dev/null; then
    install_go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Go installation failed. Please install it manually.${NC}"
        exit 1
    fi
fi
echo "Go compiler found."

if [ ! -f "../main.go" ]; then
    echo -e "${RED}Error: main.go not found in the current directory.${NC}"
    echo "Please run this script from the project's root directory."
    exit 1
fi

echo "Ensuring Go module dependencies are up to date..."
go mod tidy
if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to tidy Go modules. Please check for errors.${NC}"
    exit 1
fi

echo "Building the '${CMD_NAME}' binary..."
CGO_ENABLED=0 go build -o ${CMD_NAME} ../main.go

if [ ! -f "${CMD_NAME}" ]; then
    echo -e "${RED}Build failed. Please check for errors above.${NC}"
    exit 1
fi
echo "Build successful."

# Determine installation directory
if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ -w "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "${YELLOW}Adding $INSTALL_DIR to PATH...${NC}"
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
        echo "Please run 'source ~/.bashrc' or restart your terminal to update PATH."
    fi
else
    echo -e "${RED}No writable installation directory found.${NC}"
    echo "Please run with sudo or ensure ~/.local/bin exists and is writable."
    rm -f ${CMD_NAME}
    exit 1
fi

echo "Installing '${CMD_NAME}' to ${INSTALL_DIR}..."

if [ "$INSTALL_DIR" = "/usr/local/bin" ] && [ ! -w "/usr/local/bin" ]; then
    if sudo mv ${CMD_NAME} "${INSTALL_DIR}/${CMD_NAME}"; then
        sudo chmod +x "${INSTALL_DIR}/${CMD_NAME}"
        echo -e "${GREEN}✅ '${CMD_NAME}' was successfully installed system-wide!${NC}"
    else
        echo -e "${RED}Failed to install '${CMD_NAME}' to ${INSTALL_DIR}.${NC}"
        rm -f ${CMD_NAME}
        exit 1
    fi
else
    # User installation
    if mv ${CMD_NAME} "${INSTALL_DIR}/${CMD_NAME}"; then
        chmod +x "${INSTALL_DIR}/${CMD_NAME}"
        echo -e "${GREEN}✅ '${CMD_NAME}' was successfully installed!${NC}"
    else
        echo -e "${RED}Failed to install '${CMD_NAME}' to ${INSTALL_DIR}.${NC}"
        rm -f ${CMD_NAME}
        exit 1
    fi
fi

echo "You can now use the '${CMD_NAME}' command anywhere in your terminal."
echo "Try running: ${CMD_NAME} help"

echo -e "${GREEN}Installation completed successfully in: ${INSTALL_DIR}${NC}"
