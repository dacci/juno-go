package socks

import (
	"net"
	"sync"

	"github.com/dacci/juno-go/util"
)

type SocksService struct {
	mu        sync.Mutex
	listeners []net.Listener
	dialer    net.Dialer
}

func NewServer(config map[string]interface{}) (server *SocksService, err error) {
	server = &SocksService{
		listeners: make([]net.Listener, 0, 16),
		dialer:    net.Dialer{},
	}

	if bind, ok := config["bind"].(string); ok {
		localAddr, err := net.ResolveTCPAddr("tcp", bind+":0")
		if err != nil {
			return nil, err
		}

		server.dialer.LocalAddr = localAddr
	}

	return
}

func (s *SocksService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var finalErr error
	for _, l := range s.listeners {
		if err := l.Close(); err != nil {
			finalErr = err
		}
	}

	return finalErr
}

func (s *SocksService) Serve(l net.Listener) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.listeners = append(s.listeners, l)

	go func() {
		defer l.Close()

		for {
			conn, err := l.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
					continue
				} else {
					return
				}
			}

			go s.handleConnection(conn)
		}
	}()

	return nil
}

func (service *SocksService) handleConnection(conn net.Conn) {
	stream := util.NewStream(conn)
	defer stream.Close()

	buffer, err := stream.Peek(1)
	if err != nil {
		return
	}

	switch buffer[0] {
	case 4:
		service.handleSocks4(stream)

	case 5:
		service.handleSocks5(stream)
	}
}
