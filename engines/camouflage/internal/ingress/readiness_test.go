package ingress

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/hashicorp/yamux"
)

func TestSessionReadinessRejectsBlackhole(t *testing.T) {
	client, peer := net.Pipe()
	defer peer.Close()
	// The transport accepts writes, but no mux replies ever arrive.
	go io.Copy(io.Discard, peer)
	session, err := yamux.Client(client, hubYamuxConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := verifySession(session, 50*time.Millisecond); err == nil {
		t.Fatal("blackholed session admitted")
	}
	if !session.IsClosed() {
		t.Fatal("rejected session was not closed")
	}
}

func TestSessionReadinessAcceptsWorkingMux(t *testing.T) {
	client, peer := net.Pipe()
	server, err := yamux.Server(peer, yamux.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	session, err := yamux.Client(client, hubYamuxConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := verifySession(session, time.Second); err != nil {
		t.Fatal(err)
	}
	if session.IsClosed() {
		t.Fatal("healthy session closed")
	}
}
