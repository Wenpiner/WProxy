package main

import (
	"WProxy/common"
	"WProxy/proxy/http"
	"WProxy/proxy/socks"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr  string `yaml:"listen_addr"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	Certificate struct {
		Key  string `yaml:"key"`
		Cert string `yaml:"cert"`
	} `yaml:"certificate"`
}

func main() {
	host := flag.String("host", "0.0.0.0", "host")
	port := flag.Int("port", 1080, "port")
	username := flag.String("username", "", "username")
	password := flag.String("password", "", "password")
	certificateKey := flag.String("certificate-key", "", "certificate_key")
	certificateCert := flag.String("certificate-cert", "", "certificate_cert")
	configFile := flag.String("c", "", "config file path")
	flag.Parse()

	var config Config
	if *configFile != "" {
		// Read config file
		data, err := os.ReadFile(*configFile)
		if err != nil {
			log.Fatalln("Failed to read config file:", err)
			return
		}
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			log.Fatalln("Failed to parse config file:", err)
			return
		}
		// Override config with command line arguments if provided
		if *host != "0.0.0.0" {
			config.ListenAddr = *host
		}
		if *port != 1080 {
			config.ListenAddr = fmt.Sprintf("%s:%d", config.ListenAddr, *port)
		}
		if *username != "" {
			config.Username = *username
		}
		if *password != "" {
			config.Password = *password
		}
	} else {
		// Use command line arguments
		config.ListenAddr = fmt.Sprintf("%s:%d", *host, *port)
		config.Username = *username
		config.Password = *password
		config.Certificate.Key = *certificateKey
		config.Certificate.Cert = *certificateCert
	}

	// Parse listen address
	hostStr, portStr, err := net.SplitHostPort(config.ListenAddr)
	if err != nil {
		log.Fatalln("Invalid listen address:", err)
		return
	}
	portNum, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalln("Invalid port:", err)
		return
	}

	listen, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(hostStr),
		Port: portNum,
	})
	if err != nil {
		log.Fatalln("Failed to start TCP proxy:", err)
		return
	}

	var urlinfo *url.Userinfo
	if config.Username != "" && config.Password != "" {
		fmt.Println("TCP proxy started successfully with credentials:", config.Username, config.Password)
		urlinfo = url.UserPassword(config.Username, config.Password)
	} else {
		fmt.Println("TCP proxy started successfully without credentials")
	}
	fmt.Println("Listening on", config.ListenAddr, " Port:", portNum, " User:", urlinfo.Username())

	// load cert key
	var cert tls.Certificate
	if config.Certificate.Key != "" && config.Certificate.Cert != "" {
		cert, err = tls.LoadX509KeyPair(config.Certificate.Cert, config.Certificate.Key)
		if err != nil {
			log.Fatalln("Failed to load certificate:", err)
			return
		}
		log.Println("Certificate loaded successfully")
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			log.Fatalln("Failed to parse certificate:", err)
			return
		}

		log.Printf("Certificate Info: Subject: %s, Issuer: %s, NotBefore: %s, NotAfter: %s \n",
			cert.Leaf.Subject, cert.Leaf.Issuer, cert.Leaf.NotBefore, cert.Leaf.NotAfter)

	}

	handlerTcp(*listen, urlinfo, &cert)
}

func handlerConn(tcp *net.TCPConn, userinfo *url.Userinfo, cert *tls.Certificate) {
	defer func(tcp *net.TCPConn) {
		err := tcp.Close()
		if err != nil {
			log.Println("TCP close error:", err)
		}
	}(tcp)
	// Switch Proxy
	conn := common.BufConn{
		Conn: tcp,
	}

	conn.Start()

	buf := [3]byte{}
	_, err := conn.Read(buf[:])
	if err != nil {
		log.Printf("Connection to %v failed to detect proxy protocol type: %v\n", tcp.RemoteAddr().String(), err)
		return
	}
	conn.Stop()
	if buf[0] == 0x05 {
		s := socks.Socks{
			Conn:     &conn,
			UserInfo: userinfo,
		}
		err = s.Handshake()
		if err != nil {
			log.Printf("SOCKS handshake error: %v\n", err)
			return
		}
	} else {

		tlsResult := buf[0] == 0x16

		err := http.HandleConn(&conn, userinfo, cert, tlsResult)
		if err != nil {
			log.Printf("HTTP handshake error: %v\n", err)
			return
		}
	}
}

func handlerTcp(listener net.TCPListener, userinfo *url.Userinfo, cert *tls.Certificate) {
	for {
		tcp, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("Error accepting TCP connection:", err)
			continue
		}
		log.Println("Accepted TCP connection from ", tcp.RemoteAddr().String())
		go handlerConn(tcp, userinfo, cert)
	}
}
