package ingress

import (
	"bytes"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/config"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/internal/mimic"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/internal/resumestream"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/internal/securestream"
	"github.com/hashicorp/yamux"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestContinuationAcrossAuthenticatedSSHSessionLoss(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	const token = "resume-authenticated-test-secret"
	registry := resumestream.NewRegistry()
	var backendOpens atomic.Int32
	var serversMu sync.Mutex
	var servers []*yamux.Session
	defer func() {
		serversMu.Lock()
		defer serversMu.Unlock()
		for _, s := range servers {
			s.Close()
		}
	}()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				server := &mimic.SSHMimic{Token: token, Filter: securestream.NewReplayFilter(0)}
				authenticated, _, err := server.Accept(conn)
				if err != nil {
					conn.Close()
					return
				}
				session, err := yamux.Server(authenticated, hubYamuxConfig())
				if err != nil {
					authenticated.Close()
					return
				}
				serversMu.Lock()
				servers = append(servers, session)
				serversMu.Unlock()
				for {
					stream, err := session.AcceptStream()
					if err != nil {
						return
					}
					go func() {
						var kind [1]byte
						if _, err := io.ReadFull(stream, kind[:]); err != nil || kind[0] != resumestream.Type {
							stream.Close()
							return
						}
						if registry.Accept(stream, func(c net.Conn) { backendOpens.Add(1); io.Copy(c, c) }) != nil {
							stream.Close()
						}
					}()
				}
			}()
		}
	}()
	var mu sync.Mutex
	var current *yamux.Session
	defer func() {
		mu.Lock()
		defer mu.Unlock()
		if current != nil {
			current.Close()
		}
	}()
	opener := func() (net.Conn, error) {
		mu.Lock()
		defer mu.Unlock()
		if current == nil || current.IsClosed() {
			fresh, dialErr := DialEndpoint(config.Endpoint{Target: listener.Addr().String(), Mimic: "ssh"}, token, hubYamuxConfig())
			if dialErr != nil {
				return nil, dialErr
			}
			current = fresh
		}
		return current.Open()
	}
	conn, err := resumestream.Dial(opener)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(8 * time.Second))
	payload := bytes.Repeat([]byte("authenticated-resume"), 60000)
	done := make(chan error, 1)
	go func() { _, e := conn.Write(payload); done <- e }()
	got := make([]byte, len(payload))
	if _, err := io.ReadFull(conn, got[:32768]); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	current.Close()
	mu.Unlock()
	if _, err := io.ReadFull(conn, got[32768:]); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatal("payload changed across encrypted session loss")
	}
	if backendOpens.Load() != 1 {
		t.Fatal("resumption opened backend more than once")
	}
}
