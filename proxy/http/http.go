package http

import (
	"WProxy/common"
	"WProxy/proxy/forward"
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Error constants
var (
	ErrAuthRequired = errors.New("authentication failed: authentication required")
	ErrAuthFailed   = errors.New("authentication failed: incorrect username or password")
)

// Conn HttpConn wraps the underlying connection
type Conn struct {
	UserInfo *url.Userinfo
	// Client Conn
	clientConn net.Conn
	// Target Conn
	targetConn net.Conn
}

func HandleConn(client net.Conn, userinfo *url.Userinfo, cert *tls.Certificate, isTls bool) error {
	httpConn := Conn{
		UserInfo: userinfo,
	}
	if isTls {
		tlsConn := tls.Server(client, &tls.Config{
			Certificates: []tls.Certificate{*cert},
		})
		err := tlsConn.Handshake()
		if err != nil {
			return err
		}
		httpConn.clientConn = tlsConn
	} else {
		httpConn.clientConn = client
	}

	reqBufReader := bufio.NewReader(httpConn.clientConn)
	req, err := http.ReadRequest(reqBufReader)
	if err != nil {
		return fmt.Errorf("failed to parse HTTP request: %w", err)
	}
	defer req.Body.Close()

	if httpConn.UserInfo != nil {
		authHeader := req.Header.Get("Proxy-Authorization")
		if authHeader == "" {
			// Return 407 requesting authentication
			resp := "HTTP/1.1 407 Proxy Authentication Required\r\n" +
				"Proxy-Authenticate: Basic realm=\"Proxy\"\r\n" +
				"Content-Length: 0\r\n\r\n"
			_, err := httpConn.clientConn.Write([]byte(resp))
			if err != nil {
				return err
			}
			return ErrAuthRequired
		}

		user, password, ok := parseBasicAuth(authHeader)
		if !ok {
			return ErrAuthFailed
		}

		expectedPass, _ := httpConn.UserInfo.Password()
		if user != httpConn.UserInfo.Username() || password != expectedPass {
			return ErrAuthFailed
		}
	}

	if strings.ToUpper(req.Method) == "CONNECT" {
		return httpConn.handleConnect(req)
	} else {
		return httpConn.handleHTTPRequest(req)
	}
}

// handleConnect processes CONNECT requests
func (httpConn *Conn) handleConnect(req *http.Request) error {

	err := forward.HandleForward(req)
	if err != nil {
		return err
	}

	// Connect to target server
	targetConn, err := net.DialTimeout("tcp", req.Host, 10*time.Second)
	if err != nil {
		// Return error response
		errResp := fmt.Sprintf("HTTP/1.1 502 Bad Gateway\r\nContent-Length: %d\r\n\r\n%s",
			len(err.Error()), err.Error())
		httpConn.clientConn.Write([]byte(errResp))
		return fmt.Errorf("failed to connect to target server: %w", err)
	}
	defer targetConn.Close()

	// Send successful connection response
	_, err = httpConn.clientConn.Write([]byte(fmt.Sprintf("HTTP/%d.%d 200 Connection established\r\n\r\n", req.ProtoMajor, req.ProtoMinor)))
	if err != nil {
		return fmt.Errorf("failed to send success response: %w", err)
	}

	// Bidirectional data copy
	return tunnelConnection(httpConn.clientConn, targetConn)
}

// handleHTTPRequest processes regular HTTP requests
func (httpConn *Conn) handleHTTPRequest(req *http.Request) error {

	err := forward.HandleForward(req)
	if err != nil {
		return err
	}
	// Connect to target server
	targetConn, err := net.DialTimeout("tcp", req.Host, 10*time.Second)
	if err != nil {
		errResp := fmt.Sprintf("HTTP/1.1 502 Bad Gateway\r\nContent-Length: %d\r\n\r\n%s",
			len(err.Error()), err.Error())
		_, err := httpConn.clientConn.Write([]byte(errResp))
		if err != nil {
			log.Println("Failed to send error response:", err)
			return err
		}
		return fmt.Errorf("failed to connect to target server: %w", err)
	}
	defer targetConn.Close()

	if req.URL.Scheme == "https" {
		tlsConn := tls.Client(targetConn, &tls.Config{
			ServerName: req.URL.Hostname(),
		})
		err = tlsConn.Handshake()
		if err != nil {
			return fmt.Errorf("failed to perform TLS handshake: %w", err)
		}
		targetConn = tlsConn
	}

	// Modify request to send to target server
	// req.RequestURI = ""
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Proxy-Connection")

	// Write request to target server
	err = req.Write(targetConn)
	if err != nil {
		return fmt.Errorf("failed to send request to target server: %w", err)
	}

	// Read response from target server and forward to client
	return tunnelConnection(httpConn.clientConn, targetConn)
}

// tunnelConnection forwards data bidirectionally between two connections
func tunnelConnection(client, target net.Conn) error {
	errCh := make(chan error, 2)
	go func() {
		_, err := io.Copy(target, client)
		if closer, ok := target.(common.CloseWriter); ok {
			_ = closer.CloseWrite()
		}
		errCh <- err
	}()

	go func() {
		_, err := io.Copy(client, target)
		if closer, ok := target.(common.CloseWriter); ok {
			_ = closer.CloseWrite()
		}
		errCh <- err
	}()

	// Wait for error or completion in either direction
	err := <-errCh
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	// Wait for the other direction to complete
	err = <-errCh
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

// parseBasicAuth parses HTTP basic authentication string
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
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

// EqualFold determines if two strings are equal (case-insensitive)
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

// lower returns the lowercase version of a byte
func lower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
