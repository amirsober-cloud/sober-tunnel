package resumestream

import (
	"io"
	"net"
	"time"
)

// Relay preserves TCP half-close: finishing a request must not discard its reply.
func Relay(a, b net.Conn) {
	done := make(chan error, 2)
	copyHalf := func(dst, src net.Conn) {
		_, err := io.Copy(dst, src)
		if err == nil {
			if half, ok := dst.(interface{ CloseWrite() error }); ok {
				err = half.CloseWrite()
			}
		}
		done <- err
	}
	go copyHalf(a, b)
	go copyHalf(b, a)
	if err := <-done; err != nil {
		a.Close()
		b.Close()
	}
	if err := <-done; err != nil {
		a.Close()
		b.Close()
	}
	for _, conn := range []net.Conn{a, b} {
		if c, ok := conn.(*Conn); ok {
			c.finish()
		}
	}
}
func (c *Conn) finish() {
	c.CloseWrite()
	deadline := time.Now().Add(grace)
	for {
		c.mu.Lock()
		ready := c.closed || (len(c.tx) == 0 && c.finDelivered)
		ch := c.changed
		c.mu.Unlock()
		if ready {
			return
		}
		if wait(ch, deadline) != nil {
			return
		}
	}
}
