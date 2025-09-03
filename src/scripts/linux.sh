#!/bin/bash

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

CMD_NAME="huh"

echo -e "${GREEN}Starting '${CMD_NAME}' command line tool installation for Linux...${NC}"

# Check if Go compiler is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go compiler could not be found.${NC}"
    echo "Please install Go manually before continuing."
    echo "For installation instructions: https://golang.org/doc/install"
    exit 1
fi
echo "Go compiler found."

# Check for the main.go file
if [ ! -f "../main.go" ]; then
    echo -e "${RED}Error: main.go not found in the parent directory.${NC}"
    echo "Please run this script from the project's 'install' directory."
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

# Move the binary to the system-wide bin directory
INSTALL_DIR="/usr/local/bin"

echo "Installing '${CMD_NAME}' to ${INSTALL_DIR}..."

# Moving and granting execute permissions with sudo
if sudo mv ${CMD_NAME} "${INSTALL_DIR}/${CMD_NAME}"; then
    sudo chmod +x "${INSTALL_DIR}/${CMD_NAME}"
    echo -e "${GREEN}✅ '${CMD_NAME}' was successfully installed system-wide!${NC}"
else
    echo -e "${RED}Failed to move '${CMD_NAME}' to ${INSTALL_DIR}. Please check permissions.${NC}"
    rm -f ${CMD_NAME}
    exit 1
fi

echo "You can now use the '${CMD_NAME}' command anywhere in your terminal."
echo "Try running: ${CMD_NAME} help"

echo -e "${GREEN}Installation completed successfully in: ${INSTALL_DIR}${NC}"
