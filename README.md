# WProxy

WProxy 是一个带鉴权的 SOCKS5 和 HTTP 代理,支持端口复用的工具。它可以帮助您在安全和高效的环境下上网。

# 特性

- [x] 支持身份验证,可以防止未经授权的访问
- [x] 支持端口复用,可以复用同一个端口提供多种代理服务
- [x] 轻量级和跨平台,可在 Windows、Linux 和 macOS 上运行
- [x] 支持 SOCKS5 和 HTTP 代理协议

# 安装

您可以通过以下步骤安装 WProxy:

- 从 GitHub 仓库下载最新版本的 WProxy 二进制文件。
- 将下载的文件解压缩到您的系统中。
- 在命令行中运行 ./wproxy 命令启动代理服务器。

# 配置

您可以根据需要进行以下设置:

- listen_addr: 代理服务器的监听地址,默认为 0.0.0.0:1080。
- username 和 password: 身份验证所需的用户名和密码。

# 使用

- 将您的应用程序或浏览器配置为使用 WProxy 作为代理服务器。
- 输入代理服务器的地址和端口,以及身份验证所需的用户名和密码(如果已启用)。
- 开始通过代理服务器访问互联网。

# 贡献

如果您发现任何问题或有任何改进建议,欢迎提交 issue 或 pull request。我们很高兴能与社区一起改进 WProxy。

# 许可证

WProxy 基于 MIT 许可证发布,您可以自由使用、修改和分发本项目。


## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Wenpiner/WProxy&type=Date)](https://star-history.com/#Wenpiner/WProxy&Date)
