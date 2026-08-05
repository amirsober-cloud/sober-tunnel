package transport

import "net"

type TCPTransport struct {
	conn net.Conn
}

func NewTCP() *TCPTransport {
	return &TCPTransport{}
}

func (t *TCPTransport) Connect(address string) (net.Conn, error) {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	t.conn = conn

	return conn, nil
}

func (t *TCPTransport) Close() error {

	if t.conn == nil {
		return nil
	}

	return t.conn.Close()
}

func (t *TCPTransport) Name() string {
	return "tcp"
}
