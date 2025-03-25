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

	// Create HTTP client with proxy dialer
	httpTransport := &http.Transport{Dial: dialer.Dial}
	httpClient := &http.Client{Transport: httpTransport}

	// Create HTTP request
	req, err := http.NewRequest("GET", "https://www.baidu.com", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Handle response
	fmt.Println("Response status:", resp.Status)
	// Process response content here
}
