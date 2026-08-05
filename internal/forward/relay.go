package forward

import (
	"io"
	"net"
)

func Relay(a net.Conn, b net.Conn) {

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(a, b)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(b, a)
		done <- struct{}{}
	}()

	<-done

	a.Close()
	b.Close()
}
