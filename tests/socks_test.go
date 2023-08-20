package tests

import (
	"fmt"
	"golang.org/x/net/proxy"
	"net/http"
	"testing"
)

func TestSocks(t *testing.T) {
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:10801", nil, proxy.Direct)
	if err != nil {
		fmt.Println("Error creating SOCKS5 dialer:", err)
		return
	}

	// 使用代理 Dialer 创建 HTTP 客户端
	httpTransport := &http.Transport{Dial: dialer.Dial}
	httpClient := &http.Client{Transport: httpTransport}

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", "https://www.baidu.com", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// 处理响应
	fmt.Println("Response status:", resp.Status)
	// 在这里可以处理响应内容等
}
