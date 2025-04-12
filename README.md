
# WProxy

WProxy is an authenticated SOCKS5 and HTTP proxy tool that supports port multiplexing. It helps you browse the internet securely and efficiently.

[English](README.md) | [中文文档](README_zh-CN.md)

# Features

- [x] Supports authentication to prevent unauthorized access
- [x] Supports port multiplexing, allowing multiple proxy services on a single port
- [x] Lightweight and cross-platform, runs on Windows, Linux and macOS
- [x] Supports SOCKS5 and HTTP/HTTPS proxy protocols
- [x] Supports forwarding to target domains or IPs specified in HTTP/HTTPS headers

# Quick Installation

## One-click Installation for Linux

Use the following command to quickly install WProxy:

```bash
curl -s https://raw.githubusercontent.com/Wenpiner/WProxy/main/install.sh | sudo bash
```

The installation script will automatically:
1. Detect system architecture (supports amd64 and arm64)
2. Download the latest version from GitHub
3. Install to the system directory
4. Create a configuration file
5. Set up a system service (with auto-start on boot)

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
1. Stop and disable the WProxy service
2. Remove the system service file
3. Delete the program files
4. Remove the configuration files

## Manual Installation

If you prefer to install manually, follow these steps:

1. Download the latest WProxy binary from the GitHub repository
2. Extract the downloaded file to your system
3. Run the `./wproxy` command in the terminal to start the proxy server

# Configuration

The configuration file is located at `/etc/wproxy/config.yaml`. You can configure the following settings:

- listen_addr: The listening address of the proxy server, default is 0.0.0.0:1080
- username and password: The credentials required for authentication

Default configuration:
```yaml
listen_addr: "0.0.0.0:1080"
username: "admin"
password: "16-character random password"  # Automatically generated during installation
```

Notes:
1. A random 16-character password is automatically generated during installation
2. The generated password will be displayed after installation, please keep it safe
3. To change the password, edit the configuration file and restart the service

# Usage

- Configure your applications or browsers to use WProxy as a proxy server
- Enter the proxy server's address, port, and authentication credentials (if enabled)
- Start browsing the internet through the proxy server
- If you need to forward to target domains or IPs specified in HTTP/HTTPS headers, add the `X-Proxy-Host`, `X-Proxy-Scheme`, and `X-Proxy-Secret` fields to the request headers. For example:
  ### Unauthenticated HTTP forwarding
  ```http
  GET /xxx/xxx HTTP/1.1
  Host: example.com
  X-Proxy-Host: target-domain.com
  X-Proxy-Scheme: http
  ```
  ### Unauthenticated HTTP forwarding with custom port (non-TLS)
  ```http
  GET /xxx/xxx HTTP/1.1
  Host: example.com
  X-Proxy-Host: target-domain.com:8080
  X-Proxy-Scheme: http
  ```

  ### Unauthenticated HTTPS forwarding with custom port (TLS)
  ```http
  GET /xxx/xxx HTTP/1.1
  Host: example.com
  X-Proxy-Host: target-domain.com:8443
  X-Proxy-Scheme: https
  ```

  ### Authenticated HTTPS forwarding
  ```http
  GET /xxx/xxx HTTP/1.1
  Host: example.com
  X-Proxy-Host: target-domain.com:8443
  X-Proxy-Scheme: https
  X-Proxy-Secret: your_password
  ```

  The proxy server will forward the request to the specified target address and port based on these fields.

### ⚠️ Notes
1. `X-Proxy-Secret` corresponds to the `password` in the configuration file. If no password is set, this field is not required.

# Service Management

After installation, you can manage the WProxy service with the following commands:

```bash
# Start the service
sudo systemctl start wproxy

# Stop the service
sudo systemctl stop wproxy

# Restart the service
sudo systemctl restart wproxy

# Check service status
sudo systemctl status wproxy

# Enable auto-start on boot
sudo systemctl enable wproxy

# Disable auto-start on boot
sudo systemctl disable wproxy
```

# Contributing

If you find any issues or have suggestions for improvements, feel free to submit an issue or pull request. We're happy to improve WProxy together with the community.

# License

WProxy is released under the MIT License, allowing you to freely use, modify, and distribute this project.

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Wenpiner/WProxy&type=Date)](https://star-history.com/#Wenpiner/WProxy&Date)
