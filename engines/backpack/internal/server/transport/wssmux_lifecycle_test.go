package transport

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"testing"
	"time"

	"github.com/backpack/backpack/config"
	"github.com/sirupsen/logrus"
)

func TestWSSMuxReleasesExclusiveListener(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := probe.Addr().String()
	probe.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := logrus.New()
	log.SetOutput(io.Discard)
	server := NewWSMuxServer(ctx, &WsMuxConfig{BindAddr: addr, Mode: config.WSSMUX}, log)
	server.Start()

	dialer := &net.Dialer{Timeout: 200 * time.Millisecond}
	var conn *tls.Conn
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{InsecureSkipVerify: true})
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("WSS mux did not start: %v", err)
	}
	defer conn.Close()

	if duplicate, err := net.Listen("tcp", addr); err == nil {
		duplicate.Close()
		t.Fatal("a second listener acquired the WSS mux address while it was running")
	}

	cancel()
	wait, stop := context.WithTimeout(context.Background(), 2*time.Second)
	defer stop()
	server.Wait(wait)
	replacement, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("WSS mux did not release its address after shutdown: %v", err)
	}
	defer replacement.Close()

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	var buf [1]byte
	if _, err := conn.Read(buf[:]); err == nil {
		t.Fatal("the old TLS connection stayed open")
	} else if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
		t.Fatal("shutdown left the old TLS connection waiting for data")
	}
}
