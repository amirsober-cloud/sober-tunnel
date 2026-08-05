package forward

import (
	"net"
	"testing"
	"time"

	"github.com/amirsober-cloud/sober-tunnel/internal/protocol"
)

func TestMuxForwarderOpenAndData(t *testing.T) {

	muxConn, remoteConn := net.Pipe()

	m := NewMuxForwarder(muxConn)

	local, peer := net.Pipe()

	go m.HandleLocal(
		local,
		"example.com:443",
	)

	// منتظر بمان تا OPEN نوشته شود
	time.Sleep(200 * time.Millisecond)

	open, err := protocol.ReadFrame(remoteConn)

	if err != nil {
		t.Fatalf("read open failed: %v", err)
	}

	t.Logf(
		"OPEN: command=%d stream=%d addr=%s port=%d",
		open.Command,
		open.StreamID,
		open.Address,
		open.Port,
	)

	if open.Command != protocol.CmdOpen {
		t.Fatal("expected CmdOpen")
	}

	go func() {
		_, _ = peer.Write([]byte("hello"))
	}()

	data, err := protocol.ReadFrame(remoteConn)

	if err != nil {
		t.Fatalf("read data failed: %v", err)
	}

	if string(data.Payload) != "hello" {
		t.Fatalf(
			"payload mismatch: %s",
			string(data.Payload),
		)
	}
}
