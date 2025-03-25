#!/bin/bash

# Set color output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run this script with root privileges${NC}"
    exit 1
fi

# Get system architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported system architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# Create installation directory
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/WProxy"

echo -e "${GREEN}Starting WProxy installation...${NC}"

# Create config directory
mkdir -p $CONFIG_DIR

# Download latest version
echo "Fetching latest version information..."
LATEST_VERSION=$(curl -s https://api.github.com/repos/Wenpiner/WProxy/releases/latest | grep "tag_name" | cut -d'"' -f4)

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${RED}Failed to get latest version information${NC}"
    exit 1
fi

echo "Downloading WProxy $LATEST_VERSION..."
DOWNLOAD_URL="https://github.com/Wenpiner/WProxy/releases/download/$LATEST_VERSION/WProxy-$LATEST_VERSION-linux-$ARCH.tar.gz"
wget -q $DOWNLOAD_URL -O /tmp/WProxy.tar.gz

if [ $? -ne 0 ]; then
    echo -e "${RED}Download failed${NC}"
    exit 1
fi

# Extract files
echo "Extracting files..."
tar -xzf /tmp/WProxy.tar.gz -C /tmp

# Move binary to system directory
echo "Installing..."
mv /tmp/WProxy $INSTALL_DIR/
chmod +x $INSTALL_DIR/WProxy

# Generate random password
RANDOM_PASSWORD=$(openssl rand -base64 12 | tr -d '/+=' | head -c 16)

# Create config file
cat > $CONFIG_DIR/config.yaml << EOF
listen_addr: "0.0.0.0:1080"
username: "admin"
password: "${RANDOM_PASSWORD}"
EOF

# Create system service
cat > /etc/systemd/system/WProxy.service << EOF
[Unit]
Description=WProxy Service
After=network.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/WProxy -c $CONFIG_DIR/config.yaml
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

# Start service
echo "Starting service..."
systemctl enable WProxy
systemctl start WProxy

# Clean up temporary files
rm -f /tmp/WProxy.tar.gz

echo -e "${GREEN}Installation completed!${NC}"
echo -e "WProxy installed at: $INSTALL_DIR/WProxy"
echo -e "Config file location: $CONFIG_DIR/config.yaml"
echo -e "Default username: admin"
echo -e "Randomly generated password: ${RANDOM_PASSWORD}"
echo -e "Please keep your password safe!"
echo -e "Use the following command to check service status: systemctl status WProxy" 