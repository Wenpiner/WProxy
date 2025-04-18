package forward

import (
	"errors"
	"net/http"
	"strings"
)

var LoopDetected = errors.New("loop detected: proxy loop")

func HandleForward(req *http.Request) error {
	// loop check
	loop := req.Header.Get("x-proxy-loop")
	if loop != "" {
		return LoopDetected
	}

	// write loop check
	req.Header.Set("x-proxy-loop", "1")

	xProxyHost := req.Header.Get("x-proxy-host")
	if xProxyHost != "" {

		newScheme := "https"
		headerScheme := req.Header.Get("x-proxy-scheme")
		if headerScheme == "http" {
			newScheme = "http"
		}

		// handleHost
		host := xProxyHost
		req.Host = handleHost(newScheme, host)
		req.URL.Scheme = newScheme
		req.URL.Host = host
	} else {
		req.Host = handleHost(req.URL.Scheme, req.Host)
		req.URL.Host = req.Host
	}

	return nil
}

func handleHost(scheme, host string) string {
	if !strings.Contains(host, ":") {
		if scheme == "https" {
			host = host + ":443"
		} else {
			host = host + ":80"
		}
	}
	return host
}
