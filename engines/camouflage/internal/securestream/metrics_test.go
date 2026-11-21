package securestream

import (
	"io"
	"net"
	"testing"
)

func TestAuthenticatedTLSMeterCountsBothDirections(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	c := MeterAuthenticatedConn(a)
	beforeSent, beforeReceived := sentBytes.Load(), receivedBytes.Load()
	done := make(chan error, 1)
	go func() {
		_, err := b.Write([]byte("request"))
		if err == nil {
			_, err = io.ReadFull(b, make([]byte, 5))
		}
		done <- err
	}()
	if _, err := io.ReadFull(c, make([]byte, 7)); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Write([]byte("reply")); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if sentBytes.Load()-beforeSent != 5 || receivedBytes.Load()-beforeReceived != 7 {
		t.Fatal("TLS plaintext missing or double counted")
	}
}
