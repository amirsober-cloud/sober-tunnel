package connection

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
	"github.com/amirsober-cloud/sober-tunnel/internal/mux"
	"github.com/amirsober-cloud/sober-tunnel/internal/transport"
)

func TestManagerMultiplexedRoundTrip(t *testing.T) {
	logger.Init()
	target, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	go func() {
		for {
			conn, err := target.Accept()
			if err != nil {
				return
			}
			go func() { defer conn.Close(); io.Copy(conn, conn) }()
		}
	}()
	tunnel, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer tunnel.Close()
	go func() {
		conn, err := tunnel.Accept()
		if err == nil {
			mux.NewHandler(mux.NewManager()).Handle(conn)
		}
	}()
	host, port, _ := net.SplitHostPort(tunnel.Addr().String())
	p, _ := strconv.Atoi(port)
	manager := New(transport.NewTCP())
	manager.Start(&config.Config{RemoteHost: host, RemotePort: p, MinConnections: 1, MaxConnections: 1})
	defer manager.Stop()
	session := manager.FreeSession()
	if session == nil {
		t.Fatal("connected tunnel has no available session")
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			local, peer := net.Pipe()
			defer peer.Close()
			peer.SetDeadline(time.Now().Add(15 * time.Second))
			if !session.AcquireStream() {
				t.Error("stream unavailable")
				local.Close()
				return
			}
			done := make(chan struct{})
			go func() {
				defer close(done)
				defer session.ReleaseStream()
				session.Forwarder.HandleLocal(local, target.Addr().String())
			}()
			payload := bytes.Repeat([]byte(fmt.Sprintf("stream-%02d/", i)), 8192)
			writeDone := make(chan error, 1)
			go func() { _, err := peer.Write(payload); writeDone <- err }()
			got := make([]byte, len(payload))
			if _, err := io.ReadFull(peer, got); err != nil {
				t.Errorf("stream %d: %v", i, err)
			}
			if err := <-writeDone; err != nil {
				t.Error(err)
			}
			if !bytes.Equal(got, payload) {
				t.Errorf("stream %d payload corrupted", i)
			}
			peer.Close()
			select {
			case <-done:
			case <-time.After(15 * time.Second):
				t.Error("stream did not release")
			}
		}(i)
	}
	wg.Wait()
	if session.ActiveStreams() != 0 {
		t.Fatal("stream reservations leaked")
	}
	if manager.pool.Size() != 1 {
		t.Fatal("shared tunnel disconnected during traffic")
	}
	// A refused destination must close only its local stream, not hang it.
	target.Close()
	local, peer := net.Pipe()
	peer.SetReadDeadline(time.Now().Add(5 * time.Second))
	go session.Forwarder.HandleLocal(local, target.Addr().String())
	var one [1]byte
	_, err = peer.Read(one[:])
	peer.Close()
	if err != io.EOF {
		t.Fatalf("refused target did not close stream: %v", err)
	}
	if manager.pool.Size() != 1 {
		t.Fatal("refused target disconnected shared tunnel")
	}
	manager.pool.Remove(session.Conn)
	if manager.FreeSession() != nil {
		t.Fatal("closed session remains available")
	}
}
