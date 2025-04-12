# WProxy

WProxy 是一个带鉴权的 SOCKS5 和 HTTP 代理工具，支持端口复用。它可以帮助您在安全和高效的环境下上网。

[English](README.md) | [中文文档](README_zh-CN.md)

# 特性

- [x] 支持身份验证，可以防止未经授权的访问
- [x] 支持端口复用，可以复用同一个端口提供多种代理服务
- [x] 轻量级和跨平台，可在 Windows、Linux 和 macOS 上运行
- [x] 支持 SOCKS5 和 HTTP/HTTPS 代理协议
- [x] 支持通过 HTTP/HTTPS 请求头指定目标域名或 IP 进行转发

# 快速安装

## Linux 系统一键安装

使用以下命令快速安装 WProxy：

```bash
curl -s https://raw.githubusercontent.com/Wenpiner/WProxy/main/install.sh | sudo bash
```

安装脚本会自动完成以下操作：
1. 检测系统架构（支持 amd64 和 arm64）
2. 从 GitHub 下载最新版本
3. 安装到系统目录
4. 创建配置文件
5. 设置系统服务（支持开机自启）

安装完成后，您可以通过以下命令查看服务状态：
```bash
systemctl status wproxy
```

## 卸载

使用以下命令快速卸载 WProxy：

```bash
curl -s https://raw.githubusercontent.com/Wenpiner/WProxy/main/uninstall.sh | sudo bash
```

卸载脚本会自动完成以下操作：
1. 停止并禁用 WProxy 服务
2. 删除系统服务文件
3. 删除程序文件
4. 删除配置文件

## 手动安装

如果您想手动安装，可以按照以下步骤操作：

1. 从 GitHub 仓库下载最新版本的 WProxy 二进制文件
2. 将下载的文件解压缩到您的系统中
3. 在命令行中运行 `./wproxy` 命令启动代理服务器

# 配置

配置文件位于 `/etc/wproxy/config.yaml`，您可以根据需要进行以下设置：

- listen_addr: 代理服务器的监听地址，默认为 0.0.0.0:1080
- username 和 password: 身份验证所需的用户名和密码

默认配置：
```yaml
listen_addr: "0.0.0.0:1080"
username: "admin"
password: "16位随机密码"  # 安装时自动生成
```

注意事项：
1. 安装时会自动生成一个随机的16位密码
2. 安装完成后会显示生成的密码，请务必妥善保管
3. 如需修改密码，请编辑配置文件后重启服务

# 使用

- 将您的应用程序或浏览器配置为使用 WProxy 作为代理服务器
- 输入代理服务器的地址和端口，以及身份验证所需的用户名和密码（如果已启用）
- 开始通过代理服务器访问互联网
- 如果需要通过 HTTP/HTTPS 请求头指定目标域名或 IP 进行转发，请在请求头中添加 `X-Proxy-Host`、`X-Proxy-Scheme` 和 `X-Proxy-Secret` 字段。例如：
  ### 无鉴权、HTTP 转发
  ```http
  GET /xxx/xxx HTTP/1.1
  Host: example.com
  X-Proxy-Host: target-domain.com
  X-Proxy-Scheme: http
  ```
  ### 无鉴权、HTTP 转发、自定义端口(非TLS)
  ```http
  GET /xxx/xxx HTTP/1.1
  Host: example.com
  X-Proxy-Host: target-domain.com:8080
  X-Proxy-Scheme: http
  ```

  ### 无鉴权、HTTPS 转发、自定义端口(TLS)
  ```http
  GET /xxx/xxx HTTP/1.1
  Host: example.com
  X-Proxy-Host: target-domain.com:8443
  X-Proxy-Scheme: https
  ```

  ### 鉴权、HTTPS 转发 
  ```http
  GET /xxx/xxx HTTP/1.1
  Host: example.com
  X-Proxy-Host: target-domain.com:8443
  X-Proxy-Scheme: https
  X-Proxy-Secret: your_password
  ```


  代理服务器会根据这些字段将请求转发到指定的目标地址和端口。
### ⚠️ 注意事项
1. `X-Proxy-Secret`对应就是配置文件中的`password`，如果没有设置密码，则不需要设置该字段。

# 服务管理

安装后，您可以使用以下命令管理 WProxy 服务：

```bash
# 启动服务
sudo systemctl start wproxy

# 停止服务
sudo systemctl stop wproxy

# 重启服务
sudo systemctl restart wproxy

# 查看服务状态
sudo systemctl status wproxy

# 设置开机自启
sudo systemctl enable wproxy

# 禁用开机自启
sudo systemctl disable wproxy
```

# 贡献

如果您发现任何问题或有任何改进建议，欢迎提交 issue 或 pull request。我们很高兴能与社区一起改进 WProxy。

# 许可证

WProxy 基于 MIT 许可证发布，您可以自由使用、修改和分发本项目。

## Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=Wenpiner/WProxy&type=Date)](https://star-history.com/#Wenpiner/WProxy&Date) 
`