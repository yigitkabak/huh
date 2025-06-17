#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Starting 'huh' command line tool installation for Termux...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}Go compiler could not be found.${NC}"
    echo "Attempting to install Go via pkg..."
    pkg update && pkg install golang -y
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Go installation failed. Please install it manually.${NC}"
        exit 1
    fi
fi

echo "Go compiler found."

echo "Building the 'huh' binary..."
go build -o huh ../main.go

if [ ! -f "huh" ]; then
    echo -e "${RED}Build failed. Please check for errors above.${NC}"
    exit 1
fi

echo "Build successful."

INSTALL_DIR="$PREFIX/bin"
echo "Installing 'huh' to ${INSTALL_DIR}..."

if mv huh "${INSTALL_DIR}/huh"; then
    echo -e "${GREEN}✅ 'huh' was successfully installed!${NC}"
    echo "You can now use the 'huh' command anywhere in your Termux session."
    echo "Try running: huh help"
else
    echo -e "${RED}Failed to move 'huh' to ${INSTALL_DIR}. Please check permissions.${NC}"
    exit 1
fi
