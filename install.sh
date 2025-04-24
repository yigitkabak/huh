#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${GREEN}       HUH Image Converter          ${NC}"
echo -e "${BLUE}       Installation Script           ${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Function to detect environment
detect_environment() {
    if [ -d "/data/data/com.termux" ]; then
        echo "termux"
    else
        echo "linux"
    fi
}

# Function to detect shell
detect_shell() {
    if [ -n "$ZSH_VERSION" ]; then
        echo "zsh"
    elif [ -n "$BASH_VERSION" ]; then
        echo "bash"
    else
        # Default to bash if we can't determine
        echo "bash"
    fi
}

# Get the environment
ENV=$(detect_environment)
SHELL_TYPE=$(detect_shell)

echo -e "${YELLOW}Detected environment: ${ENV}${NC}"
echo -e "${YELLOW}Detected shell: ${SHELL_TYPE}${NC}"
echo ""

# Build and install the binary
echo -e "${BLUE}Building HUH Image Converter...${NC}"
cargo build --release

if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed. Please check your Rust installation and dependencies.${NC}"
    exit 1
fi

# Install the binary
echo -e "${BLUE}Installing HUH command...${NC}"
cargo install --path .

if [ $? -ne 0 ]; then
    echo -e "${RED}Installation failed.${NC}"
    exit 1
fi

# Function to update PATH in shell config
update_path_in_config() {
    local config_file=$1
    local cargo_bin_path=$2
    
    # Check if config file exists
    if [ ! -f "$config_file" ]; then
        echo -e "${YELLOW}Creating $config_file${NC}"
        touch "$config_file"
    fi
    
    # Check if PATH already includes cargo bin
    if grep -q "export PATH=.*$cargo_bin_path" "$config_file"; then
        echo -e "${GREEN}PATH already configured in $config_file${NC}"
    else
        echo -e "${BLUE}Updating PATH in $config_file${NC}"
        echo "" >> "$config_file"
        echo "# Added by HUH Image Converter installer" >> "$config_file"
        echo "export PATH=\$PATH:$cargo_bin_path" >> "$config_file"
    fi
}

# Configure environment-specific settings
if [ "$ENV" = "termux" ]; then
    CARGO_BIN="/data/data/com.termux/files/home/.cargo/bin"
    
    # Check for ZSH
    if [ -f "$HOME/.zshrc" ]; then
        update_path_in_config "$HOME/.zshrc" "$CARGO_BIN"
    fi
    
    # Check for Bash
    if [ -f "$HOME/.bashrc" ]; then
        update_path_in_config "$HOME/.bashrc" "$CARGO_BIN"
    fi
    
    # Create symlink in Termux bin directory
    echo -e "${BLUE}Creating symlink in Termux bin directory...${NC}"
    ln -sf "$CARGO_BIN/huh" "/data/data/com.termux/files/usr/bin/huh"
else
    # Regular Linux
    CARGO_BIN="$HOME/.cargo/bin"
    
    # Check for ZSH
    if [ -f "$HOME/.zshrc" ]; then
        update_path_in_config "$HOME/.zshrc" "$CARGO_BIN"
    fi
    
    # Check for Bash
    if [ -f "$HOME/.bashrc" ]; then
        update_path_in_config "$HOME/.bashrc" "$CARGO_BIN"
    fi
    
    # Create symlink in /usr/local/bin if we have permission
    if [ -w "/usr/local/bin" ]; then
        echo -e "${BLUE}Creating symlink in /usr/local/bin...${NC}"
        sudo ln -sf "$CARGO_BIN/huh" "/usr/local/bin/huh"
    else
        echo -e "${YELLOW}No write permission to /usr/local/bin. Skipping symlink creation.${NC}"
    fi
fi

echo ""
echo -e "${GREEN}Installation completed!${NC}"
echo ""
echo -e "${YELLOW}To use the HUH command immediately, run:${NC}"
echo -e "    ${BLUE}source ~/.$([ \"$SHELL_TYPE\" = \"zsh\" ] && echo \"zshrc\" || echo \"bashrc\")${NC}"
echo ""
echo -e "${YELLOW}Or restart your terminal.${NC}"
echo ""
echo -e "${GREEN}Try the following commands:${NC}"
echo -e "    ${BLUE}huh help${NC}             - Show help"
echo -e "    ${BLUE}huh convert img.png img.huh${NC} - Convert image to HUH format"
echo -e "    ${BLUE}huh view img.png${NC}     - View an image"
echo ""
echo -e "${BLUE}======================================${NC}"