package socks

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
)

type Socks struct {
	net.Conn
	UserInfo *url.Userinfo
}

type TargetAddr struct {
	Name    string // Domain name
	IP      net.IP // IP address
	Port    int    // Port
	Command byte   // Command
	Type    byte   // Type
}

type Request struct {
	Version byte   // Version
	Name    string // Domain name
	IP      net.IP // IP address
	Port    int    // Port
	Command byte   // Command
	Type    byte   // Type
}

func (s *Socks) Handshake() error {
	// Handle socks version
	err := s.handleVersion()
	if err != nil {
		return err
	}
	// Select authentication method
	err = s.selectAuthMethod()
	if err != nil {
		return err
	}

	// Receive target address
	targetAddr, err := s.receiveTargetAddress()
	if err != nil {
		return err
	}
	// Handle IP request forwarding
	if targetAddr.Command == 0x01 {
		dial, err := net.Dial("tcp", fmt.Sprintf("%s:%d", targetAddr.IP.String(), targetAddr.Port))
		if err != nil {
			return err
		}
		go io.Copy(s.Conn, dial)
		io.Copy(dial, s.Conn)
		return nil
	}
	return nil
}

// Handle socks version
func (s *Socks) handleVersion() error {
	verByte := [1]byte{}
	_, err := s.Conn.Read(verByte[:])
	if err != nil {
		return err
	}
	if verByte[0] != 0x05 {
		return errors.New(fmt.Sprintf("invalid socks version %d", verByte[0]))
	}
	return nil
}

// Select authentication method
func (s *Socks) selectAuthMethod() error {
	// Read number of authentication methods
	methodsByte := make([]byte, 1)
	_, err := s.Conn.Read(methodsByte)
	if err != nil {
		return err
	}
	methodsLen := int(methodsByte[0])
	// Read authentication methods
	methods := make([]byte, methodsLen)
	_, err = s.Conn.Read(methods)
	if err != nil {
		return err
	}
	// Check if username/password authentication method exists
	if s.UserInfo != nil {
		auth := false
		for _, method := range methods {
			if method == 0x02 {
				auth = true
			}
		}
		if !auth {
			return errors.New("username/password auth not support")
		}
		// Select username/password authentication method
		_, err = s.Conn.Write([]byte{0x05, 0x02})
		if err != nil {
			return err
		}
		// Read username/password authentication result
		authResult := make([]byte, 1)
		_, err = s.Conn.Read(authResult)
		if err != nil {
			return err
		}
		if authResult[0] != 0x01 {
			return errors.New("unknown socks auth version")
		}
		// Read username length
		usernameLen := make([]byte, 1)
		_, err = s.Conn.Read(usernameLen)
		if err != nil {
			return err
		}
		// Read username
		username := make([]byte, int(usernameLen[0]))
		_, err = s.Conn.Read(username)
		if err != nil {
			return err
		}
		// Read password length
		passwordLen := make([]byte, 1)
		_, err = s.Conn.Read(passwordLen)
		if err != nil {
			return err
		}
		// Read password
		password := make([]byte, int(passwordLen[0]))
		_, err = s.Conn.Read(password)
		if err != nil {
			return err
		}
		usernameStr := string(username)
		passwordStr := string(password)
		s2, _ := s.UserInfo.Password()
		// Verify username/password
		if s.UserInfo.Username() != usernameStr || s2 != passwordStr {
			// Verify failed
			_, err = s.Conn.Write([]byte{0x01, 0x01})
			if err != nil {
				return err
			}
			return errors.New("username/password auth failed")
		} else {
			// Verify succeeded
			_, err = s.Conn.Write([]byte{0x01, 0x00})
			if err != nil {
				return err
			}
		}
	} else {
		// Select no authentication method
		_, err = s.Conn.Write([]byte{0x05, 0x00})
		if err != nil {
			return err
		}
	}
	return nil
}

// Receive target address
func (s *Socks) receiveTargetAddress() (target TargetAddr, err error) {
	var ret []byte
	// Read request header
	header := [3]byte{}
	_, err = s.Conn.Read(header[:])
	if err != nil {
		return
	}
	// Check if request header is correct
	if header[0] != 0x05 {
		err = errors.New(fmt.Sprintf("unknown socks version:%s", hex.EncodeToString(header[:])))
		return
	}
	ret = append(ret, 0x05)
	if header[1] != 0x01 {
		err = errors.New("unknown socks cmd")
		return
	}
	target.Command = header[1]
	ret = append(ret, 0x00)
	if header[2] != 0x00 {
		err = errors.New("unknown socks rsv")
		return
	}
	ret = append(ret, 0x00)
	// Read request address type
	addrTypeByte := make([]byte, 1)
	_, err = s.Conn.Read(addrTypeByte)
	if err != nil {
		return
	}
	ret = append(ret, addrTypeByte...)
	target.Type = addrTypeByte[0]
	switch target.Type {
	case 0x01:
		// IPv4
		addrByte := make([]byte, 4)
		_, err = s.Conn.Read(addrByte)
		if err != nil {
			return
		}
		target.IP = net.IPv4(addrByte[0], addrByte[1], addrByte[2], addrByte[3])
		ret = append(ret, addrByte...)
	case 0x03:
		// Domain name
		addrLenByte := make([]byte, 1)
		_, err = s.Conn.Read(addrLenByte)
		if err != nil {
			return
		}
		ret = append(ret, addrLenByte[0])
		addrLen := int(addrLenByte[0])
		addrByte := make([]byte, addrLen)
		_, err = s.Conn.Read(addrByte)
		if err != nil {
			return
		}
		ret = append(ret, addrByte...)
		target.Name = string(addrByte)
	case 0x04:
		// IPv6
		addrByte := make([]byte, 16)
		_, err = s.Conn.Read(addrByte)
		if err != nil {
			return
		}
		target.IP = net.IP(addrByte)
		ret = append(ret, addrByte...)
	default:
		err = errors.New("unknown socks addr type")
		return
	}
	// Read request port
	portByte := make([]byte, 2)
	_, err = s.Conn.Read(portByte)
	if err != nil {
		return
	}
	ret = append(ret, portByte...)
	// Use built-in function to convert to big endian
	target.Port = int(binary.BigEndian.Uint16(portByte))

	// Reply client connection success
	_, err = s.Conn.Write(ret)
	return
}
