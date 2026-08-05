package forward

import (
	"net"
	"testing"
	"time"
)

func TestMuxForwarderConcurrentStreams(t *testing.T) {

	a, b := net.Pipe()

	defer a.Close()
	defer b.Close()

	m := NewMuxForwarder(a)

	m.cleanInterval = 50 * time.Millisecond

	go m.StartCleaner(
		50 * time.Millisecond,
	)

	for i := 0; i < 100; i++ {

		conn1, conn2 := net.Pipe()

		go func() {

			m.mu.Lock()

			id := uint32(time.Now().UnixNano())

			m.streams[id] = conn1
			m.lastActive[id] = time.Now()

			m.mu.Unlock()

			time.Sleep(
				time.Millisecond * 20,
			)

			conn1.Close()

			conn2.Close()

		}()
	}

	time.Sleep(
		500 * time.Millisecond,
	)

	m.mu.Lock()

	count := len(m.streams)

	m.mu.Unlock()

	if count > 100 {
		t.Fatalf(
			"unexpected stream count: %d",
			count,
		)
	}
}
