#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="berbyte/sinkzone"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="sinkzone"
SERVICE_NAME="sinkzone"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64)
            ARCH="x86_64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    echo "${OS}_${ARCH}"
}

# Function to get latest version
get_latest_version() {
    curl -s "https://api.github.com/repos/$REPO/releases/latest" | \
    grep '"tag_name":' | \
    sed -E 's/.*"([^"]+)".*/\1/'
}

# Function to download binary
download_binary() {
    local version=$1
    local platform=$2
    local download_url="https://github.com/$REPO/releases/download/$version/sinkzone_${platform}.tar.gz"
    
    print_status "Downloading sinkzone $version for $platform..."
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Download and extract
    curl -L -o sinkzone.tar.gz "$download_url"
    tar -xzf sinkzone.tar.gz
    
    # Move binary to install directory
    sudo mv sinkzone "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/sinkzone"
    
    # Cleanup
    cd - > /dev/null
    rm -rf "$TEMP_DIR"
}

# Function to create systemd service
create_service() {
    print_status "Creating systemd service..."
    
    sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null <<EOF
[Unit]
Description=Sinkzone DNS Filter
After=network.target

[Service]
Type=simple
User=root
ExecStart=$INSTALL_DIR/sinkzone dns start
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable $SERVICE_NAME
}

# Function to setup DNS configuration
setup_dns() {
    print_status "Setting up DNS configuration..."
    
    # Check if NetworkManager is available
    if command -v nmcli >/dev/null 2>&1; then
        print_status "Configuring NetworkManager DNS..."
        sudo nmcli connection modify "$(nmcli -t -f UUID,TYPE,DEVICE connection show --active | grep ethernet | cut -d: -f1)" ipv4.dns "127.0.0.1"
        sudo nmcli connection modify "$(nmcli -t -f UUID,TYPE,DEVICE connection show --active | grep ethernet | cut -d: -f1)" ipv4.ignore-auto-dns yes
        sudo systemctl restart NetworkManager
    else
        print_warning "NetworkManager not found. Please manually configure your DNS to use 127.0.0.1"
    fi
}

# Main installation function
main() {
    print_status "Installing Sinkzone DNS Filter..."
    
    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1; then
        print_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v tar >/dev/null 2>&1; then
        print_error "tar is required but not installed"
        exit 1
    fi
    
    # Detect platform
    PLATFORM=$(detect_platform)
    print_status "Detected platform: $PLATFORM"
    
    # Get latest version
    VERSION=$(get_latest_version)
    print_status "Latest version: $VERSION"
    
    # Download and install binary
    download_binary "$VERSION" "$PLATFORM"
    
    # Create systemd service
    create_service
    
    # Setup DNS configuration
    setup_dns
    
    print_status "Installation completed successfully!"
    print_status "Sinkzone has been installed to $INSTALL_DIR/$BINARY_NAME"
    print_status "Service has been created and enabled"
    print_status ""
    print_status "To start the service:"
    echo "  sudo systemctl start $SERVICE_NAME"
    print_status ""
    print_status "To check status:"
    echo "  sudo systemctl status $SERVICE_NAME"
    print_status ""
    print_status "To view logs:"
    echo "  sudo journalctl -u $SERVICE_NAME -f"
    print_status ""
    print_status "To configure sinkzone:"
    echo "  $BINARY_NAME mode focus    # Enable focus mode"
    echo "  $BINARY_NAME allow example.com    # Add to allowlist"
    echo "  $BINARY_NAME block twitter.com    # Add to blocklist"
}

# Run main function
main "$@" 