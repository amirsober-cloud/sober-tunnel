package transport

import (
	"net"
	"time"
)

type TCPTransport struct {
}

func NewTCP() *TCPTransport {
	return &TCPTransport{}
}

func (t *TCPTransport) Connect(address string) (net.Conn, error) {

	dialer := net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	conn, err := dialer.Dial(
		"tcp",
		address,
	)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (t *TCPTransport) Close() error {
	return nil
}

func (t *TCPTransport) Name() string {
	return "tcp"
}
