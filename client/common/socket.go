package common

import (
	"bufio"
	"errors"
	"io"
	"net"
)

type Socket struct {
	conn   net.Conn
	buffer *bufio.Reader
}

func NewSocket(conn net.Conn) *Socket {
	return &Socket{
		conn:   conn,
		buffer: bufio.NewReaderSize(conn, 2048),
	}
}

func ConnectTCP(ServerAddress string) (*Socket, error) {
	conn, err := net.Dial("tcp", ServerAddress)
	if err != nil {
		return nil, err
	}
	return NewSocket(conn), nil
}

// Reads N bytes from socket and returns its content in a buffer on success
func (s *Socket) ReadBytes(n int) ([]byte, error) {
	data := make([]byte, n)
	_, err := io.ReadFull(s.buffer, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Socket) SendBytes(data []byte) error {
	size := len(data)
	bytesSent := 0
	for bytesSent < size {
		n, err := s.conn.Write(data[bytesSent:])
		if err != nil {
			return err
		}
		if n == 0 {
			return errors.New("0 bytes sent. Write attempt cancelled")
		}
		bytesSent += n
	}
	return nil
}

func (s *Socket) ReadLine() (string, error) {
	line, err := s.buffer.ReadString('\n')
	if err != nil {
		return "", err
	}

	return line, nil
}

func (s *Socket) Disconnect() error {
	err := s.conn.Close()
	s.conn = nil
	return err
}
