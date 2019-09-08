package socks

import (
	"fmt"
	"io"
	"net"

	"github.com/dacci/juno-go/util"
)

const (
	socks4Version      = 4
	socks4Connect      = 1
	socks4Bind         = 2
	socks4Granted      = 90
	socks4Rejected     = 91
	socks4Failed       = 92
	socks4Unauthorized = 93
)

type socks4Request struct {
	version byte
	code    byte
	user    string
	host    string
	port    uint16
}

func socks4ReadString(stream *util.Stream) (s string, err error) {
	if slice, err := stream.ReadSlice(0); err == nil {
		s = string(slice[:len(slice)-1])
	}

	return
}

func socks4ReadRequest(stream *util.Stream) (request *socks4Request, err error) {
	buffer := make([]byte, 8)
	if _, err = io.ReadFull(stream, buffer); err != nil {
		return
	}

	if buffer[0] != 4 {
		err = fmt.Errorf("invalid version number: %d", buffer[0])
		return
	}

	user, err := socks4ReadString(stream)
	if err != nil {
		return
	}

	request = &socks4Request{
		version: buffer[0],
		code:    buffer[1],
		user:    user,
		port:    uint16(buffer[2])<<8 | uint16(buffer[3]),
	}

	if buffer[4] == 0 && buffer[5] == 0 && buffer[6] == 0 && buffer[7] != 0 {
		request.host, err = socks4ReadString(stream)
		if err != nil {
			return
		}
	} else {
		request.host = net.IP(buffer[4:8]).String()
	}

	return
}

func (service *SocksService) handleSocks4(stream *util.Stream) {
	request, err := socks4ReadRequest(stream)
	if err != nil {
		return
	}

	reply := make([]byte, 8)

	if request.code != socks4Connect {
		reply[1] = socks4Rejected
		stream.Write(reply)
		return
	}

	address := net.JoinHostPort(request.host, fmt.Sprint(request.port))

	dest, err := service.dialer.Dial("tcp", address)
	if err != nil {
		reply[1] = socks4Rejected
		stream.Write(reply)
		return
	}
	defer dest.Close()

	reply[1] = socks4Granted
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
