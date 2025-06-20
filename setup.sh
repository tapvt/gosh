#!/bin/bash

# Gosh Setup Script
# This script sets up gosh shell with default configurations and integrations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GOSH_CONFIG_DIR="$HOME/.config/gosh"
GOSH_BIN_PATH="/usr/local/bin/gosh"

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.19 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $GO_VERSION found"
    
    if ! command_exists git; then
        print_warning "Git is not installed. Git integration features will be limited."
    else
        print_success "Git found"
    fi
}

# Build gosh
build_gosh() {
    print_status "Building gosh..."
    
    if [ ! -f "cmd/main.go" ]; then
        print_error "cmd/main.go not found. Are you in the gosh project directory?"
        exit 1
    fi
    
    go build -o gosh cmd/main.go
    print_success "Gosh built successfully"
}

# Install gosh binary
install_binary() {
    print_status "Installing gosh binary..."
    
    if [ ! -f "gosh" ]; then
        print_error "gosh binary not found. Build failed?"
        exit 1
    fi
    
    # Check if we can write to /usr/local/bin
    if [ -w "/usr/local/bin" ]; then
        cp gosh "$GOSH_BIN_PATH"
        chmod +x "$GOSH_BIN_PATH"
    else
        print_status "Installing to /usr/local/bin requires sudo..."
        sudo cp gosh "$GOSH_BIN_PATH"
        sudo chmod +x "$GOSH_BIN_PATH"
    fi
    
    print_success "Gosh installed to $GOSH_BIN_PATH"
}

# Create configuration directories
create_config_dirs() {
    print_status "Creating configuration directories..."
    
    mkdir -p "$GOSH_CONFIG_DIR"
    mkdir -p "$HOME/.local/share/gosh"
    
    print_success "Configuration directories created"
}

# Install configuration files
install_configs() {
    print_status "Installing configuration files..."
    
    # Install .goshrc
    if [ ! -f "$HOME/.goshrc" ]; then
        if [ -f "docs/sample.goshrc" ]; then
            cp docs/sample.goshrc "$HOME/.goshrc"
            print_success "Installed .goshrc"
        else
            print_warning "Sample .goshrc not found"
        fi
    else
        print_warning ".goshrc already exists, skipping"
    fi
    
    # Install .gosh_profile
    if [ ! -f "$HOME/.gosh_profile" ]; then
        if [ -f "docs/sample.gosh_profile" ]; then
            cp docs/sample.gosh_profile "$HOME/.gosh_profile"
            print_success "Installed .gosh_profile"
        else
            print_warning "Sample .gosh_profile not found"
        fi
    else
        print_warning ".gosh_profile already exists, skipping"
    fi
    
    # Create gosh config in config directory
    if [ ! -f "$GOSH_CONFIG_DIR/goshrc" ]; then
        if [ -f "docs/sample.goshrc" ]; then
            cp docs/sample.goshrc "$GOSH_CONFIG_DIR/goshrc"
            print_success "Installed goshrc to config directory"
        fi
    fi
}

# Setup shell integration
setup_shell_integration() {
    print_status "Setting up shell integration..."
    
    # Add gosh to /etc/shells if not already there
    if [ -f "/etc/shells" ]; then
        if ! grep -q "$GOSH_BIN_PATH" /etc/shells; then
            print_status "Adding gosh to /etc/shells (requires sudo)..."
            echo "$GOSH_BIN_PATH" | sudo tee -a /etc/shells >/dev/null
            print_success "Added gosh to /etc/shells"
        else
            print_success "Gosh already in /etc/shells"
        fi
    fi
    
    # Offer to set as default shell
    echo
    read -p "Do you want to set gosh as your default shell? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Changing default shell. You can revert with: chsh -s /bin/bash"
        chsh -s "$GOSH_BIN_PATH"
        print_success "Default shell changed to gosh"
    else
        print_status "Keeping current default shell"
    fi
}

# Test installation
test_installation() {
    print_status "Testing installation..."
    
    if ! command_exists gosh; then
        print_error "Gosh not found in PATH"
        return 1
    fi
    
    # Test version command
    if gosh --version >/dev/null 2>&1; then
        VERSION=$(gosh --version | head -n1)
        print_success "Gosh is working: $VERSION"
    else
        print_error "Gosh version command failed"
        return 1
    fi
    
    # Test configuration loading
    if [ -f "$HOME/.goshrc" ]; then
        print_success "Configuration files are in place"
    else
        print_warning "Configuration files not found"
    fi
}

# Create desktop entry (Linux)
create_desktop_entry() {
    if [ "$XDG_CURRENT_DESKTOP" ] && command_exists desktop-file-install; then
        print_status "Creating desktop entry..."
        
        cat > /tmp/gosh.desktop << EOF
[Desktop Entry]
Name=Gosh Shell
Comment=A modern shell written in Go
Exec=gosh
Icon=utilities-terminal
Type=Application
Categories=System;TerminalEmulator;
Terminal=true
EOF
        
        if desktop-file-install --dir="$HOME/.local/share/applications" /tmp/gosh.desktop 2>/dev/null; then
            print_success "Desktop entry created"
        else
            print_warning "Could not create desktop entry"
        fi
        
        rm -f /tmp/gosh.desktop
    fi
}

# Print completion message
print_completion() {
    echo
    print_success "Gosh setup completed successfully!"
    echo
    echo "Next steps:"
    echo "1. Start gosh by running: gosh"
    echo "2. Customize your configuration in ~/.goshrc"
    echo "3. Read the user guide: docs/USER_GUIDE.md"
    echo
    echo "Useful commands:"
    echo "  gosh --help     Show help"
    echo "  gosh --version  Show version"
    echo "  help            Show built-in commands (in gosh)"
    echo
    echo "Configuration files:"
    echo "  ~/.goshrc       Interactive shell config"
    echo "  ~/.gosh_profile Login shell config"
    echo "  $GOSH_CONFIG_DIR/goshrc  Alternative config location"
    echo
    if [ -f "$HOME/.goshrc" ]; then
        echo "To start using gosh with your current terminal:"
        echo "  exec gosh"
    fi
    echo
}

# Main setup function
main() {
    echo "============================================"
    echo "         Gosh Shell Setup Script"
    echo "============================================"
    echo
    
    check_prerequisites
    build_gosh
    install_binary
    create_config_dirs
    install_configs
    setup_shell_integration
    test_installation
    create_desktop_entry
    print_completion
}

# Handle script interruption
trap 'print_error "Setup interrupted"; exit 1' INT TERM

# Run main function
main "$@"
