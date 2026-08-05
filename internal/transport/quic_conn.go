package transport

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

type QUICConn struct {
	stream *quic.Stream
	addr   net.Addr
}

func NewQUICConn(
	stream *quic.Stream,
	addr net.Addr,
) *QUICConn {

	return &QUICConn{
		stream: stream,
		addr:   addr,
	}
}

func (c *QUICConn) Read(b []byte) (int, error) {
	return c.stream.Read(b)
}

func (c *QUICConn) Write(b []byte) (int, error) {
	return c.stream.Write(b)
}

func (c *QUICConn) Close() error {
	return c.stream.Close()
}

func (c *QUICConn) LocalAddr() net.Addr {
	return c.addr
}

func (c *QUICConn) RemoteAddr() net.Addr {
	return c.addr
}

func (c *QUICConn) SetDeadline(t time.Time) error {
	c.stream.SetReadDeadline(t)
	c.stream.SetWriteDeadline(t)
	return nil
}

func (c *QUICConn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

func (c *QUICConn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}
