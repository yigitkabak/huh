#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' 

CMD_NAME="huh"

echo -e "${GREEN}Starting '${CMD_NAME}' command line tool installation for Termux...${NC}"

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

INSTALL_DIR="$PREFIX/bin"
echo "Installing '${CMD_NAME}' to ${INSTALL_DIR}..."

if mv ${CMD_NAME} "${INSTALL_DIR}/${CMD_NAME}"; then
    echo -e "${GREEN}✅ '${CMD_NAME}' was successfully installed!${NC}"
    echo "You can now use the '${CMD_NAME}' command anywhere in your Termux session."
    echo "Try running: ${CMD_NAME} help"
else
    echo -e "${RED}Failed to move '${CMD_NAME}' to ${INSTALL_DIR}. Please check permissions.${NC}"
    rm -f ${CMD_NAME}
    exit 1
fi
