package util

import (
	"bufio"
	"net"
	"time"
)

type Stream struct {
	*bufio.Reader
	conn net.Conn
}

func (s *Stream) Write(b []byte) (int, error) {
	return s.conn.Write(b)
}

func (s *Stream) Close() error {
	return s.conn.Close()
}

func (s *Stream) SetDeadline(t time.Time) error {
	return s.conn.SetDeadline(t)
}

func (s *Stream) SetReadDeadline(t time.Time) error {
	return s.conn.SetReadDeadline(t)
}

func (s *Stream) SetWriteDeadline(t time.Time) error {
	return s.conn.SetWriteDeadline(t)
}

func NewStream(conn net.Conn) *Stream {
	return &Stream{
		bufio.NewReader(conn),
		conn,
	}
}
