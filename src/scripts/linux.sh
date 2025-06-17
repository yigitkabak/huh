#!/bin/bash

# Renkler
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting 'huh' command line tool installation...${NC}"

# Go'nun yüklü olup olmadığını kontrol et
if ! command -v go &> /dev/null
then
    echo -e "${RED}Go compiler could not be found.${NC}"
    echo "Please install Go first. For Debian/Ubuntu, you can use: sudo apt update && sudo apt install golang"
    exit 1
fi

echo "Go compiler found."

# Kodu derle
echo "Building the 'huh' binary..."
go build -o huh ../main.go

# Derlemenin başarılı olup olmadığını kontrol et
if [ ! -f "huh" ]; then
    echo -e "${RED}Build failed. Please check for errors above.${NC}"
    exit 1
fi

echo "Build successful."

# Binary'yi /usr/local/bin dizinine taşı
INSTALL_DIR="/usr/local/bin"
echo "Attempting to install 'huh' to ${INSTALL_DIR}..."

if mv huh ${INSTALL_DIR}/huh; then
    echo -e "${GREEN}✅ 'huh' was successfully installed!${NC}"
    echo "You can now use the 'huh' command anywhere in your terminal."
    echo "Try running: huh help"
else
    echo -e "${RED}Failed to move 'huh' to ${INSTALL_DIR}.${NC}"
    echo "Please try running this script with sudo:"
    echo "sudo bash install.sh"
    exit 1
fi
