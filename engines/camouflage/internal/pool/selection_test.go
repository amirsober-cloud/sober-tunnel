package pool

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/hashicorp/yamux"
)

func selectionSession(t *testing.T) *YamuxSession {
	t.Helper()
	a, b := net.Pipe()
	cfg := yamux.DefaultConfig()
	cfg.EnableKeepAlive = false
	server, err := yamux.Server(b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	client, err := yamux.Client(a, cfg)
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close(); server.Close() })
	go func() {
		for {
			s, err := server.AcceptStream()
			if err != nil {
				return
			}
			go func() { defer s.Close(); io.Copy(io.Discard, s) }()
		}
	}()
	return NewYamuxSession(client, 100, 0, "ssh", NewLifecyclePolicy("selection-test"))
}

func waitEmpty(t *testing.T, s *YamuxSession) {
	t.Helper()
	until := time.Now().Add(time.Second)
	for s.ActiveStreams() != 0 {
		if time.Now().After(until) {
			t.Fatal("stream did not finish")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestSequentialShortConnectionsUseAllEqualLoadPipes(t *testing.T) {
	p := &NodePool{sessions: []*YamuxSession{selectionSession(t), selectionSession(t), selectionSession(t)}}
	counts := map[*YamuxSession]int{}
	for range 30 {
		stream, err := p.getStreamLeastLoaded()
		if err != nil {
			t.Fatal(err)
		}
		parent := stream.(*monitoredStream).parent
		counts[parent]++
		stream.Close()
		waitEmpty(t, parent)
	}
	for _, s := range p.sessions {
		if counts[s] != 10 {
			t.Fatalf("short requests concentrated on a pipe: %v", counts)
		}
	}
}

func TestPoolPrefersLowerLoadAndSkipsDrainingAndClosed(t *testing.T) {
	busy, draining, closed, ready := selectionSession(t), selectionSession(t), selectionSession(t), selectionSession(t)
	held, err := busy.OpenStream()
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	draining.SetDraining()
	closed.Close()
	p := &NodePool{sessions: []*YamuxSession{busy, draining, closed, ready}}
	stream, err := p.getStreamLeastLoaded()
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if stream.(*monitoredStream).parent != ready {
		t.Fatal("selected a busy, draining or closed pipe")
	}
}

func TestStalledStreamOpenDoesNotBlockPoolRepair(t *testing.T) {
	a, b := net.Pipe()
	defer b.Close()
	cfg := yamux.DefaultConfig()
	cfg.EnableKeepAlive = false
	cfg.ConnectionWriteTimeout = 3 * time.Second
	cfg.StreamOpenTimeout = 0
	client, err := yamux.Client(a, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	s := NewYamuxSession(client, 100, 0, "ssh", NewLifecyclePolicy("selection-test"))
	p := &NodePool{sessions: []*YamuxSession{s}}
	opened := make(chan struct{})
	go func() {
		c, _ := p.getStreamLeastLoaded()
		if c != nil {
			c.Close()
		}
		close(opened)
	}()
	until := time.Now().Add(time.Second)
	for client.NumStreams() == 0 {
		if time.Now().After(until) {
			t.Fatal("open did not reach stalled carrier")
		}
		time.Sleep(time.Millisecond)
	}
	repaired := make(chan struct{})
	go func() { p.mu.Lock(); p.mu.Unlock(); close(repaired) }()
	select {
	case <-repaired:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stream open held the pool lock and blocked maintenance")
	}
	if s.pendingOpens.Load() != 1 {
		t.Fatal("opening stream lost its admission reservation")
	}
	client.Close()
	select {
	case <-opened:
	case <-time.After(time.Second):
		t.Fatal("closing the carrier did not release admission")
	}
	if s.pendingOpens.Load() != 0 {
		t.Fatal("failed admission leaked a reservation")
	}
}
