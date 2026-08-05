package forward

import (
	"net"
	"testing"
	"time"
)

func TestMuxForwarderCleaner(t *testing.T) {

	a, b := net.Pipe()

	defer a.Close()
	defer b.Close()

	m := NewMuxForwarder(a)
	m.cleanInterval = 100 * time.Millisecond

	streamConn, _ := net.Pipe()

	m.mu.Lock()

	m.streams[1] = streamConn

	m.lastActive[1] = time.Now().Add(
		-10 * time.Minute,
	)

	m.mu.Unlock()

	m.StartCleaner(
		time.Minute,
	)

	time.Sleep(
		1500 * time.Millisecond,
	)

	m.mu.Lock()

	_, exists := m.streams[1]

	m.mu.Unlock()

	if exists {
		t.Fatal("idle stream was not removed")
	}
}
