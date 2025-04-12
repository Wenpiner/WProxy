package forward

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

var ErrAuthFailed = errors.New("authentication failed: incorrect username")

func HandleForward(req *http.Request, userinfo *url.Userinfo) error {
	password := ""
	if userinfo != nil {
		password, _ = userinfo.Password()
	}
	xProxyHost := req.Header.Get("x-proxy-host")
	if xProxyHost != "" {
		if password != "" {
			xProxySecret := req.Header.Get("x-proxy-secret")
			if xProxySecret != password {
				return ErrAuthFailed
			}
		}

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
