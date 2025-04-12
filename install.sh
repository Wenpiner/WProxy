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
CONFIG_DIR="/etc/wproxy"

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

# 获取当前平台
# 获取当前操作系统平台
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case $OS in
    linux)
        ;;
    darwin)
        ;;
    *)
        echo -e "${RED}当前仅支持 Linux 和 macOS 平台${NC}"
        exit 1
        ;;
esac

echo "Downloading WProxy $LATEST_VERSION..."
# 显示系统信息用于调试
echo "System Info:"
echo "- OS: $OS"
echo "- Architecture: $ARCH"

DOWNLOAD_URL="https://github.com/Wenpiner/WProxy/releases/download/$LATEST_VERSION/WProxy-${LATEST_VERSION}-${OS}-${ARCH}.tar.gz"
echo "Download URL: $DOWNLOAD_URL"

# 使用 curl 替代 wget，提供更好的错误处理
echo "Downloading..."
if ! curl -L --fail "$DOWNLOAD_URL" -o /tmp/WProxy.tar.gz; then
    echo -e "${RED}Download failed. Please check the URL and try again.${NC}"
    exit 1
fi

# 验证下载文件是否存在且大小不为0
if [ ! -s /tmp/WProxy.tar.gz ]; then
    echo -e "${RED}Downloaded file is empty or does not exist${NC}"
    exit 1
fi

# Extract files
echo "Extracting files..."
tar -xzvf /tmp/WProxy.tar.gz -C /tmp

if [ $? -ne 0 ]; then
    echo -e "${RED}Extraction failed${NC}"
    exit 1
fi

# 验证解压后的文件是否存在
if [ ! -f /tmp/wproxy ]; then
    echo -e "${RED}Extracted binary not found${NC}"
    exit 1
fi

# Move binary to system directory
echo "Installing..."
mv /tmp/wproxy $INSTALL_DIR/
chmod +x $INSTALL_DIR/wproxy

# Generate random password
RANDOM_PASSWORD=$(openssl rand -base64 12 | tr -d '/+=' | head -c 16)

# Create config file
cat > $CONFIG_DIR/config.yaml << EOF
listen_addr: "0.0.0.0:1080"
username: "admin"
password: "${RANDOM_PASSWORD}"
EOF

# Create system service
cat > /etc/systemd/system/wproxy.service << EOF
[Unit]
Description=WProxy Service
After=network.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/wproxy -c $CONFIG_DIR/config.yaml
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

# Start service
echo "Starting service..."
systemctl enable wproxy
systemctl start wproxy

# Clean up temporary files
rm -f /tmp/WProxy.tar.gz

echo -e "${GREEN}Installation completed!${NC}"
echo -e "WProxy installed at: $INSTALL_DIR/wproxy"
echo -e "Config file location: $CONFIG_DIR/config.yaml"
echo -e "Default username: admin"
echo -e "Randomly generated password: ${RANDOM_PASSWORD}"
echo -e "Please keep your password safe!"
echo -e "Use the following command to check service status: systemctl status WProxy" 