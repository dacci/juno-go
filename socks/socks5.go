package socks

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"syscall"

	"github.com/dacci/juno-go/util"
)

const (
	socks5Version = 5

	socks5AuthNone        = 0x00
	socks5AuthGSSAPI      = 0x01
	socks5AuthPassword    = 0x02
	socks5AuthUnsupported = 0xFF

	socks5Connect   = 1
	socks5Bind      = 2
	socks5Associate = 3

	socks5IPv4       = 1
	socks5DomainName = 3
	socks5IPv6       = 4

	socks5Succeeded       = 0
	socks5GeneralError    = 1
	socks5ConnProhibited  = 2
	socks5NetUnreachable  = 3
	socks5HostUnreachable = 4
	socks5ConnRefused     = 5
	socks5TTLExpired      = 6
	socks5InvalidCommand  = 7
	socks5IllegalAddress  = 8
)

type socks5Reply []byte

func socks5NewReply() (reply socks5Reply) {
	reply = make([]byte, 10, 4+net.IPv6len+2)
	reply[0] = socks5Version
	reply[3] = socks5IPv4
	return
}

func (reply *socks5Reply) setAddress(hostport string) (err error) {
	host, portStr, err := net.SplitHostPort(hostport)
	if err != nil {
		return
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return
	}

	ip := net.ParseIP(host)
	if ipv4 := ip.To4(); ipv4 != nil {
		(*reply)[3] = socks5IPv4
		*reply = append((*reply)[0:4], ipv4...)
	} else {
		(*reply)[3] = socks5IPv6
		*reply = append((*reply)[0:4], ip...)
	}

	*reply = append(*reply, byte(port>>8), byte(port&0xFF))

	return
}

func socks5ReadMethods(stream *util.Stream) (methods []byte, err error) {
	header := make([]byte, 2)
	if _, err = io.ReadFull(stream, header); err != nil {
		return
	}

	if header[0] != 5 {
		err = fmt.Errorf("invalid version number: %d", header[0])
		return
	}

	if header[1] < 1 {
		err = fmt.Errorf("invalid number of methods: %d", header[1])
		return
	}

	methods = make([]byte, header[1])
	if _, err = io.ReadFull(stream, methods); err != nil {
		return
	}

	return
}

func socks5FindMethod(wanted []byte, available ...byte) byte {
	for _, a := range available {
		for _, w := range wanted {
			if w == a {
				return a
			}
		}
	}

	return socks5AuthUnsupported
}

type socks5Request struct {
	version byte
	command byte
	host    string
	port    uint16
}

func socks5ReadRequest(stream *util.Stream) (request *socks5Request, err error) {
	header := make([]byte, 4)
	if _, err = io.ReadFull(stream, header); err != nil {
		return
	}

	if header[0] != 5 {
		err = fmt.Errorf("invalid version number: %d", header[0])
		return
	}

	request = &socks5Request{
		version: header[0],
		command: header[1],
	}

	switch header[3] {
	case socks5IPv4:
		ip := make([]byte, net.IPv4len)
		if _, err = io.ReadFull(stream, ip); err != nil {
			return
		}

		request.host = net.IP(ip).String()

	case socks5DomainName:
		var n byte
		if n, err = stream.ReadByte(); err != nil {
			return
		}

		buffer := make([]byte, n)
		if _, err = io.ReadFull(stream, buffer); err != nil {
			return
		}

		request.host = string(buffer)

	case socks5IPv6:
		ip := make([]byte, net.IPv6len)
		if _, err = io.ReadFull(stream, ip); err != nil {
			return
		}

		request.host = net.IP(ip).String()
	}

	buffer := make([]byte, 2)
	if _, err = io.ReadFull(stream, buffer); err != nil {
		return
	}

	request.port = uint16(buffer[0])<<8 | uint16(buffer[1])

	return
}

func socks5MapErrToCode(err error) (code byte) {
	switch v := err.(type) {
	case *net.OpError:
		code = socks5MapErrToCode(v.Err)

	case *os.SyscallError:
		code = socks5MapErrToCode(v.Err)

	case *net.DNSError:
		code = socks5HostUnreachable

	case syscall.Errno:
		switch v {
		case syscall.ECONNREFUSED:
			code = socks5ConnRefused

		case syscall.ENETUNREACH:
			code = socks5NetUnreachable

		case syscall.ETIMEDOUT:
			code = socks5TTLExpired
		}

	default:
		code = socks5GeneralError
	}

	return
}

func (service *SocksService) handleSocks5(stream *util.Stream) {
	methods, err := socks5ReadMethods(stream)
	if err != nil {
		return
	}

	methodReply := make([]byte, 2)
	methodReply[0] = socks5Version
	methodReply[1] = socks5FindMethod(methods, socks5AuthNone)
	if _, err := stream.Write(methodReply); err != nil {
		return
	}
	if methodReply[1] == socks5AuthUnsupported {
		return
	}

	request, err := socks5ReadRequest(stream)
	if err != nil {
		return
	}

	reply := socks5NewReply()

	if request.command != socks5Connect {
		reply[1] = socks5InvalidCommand
		stream.Write(reply)
		return
	}

	if request.host == "" {
		reply[1] = socks5IllegalAddress
		stream.Write(reply)
		return
	}

	address := net.JoinHostPort(request.host, fmt.Sprint(request.port))

	dest, err := service.dialer.Dial("tcp", address)
	if err != nil {
		reply[1] = socks5MapErrToCode(err)
		stream.Write(reply)
		return
	}
	defer dest.Close()

	reply[1] = socks5Succeeded
	reply.setAddress(dest.LocalAddr().String())
	if _, err = stream.Write(reply); err != nil {
		return
	}

	go func() {
		defer stream.Close()
		defer dest.Close()
		io.Copy(stream, dest)
	}()

	io.Copy(dest, stream)
}
