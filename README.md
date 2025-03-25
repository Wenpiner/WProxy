# WProxy

WProxy is an authenticated SOCKS5 and HTTP proxy tool with port multiplexing support. It helps you access the internet in a secure and efficient environment.

[中文文档](README_zh-CN.md) | [English](README.md)

# Features

- [x] Authentication support to prevent unauthorized access
- [x] Port multiplexing to provide multiple proxy services on the same port
- [x] Lightweight and cross-platform, runs on Windows, Linux, and macOS
- [x] Supports SOCKS5 and HTTP proxy protocols

# Quick Installation

## One-Click Installation for Linux

Use the following command to quickly install WProxy:

```bash
curl -s https://raw.githubusercontent.com/Wenpiner/WProxy/main/install.sh | sudo bash
```

The installation script will automatically:
1. Detect system architecture (supports amd64 and arm64)
2. Download the latest version from GitHub
3. Install to system directory
4. Create configuration file
5. Set up system service (auto-start on boot)

After installation, you can check the service status with:
```bash
systemctl status wproxy
```

## Uninstallation

Use the following command to quickly uninstall WProxy:

```bash
curl -s https://raw.githubusercontent.com/Wenpiner/WProxy/main/uninstall.sh | sudo bash
```

The uninstallation script will automatically:
1. Stop and disable WProxy service
2. Remove system service file
3. Remove program files
4. Remove configuration files

## Manual Installation

If you prefer manual installation, follow these steps:

1. Download the latest WProxy binary from GitHub repository
2. Extract the downloaded file to your system
3. Run `./wproxy` command to start the proxy server

# Configuration

The configuration file is located at `/etc/wproxy/config.yaml`. You can configure the following settings:

- listen_addr: Proxy server listening address, default is 0.0.0.0:1080
- username and password: Authentication credentials

Default configuration:
```yaml
listen_addr: "0.0.0.0:1080"
username: "admin"
password: "16-digit random password"  # Generated during installation
```

Notes:
1. A random 16-digit password is generated during installation
2. The generated password will be displayed after installation, please keep it safe
3. To change the password, edit the config file and restart the service

# Usage

- Configure your applications or browsers to use WProxy as the proxy server
- Enter the proxy server address, port, and authentication credentials (if enabled)
- Start accessing the internet through the proxy server

# Service Management

After installation, you can manage the WProxy service using these commands:

```bash
# Start service
sudo systemctl start wproxy

# Stop service
sudo systemctl stop wproxy

# Restart service
sudo systemctl restart wproxy

# Check service status
sudo systemctl status wproxy

# Enable auto-start
sudo systemctl enable wproxy

# Disable auto-start
sudo systemctl disable wproxy
```

# Contributing

If you find any issues or have suggestions for improvements, feel free to submit issues or pull requests. We welcome community contributions to improve WProxy.

# License

WProxy is released under the MIT License. You are free to use, modify, and distribute this project.

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Wenpiner/WProxy&type=Date)](https://star-history.com/#Wenpiner/WProxy&Date)
