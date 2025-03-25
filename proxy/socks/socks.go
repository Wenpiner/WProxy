package socks

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
)

type Socks struct {
	net.Conn
	UserInfo *url.Userinfo
}

type TargetAddr struct {
	Name    string // 域名
	IP      net.IP // IP
	Port    int    // 端口
	Command byte   // 命令
	Type    byte   // 类型
}

func (s *Socks) Handshake() error {
	// 处理 socks 版本
	err := s.handlerVersion()
	if err != nil {
		return err
	}
	// 选择认证方式
	err = s.handlerAuth()
	if err != nil {
		return err
	}

	// 接收目标地址
	targetAddr, err := s.handlerTargetAddr()
	if err != nil {
		return err
	}
	// 处理IP请求转发
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

func (s *Socks) handlerVersion() error {
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

func (s *Socks) handlerAuth() error {
	// 读取认证方式数量
	methodsByte := make([]byte, 1)
	_, err := s.Conn.Read(methodsByte)
	if err != nil {
		return err
	}
	methodsLen := int(methodsByte[0])
	// 读取认证方式
	methods := make([]byte, methodsLen)
	_, err = s.Conn.Read(methods)
	if err != nil {
		return err
	}
	// 判断是否存在用户名密码的认证方式
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
		// 选择用户名密码认证方式
		_, err = s.Conn.Write([]byte{0x05, 0x02})
		if err != nil {
			return err
		}
		// 读取用户名密码认证结果
		authResult := make([]byte, 1)
		_, err = s.Conn.Read(authResult)
		if err != nil {
			return err
		}
		if authResult[0] != 0x01 {
			return errors.New("unknown socks auth version")
		}
		// 用户名长度
		usernameLen := make([]byte, 1)
		_, err = s.Conn.Read(usernameLen)
		if err != nil {
			return err
		}
		// 用户名
		username := make([]byte, int(usernameLen[0]))
		_, err = s.Conn.Read(username)
		if err != nil {
			return err
		}
		// 密码长度
		passwordLen := make([]byte, 1)
		_, err = s.Conn.Read(passwordLen)
		if err != nil {
			return err
		}
		// 密码
		password := make([]byte, int(passwordLen[0]))
		_, err = s.Conn.Read(password)
		if err != nil {
			return err
		}
		usernameStr := string(username)
		passwordStr := string(password)
		s2, _ := s.UserInfo.Password()
		// 验证用户名密码
		if s.UserInfo.Username() != usernameStr || s2 != passwordStr {
			// 验证失败
			_, err = s.Conn.Write([]byte{0x01, 0x01})
			if err != nil {
				return err
			}
			return errors.New("username/password auth failed")
		} else {
			// 验证成功
			_, err = s.Conn.Write([]byte{0x01, 0x00})
			if err != nil {
				return err
			}
		}
	} else {
		log.Println("No authentication")
		// 选择无需认证方式
		_, err = s.Conn.Write([]byte{0x05, 0x00})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Socks) handlerTargetAddr() (target TargetAddr, err error) {
	var ret []byte
	// 读取请求头
	header := [3]byte{}
	_, err = s.Conn.Read(header[:])
	if err != nil {
		return
	}
	// 判断请求头是否正确
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
	// 读取请求地址类型
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
		// 域名
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
	// 读取请求端口
	portByte := make([]byte, 2)
	_, err = s.Conn.Read(portByte)
	if err != nil {
		return
	}
	ret = append(ret, portByte...)
	// 使用内置函数转换成大端
	target.Port = int(binary.BigEndian.Uint16(portByte))

	// 回复客户端连接成功
	_, err = s.Conn.Write(ret)
	return
}
