package network

import (
	"bufio"
	"net"
)

// upgradeBoundaryConn retains bytes read ahead while parsing an HTTP upgrade.
// Mux carriers switch from the WebSocket handshake to raw SMUX via NetConn;
// Gorilla's private reader must not take the first SMUX frame with the headers.
// The inner reader keeps read-ahead bytes accessible after that switch.
type upgradeBoundaryConn struct {
	net.Conn
	reader      *bufio.Reader
	tail        uint32
	headersDone bool
}

func newUpgradeBoundaryConn(conn net.Conn) net.Conn {
	return &upgradeBoundaryConn{Conn: conn, reader: bufio.NewReader(conn)}
}

func (c *upgradeBoundaryConn) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if c.headersDone {
		return c.reader.Read(p)
	}
	n, err := c.reader.Read(p[:1])
	if n > 0 {
		c.tail = c.tail<<8 | uint32(p[0])
		c.headersDone = c.tail == 0x0d0a0d0a
	}
	return n, err
}
