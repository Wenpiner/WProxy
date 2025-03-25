package http

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type HttpConn struct {
	net.Conn
	UserInfo *url.Userinfo
}

func (httpConn *HttpConn) Close() error {
	return httpConn.Conn.Close()
}

func (httpConn *HttpConn) Write(byte []byte) (int, error) {
	return httpConn.Conn.Write(byte)
}

func (httpConn *HttpConn) Read(b []byte) (n int, err error) {
	n, err = httpConn.Conn.Read(b)
	return
}

func (httpConn *HttpConn) Handshake() error {
	reqBufReader := bufio.NewReader(io.NopCloser(httpConn.Conn))
	req, err := http.ReadRequest(reqBufReader)
	if err != nil {
		return err
	}
	if httpConn.UserInfo != nil {
		get := req.Header.Get("Proxy-Authorization")
		// Parse password
		user, password, ok := parseBasicAuth(get)
		if !ok {
			return errors.New("parse password error")
		}
		s, _ := httpConn.UserInfo.Password()
		if user != httpConn.UserInfo.Username() || password != s {
			return errors.New("password error")
		}
	}
	var buf bytes.Buffer
	err = req.Write(&buf)
	if err != nil {
		log.Println("Not a valid HTTP request")
		return err
	}

	// Parse hostname
	host := req.URL.Hostname()
	port := req.URL.Port()
	if port == "" {
		port = "80"
		if req.URL.Scheme == "https" {
			port = "443"
		}
	}
	// Print parsed domain and port
	fmt.Println("Domain:", host, "Port:", port)
	dial, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return err
	}

	if strings.ToUpper(req.Method) == "CONNECT" {
		// CONNECT
		_, err = fmt.Fprintf(httpConn.Conn, "HTTP/1.1 200 Connection established\r\n\r\n")
		if err != nil {
			return err
		}
	} else {
		// HTTP
		_, err = dial.Write(buf.Bytes())
		if err != nil {
			return err
		}
	}
	go io.Copy(httpConn.Conn, dial)
	io.Copy(dial, httpConn.Conn)
	return nil
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !EqualFold(auth[:len(prefix)], prefix) {
		return "", "", false
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}

// EqualFold is strings.EqualFold, ASCII only. It reports whether s and t
// are equal, ASCII-case-insensitively.
func EqualFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if lower(s[i]) != lower(t[i]) {
			return false
		}
	}
	return true
}

// lower returns the ASCII lowercase version of b.
func lower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
