package resumestream

import (
	"bytes"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testOpener(registry *Registry, serve func(net.Conn)) (OpenFunc, func()) {
	var mu sync.Mutex
	var wires []net.Conn
	return func() (net.Conn, error) {
			a, b := net.Pipe()
			mu.Lock()
			wires = append(wires, a)
			mu.Unlock()
			go func() {
				var kind [1]byte
				if _, err := io.ReadFull(b, kind[:]); err != nil {
					b.Close()
					return
				}
				if kind[0] != Type {
					b.Close()
					return
				}
				if registry.Accept(b, serve) != nil {
					b.Close()
				}
			}()
			return a, nil
		}, func() {
			mu.Lock()
			defer mu.Unlock()
			if len(wires) > 0 {
				wires[len(wires)-1].Close()
			}
		}
}
func TestResumePreservesBytesAfterPhysicalDrop(t *testing.T) {
	registry := NewRegistry()
	opener, drop := testOpener(registry, func(c net.Conn) { io.Copy(c, c) })
	conn, err := Dial(opener)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(8 * time.Second))
	payload := bytes.Repeat([]byte("resumable-payload-0123456789"), 40000)
	done := make(chan error, 1)
	go func() { _, err := conn.Write(payload); done <- err }()
	got := make([]byte, len(payload))
	if _, err := io.ReadFull(conn, got[:32768]); err != nil {
		t.Fatal(err)
	}
	drop()
	if _, err := io.ReadFull(conn, got[32768:]); err != nil {
		client := conn.(*Conn)
		client.mu.Lock()
		t.Logf("CLIENT base=%d end=%d recv=%d tx=%d rx=%d wire=%v generation=%d closed=%v", client.base, client.end, client.received, len(client.tx), len(client.rx), client.wire != nil, client.generation, client.closed)
		client.mu.Unlock()
		registry.mu.Lock()
		for _, server := range registry.streams {
			server.mu.Lock()
			t.Logf("SERVER base=%d end=%d recv=%d tx=%d rx=%d wire=%v generation=%d closed=%v", server.base, server.end, server.received, len(server.tx), len(server.rx), server.wire != nil, server.generation, server.closed)
			server.mu.Unlock()
		}
		registry.mu.Unlock()
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatal("duplicate, missing or corrupted bytes after resume")
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
func TestResumeReadDeadlineAndCloseReleaseWaiters(t *testing.T) {
	c := newConn()
	c.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	if _, err := c.Read(make([]byte, 1)); err == nil {
		t.Fatal("deadline ignored")
	}
	c.Close()
	if _, err := c.Write([]byte("x")); err == nil {
		t.Fatal("write after close")
	}
}
func TestUnknownResumeNeverCreatesNewBackend(t *testing.T) {
	registry := NewRegistry()
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	called := false
	done := make(chan error, 1)
	go func() { done <- registry.Accept(b, func(net.Conn) { called = true }) }()
	// Accept is called after the stream type byte has already been consumed.
	header := make([]byte, 26)
	header[0] = 1
	header[2] = 42
	if err := writeAll(a, header); err != nil {
		t.Fatal(err)
	}
	var reply [9]byte
	if _, err := io.ReadFull(a, reply[:]); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err == nil || called || reply[0] != 0 {
		t.Fatal("unknown resumed stream recreated backend")
	}
}

func TestHalfCloseKeepsReply(t *testing.T) {
	registry := NewRegistry()
	opener, _ := testOpener(registry, func(conn net.Conn) {
		c := conn.(*Conn)
		request, err := io.ReadAll(c)
		if err != nil {
			return
		}
		c.Write(append([]byte("reply:"), request...))
		c.CloseWrite()
		c.finish()
	})
	conn, err := Dial(opener)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write([]byte("request")); err != nil {
		t.Fatal(err)
	}
	conn.(*Conn).CloseWrite()
	reply, err := io.ReadAll(conn)
	if err != nil || string(reply) != "reply:request" {
		t.Fatalf("reply %q error %v", reply, err)
	}
}
func TestRejectsFutureAcknowledgement(t *testing.T) {
	c := newConn()
	c.mu.Lock()
	err := c.acknowledge(1)
	c.mu.Unlock()
	if err == nil {
		t.Fatal("peer acknowledged unsent bytes")
	}
}
func TestClosedResumedStreamReleasesRegistry(t *testing.T) {
	registry := NewRegistry()
	opener, _ := testOpener(registry, func(c net.Conn) { io.Copy(c, c) })
	conn, err := Dial(opener)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		registry.mu.Lock()
		count := len(registry.streams)
		registry.mu.Unlock()
		if count == 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	// An unclean wire close may retain exactly one bounded slot until the grace expires.
	registry.mu.Lock()
	count := len(registry.streams)
	registry.mu.Unlock()
	if count > 1 {
		t.Fatal("registry grew after a single stream")
	}
}

func TestReceiveWindowBackpressuresWithoutDeadlockingEcho(t *testing.T) {
	registry := NewRegistry()
	opener, _ := testOpener(registry, func(c net.Conn) { io.Copy(c, c) })
	conn, err := Dial(opener)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(4 * time.Second))
	payload := bytes.Repeat([]byte("window"), window)
	writeDone := make(chan error, 1)
	go func() { _, err := conn.Write(payload); writeDone <- err }()
	time.Sleep(10 * time.Millisecond)
	got := make([]byte, len(payload))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if err := <-writeDone; err != nil || !bytes.Equal(got, payload) {
		t.Fatal("credit accounting corrupted or stalled stream", err)
	}
}

func TestRepeatedLaneLossKeepsOneOrderedStream(t *testing.T) {
	registry := NewRegistry()
	var opens atomic.Int32
	opener, drop := testOpener(registry, func(c net.Conn) { opens.Add(1); io.Copy(c, c) })
	conn, err := Dial(opener)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	payload := bytes.Repeat([]byte("repeated-loss"), 180000)
	written := make(chan error, 1)
	go func() { _, err := conn.Write(payload); written <- err }()
	got := make([]byte, len(payload))
	offset := 0
	for i := 0; i < 3; i++ {
		if _, err := io.ReadFull(conn, got[offset:offset+65536]); err != nil {
			t.Fatal(err)
		}
		offset += 65536
		drop()
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := io.ReadFull(conn, got[offset:]); err != nil {
		t.Fatal(err)
	}
	if err := <-written; err != nil || !bytes.Equal(got, payload) || opens.Load() != 1 {
		t.Fatal("resume duplicated backend or corrupted bytes", err)
	}
}

func TestReplayOfBufferedBytesAfterFIN(t *testing.T) {
	c := newConn()
	defer c.Close()
	a, b := net.Pipe()
	defer b.Close()
	c.wire = a
	c.rx = []byte("reply")
	c.received = 5
	c.rxFIN = true
	go c.readWire(a)
	if err := writeFrame(b, dataFrame, 0, []byte("reply")); err != nil {
		t.Fatal(err)
	}
	if err := writeFrame(b, finFrame, 5, nil); err != nil {
		t.Fatal(err)
	}
	c.SetReadDeadline(time.Now().Add(time.Second))
	got, err := io.ReadAll(c)
	if err != nil || string(got) != "reply" {
		t.Fatalf("reply after FIN replay: %q %v", got, err)
	}
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		t.Fatal("duplicate unconsumed bytes after FIN closed the logical stream")
	}
}
