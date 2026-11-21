package ingress

import (
	"github.com/hashicorp/yamux"
	"io"
	"net"
	"testing"
	"time"
)

func TestEstablishedBlackholeFeedsEndpointCooldown(t *testing.T) {
	d := threeEndpointDialer()
	ep := d.node.Endpoints[0]
	for i := 0; i < epFailThreshold; i++ {
		client, peer := net.Pipe()
		go io.Copy(io.Discard, peer)
		session, err := yamux.Client(client, hubYamuxConfig())
		if err != nil {
			t.Fatal(err)
		}
		d.monitorSessionWithTiming(session, ep, func(int) time.Duration { return time.Millisecond }, 20*time.Millisecond)
		if !session.IsClosed() {
			t.Fatal("dead session retained")
		}
		peer.Close()
	}
	d.mu.Lock()
	cooled := d.health[ep.Target].cooldownUntil.After(time.Now())
	d.mu.Unlock()
	if !cooled {
		t.Fatal("post-handshake failures did not cool path")
	}
	for _, candidate := range d.attemptOrder() {
		if candidate.Target == ep.Target {
			t.Fatal("cooled endpoint retried despite eligible alternatives")
		}
	}
}
func TestHealthyRuntimeRecoversAfterTwoChecks(t *testing.T) {
	d := threeEndpointDialer()
	ep := d.node.Endpoints[0]
	d.recordFailure(ep.Target)
	d.recordFailure(ep.Target)
	client, peer := net.Pipe()
	server, err := yamux.Server(peer, hubYamuxConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	session, err := yamux.Client(client, hubYamuxConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	d.monitorSessionWithTiming(session, ep, func(successes int) time.Duration {
		if successes == 2 {
			session.Close()
		}
		return time.Millisecond
	}, time.Second)
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.health[ep.Target].fails != 0 {
		t.Fatal("healthy path did not recover")
	}
}
