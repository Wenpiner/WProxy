package main

import (
	"WProxy/common"
	"WProxy/proxy/http"
	"WProxy/proxy/socks"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
)

func main() {
	host := flag.String("host", "0.0.0.0", "host")
	port := flag.Int("port", 1080, "port")
	username := flag.String("username", "", "username")
	password := flag.String("password", "", "password")
	flag.Parse()
	listen, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(*host),
		Port: *port,
	})
	if err != nil {
		log.Fatalln("Failed to start TCP proxy:", err)
		return
	}
	var urlinfo *url.Userinfo
	if *username != "" && *password != "" {
		fmt.Println("TCP proxy started successfully with credentials:", *username, *password)
		urlinfo = url.UserPassword(*username, *password)
	} else {
		fmt.Println("TCP proxy started successfully without credentials")
	}

	handlerTcp(*listen, urlinfo)
	// TODO: Implement UDP proxy
}

func handlerConn(tcp *net.TCPConn, userinfo *url.Userinfo) {
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
		return
	}
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
		h := http.HttpConn{
			Conn:     &conn,
			UserInfo: userinfo,
		}
		err = h.Handshake()
		if err != nil {
			log.Printf("HTTP handshake error: %v\n", err)
			return
		}
	}
}

func handlerTcp(listener net.TCPListener, userinfo *url.Userinfo) {
	for {
		tcp, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("Error accepting TCP connection:", err)
			continue
		}
		go handlerConn(tcp, userinfo)
	}
}
