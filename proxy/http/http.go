package http

import (
	"WProxy/common"
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
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
	ErrAuthRequired   = errors.New("Authentication failed: authentication required")
	ErrAuthFailed     = errors.New("Authentication failed: incorrect username or password")
	ErrInvalidRequest = errors.New("Invalid HTTP request")
	ErrInvalidTLS     = errors.New("Invalid TLS handshake")
)

// HttpConn wraps the underlying connection
type HttpConn struct {
	net.Conn
	UserInfo *url.Userinfo
	// Add timeout setting
	Timeout time.Duration
}

// Close closes the connection
func (httpConn *HttpConn) Close() error {
	return httpConn.Conn.Close()
}

// Handshake handles the proxy handshake process
func (httpConn *HttpConn) Handshake(tls bool, cert *tls.Certificate) error {
	// Set read timeout
	if httpConn.Timeout > 0 {
		httpConn.Conn.SetReadDeadline(time.Now().Add(httpConn.Timeout))
	}

	// Determine if HTTP or TLS based on first byte
	if tls {
		//err := httpConn.handleClientTlsHandshake(cert)
		//if err != nil {
		//	return err
		//}
		// TLS connection
		return httpConn.handleTLSConnection()
	}

	// HTTP connection
	return httpConn.handleHTTPConnection()
}

// handleClientTlsHandshake
func (httpConn *HttpConn) handleClientTlsHandshake(cert *tls.Certificate) error {
	// 使用 tls.Server 包装连接
	tlsConfig := &tls.Config{
		// 配置证书和密钥
		Certificates: []tls.Certificate{
			*cert,
		},
	}
	tlsConn := tls.Server(httpConn.Conn, tlsConfig)

	// 完成 TLS 握手
	if err := tlsConn.Handshake(); err != nil {
		log.Println("TLS handshake error:", err)
		return err
	}
	httpConn.Conn = tlsConn
	return nil
}

// handleHTTPConnection processes HTTP connections
func (httpConn *HttpConn) handleHTTPConnection() error {
	reqBufReader := bufio.NewReader(httpConn.Conn)
	req, err := http.ReadRequest(reqBufReader)
	if err != nil {
		return fmt.Errorf("failed to parse HTTP request: %w", err)
	}
	defer req.Body.Close()

	// Handle authentication
	if httpConn.UserInfo != nil {
		authHeader := req.Header.Get("Proxy-Authorization")
		if authHeader == "" {
			// Return 407 requesting authentication
			resp := "HTTP/1.1 407 Proxy Authentication Required\r\n" +
				"Proxy-Authenticate: Basic realm=\"Proxy\"\r\n" +
				"Content-Length: 0\r\n\r\n"
			httpConn.Conn.Write([]byte(resp))
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
func (httpConn *HttpConn) handleConnect(req *http.Request) error {
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}

	// Ensure hostname has correct format
	if !strings.Contains(host, ":") {
		host = host + ":443" // Default HTTPS port
	}

	log.Printf("CONNECT request: %s", host)

	// Connect to target server
	targetConn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		// Return error response
		errResp := fmt.Sprintf("HTTP/1.1 502 Bad Gateway\r\nContent-Length: %d\r\n\r\n%s",
			len(err.Error()), err.Error())
		httpConn.Conn.Write([]byte(errResp))
		return fmt.Errorf("failed to connect to target server: %w", err)
	}
	defer targetConn.Close()

	// Send successful connection response
	_, err = httpConn.Conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		return fmt.Errorf("failed to send success response: %w", err)
	}

	// Bidirectional data copy
	return tunnelConnection(httpConn.Conn, targetConn)
}

// handleHTTPRequest processes regular HTTP requests
func (httpConn *HttpConn) handleHTTPRequest(req *http.Request) error {
	// Build target URL
	targetURL := *req.URL
	if targetURL.Scheme == "" {
		targetURL.Scheme = "http"
	}

	// Ensure Host header exists
	if req.Host != "" && req.URL.Host == "" {
		targetURL.Host = req.Host
	}

	host := targetURL.Host
	if !strings.Contains(host, ":") {
		if targetURL.Scheme == "https" {
			host = host + ":443"
		} else {
			host = host + ":80"
		}
	}

	log.Printf("HTTP request: %s %s", req.Method, targetURL.String())

	// Connect to target server
	targetConn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		errResp := fmt.Sprintf("HTTP/1.1 502 Bad Gateway\r\nContent-Length: %d\r\n\r\n%s",
			len(err.Error()), err.Error())
		httpConn.Conn.Write([]byte(errResp))
		return fmt.Errorf("failed to connect to target server: %w", err)
	}
	defer targetConn.Close()

	// Modify request to send to target server
	req.RequestURI = ""
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Proxy-Connection")

	// Write request to target server
	err = req.Write(targetConn)
	if err != nil {
		return fmt.Errorf("failed to send request to target server: %w", err)
	}

	// Read response from target server and forward to client
	return tunnelConnection(httpConn.Conn, targetConn)
}

// handleTLSConnection processes TLS connections
func (httpConn *HttpConn) handleTLSConnection() error {
	// Extract SNI information from TLS ClientHello
	hostname, err := getHostFromTLS(httpConn.Conn)
	if err != nil {
		return fmt.Errorf("failed to parse TLS handshake: %w", err)
	}

	if hostname == "" {
		return ErrInvalidTLS
	}

	// Ensure hostname includes port
	if !strings.Contains(hostname, ":") {
		hostname = hostname + ":443"
	}

	log.Printf("TLS connection to: %s", hostname)

	// Connect to target server
	targetConn, err := net.DialTimeout("tcp", hostname, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to target server: %w", err)
	}
	defer targetConn.Close()

	// Note: We're not checking authentication here, as we can't get auth info from TLS handshake
	// If authentication is needed, it should be handled at the proxy server level

	// Bidirectional data forwarding
	return tunnelConnection(httpConn.Conn, targetConn)
}

// getHostFromTLS extracts SNI information from TLS ClientHello
func getHostFromTLS(reader io.Reader) (string, error) {
	// Read TLS record header
	header := make([]byte, 5)
	if _, err := io.ReadFull(reader, header); err != nil {
		return "", err
	}

	// Check if it's a TLS handshake
	if header[0] != 0x16 { // 0x16 is the TLS handshake identifier
		return "", errors.New("not a TLS handshake")
	}

	// Calculate handshake message length
	length := int(binary.BigEndian.Uint16(header[3:5]))
	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return "", err
	}

	// Check if it's a ClientHello message
	if data[0] != 0x01 { // 0x01 is the ClientHello message type
		return "", errors.New("not a ClientHello message")
	}
	// Skip unnecessary fields
	pos := 34 // 1(msg type) + 3(len) + 2(version) + 32(random)

	// Skip session ID
	if pos+1 > len(data) {
		return "", errors.New("handshake message incomplete")
	}
	sessionIDLen := int(data[pos])
	pos += 1 + sessionIDLen

	// Skip cipher suites
	if pos+2 > len(data) {
		return "", errors.New("handshake message incomplete")
	}
	cipherSuitesLen := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2 + cipherSuitesLen

	// Skip compression methods
	if pos+1 > len(data) {
		return "", errors.New("handshake message incomplete")
	}
	compressionMethodsLen := int(data[pos])
	pos += 1 + compressionMethodsLen

	// Check if there are extensions
	if pos+2 > len(data) {
		return "", nil // No extensions, return empty hostname
	}
	extensionsLen := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2
	extensionsEnd := pos + extensionsLen

	// Loop through all extensions looking for SNI
	for pos+4 <= extensionsEnd {
		extType := binary.BigEndian.Uint16(data[pos : pos+2])
		extLen := int(binary.BigEndian.Uint16(data[pos+2 : pos+4]))
		pos += 4

		// SNI extension type is 0
		if extType == 0 && pos+extLen <= extensionsEnd {
			sniData := data[pos : pos+extLen]

			// Parse SNI
			if len(sniData) > 2 {
				sniListLen := int(binary.BigEndian.Uint16(sniData[0:2]))
				sniList := sniData[2 : 2+sniListLen]

				if len(sniList) > 3 {
					nameType := sniList[0]
					if nameType == 0 { // Hostname type
						nameLen := int(binary.BigEndian.Uint16(sniList[1:3]))
						if 3+nameLen <= len(sniList) {
							hostname := string(sniList[3 : 3+nameLen])
							return hostname, nil
						}
					}
				}
			}
		}

		pos += extLen
	}

	return "", nil // SNI extension not found, return empty hostname
}

// tunnelConnection forwards data bidirectionally between two connections
func tunnelConnection(client, target net.Conn) error {
	errCh := make(chan error, 2)

	// From client to target server
	go func() {
		_, err := io.Copy(target, client)
		target.(*net.TCPConn).CloseWrite()
		errCh <- err
	}()

	// From target server to client
	go func() {
		_, err := io.Copy(client, target)
		client.(*common.BufConn).CloseWrite()
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
