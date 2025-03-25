#!/bin/bash

# Set color output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run this script as root${NC}"
    exit 1
fi

echo -e "${GREEN}Starting WProxy uninstallation...${NC}"

# Stop service
echo "Stopping service..."
systemctl stop WProxy

# Disable service
echo "Disabling service..."
systemctl disable WProxy

# Remove service file
echo "Removing service file..."
rm -f /etc/systemd/system/WProxy.service

# Reload systemd
systemctl daemon-reload

# Remove binary file
echo "Removing program files..."
rm -f /usr/local/bin/WProxy

# Remove config files
echo "Removing configuration files..."
rm -rf /etc/WProxy

echo -e "${GREEN}Uninstallation completed!${NC}" 