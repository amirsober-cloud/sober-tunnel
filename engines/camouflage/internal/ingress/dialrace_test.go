package ingress

import (
	"context"
	"errors"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/config"
	"github.com/hashicorp/yamux"
	"net"
	"testing"
	"time"
)

func TestBlockedPrimaryDoesNotDelayFallbackAndIsCancelled(t *testing.T) {
	c, p := net.Pipe()
	server, _ := yamux.Server(p, hubYamuxConfig())
	defer server.Close()
	session, _ := yamux.Client(c, hubYamuxConfig())
	defer session.Close()
	cancelled := make(chan struct{})
	started := time.Now()
	got, ep, err := raceEndpointAttempts(context.Background(), []config.Endpoint{{Target: "blocked"}, {Target: "working"}}, 15*time.Millisecond, func(ctx context.Context, ep config.Endpoint) (*yamux.Session, error) {
		if ep.Target == "blocked" {
			<-ctx.Done()
			close(cancelled)
			return nil, ctx.Err()
		}
		return session, nil
	})
	if err != nil || got != session || ep.Target != "working" {
		t.Fatalf("unexpected winner: %v %v", ep, err)
	}
	if time.Since(started) > time.Second {
		t.Fatal("fallback waited for blocked primary")
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("losing attempt leaked")
	}
}
func TestEndpointRaceReportsFailureAndHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := raceEndpointAttempts(ctx, []config.Endpoint{{Target: "dead"}}, time.Second, func(ctx context.Context, _ config.Endpoint) (*yamux.Session, error) { return nil, errors.New("failed") })
	if err == nil {
		t.Fatal("cancelled race succeeded")
	}
	_, _, err = raceEndpointAttempts(context.Background(), nil, time.Second, nil)
	if err == nil {
		t.Fatal("empty race succeeded")
	}
}
