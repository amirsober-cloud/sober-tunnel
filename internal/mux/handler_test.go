package mux

import (
	"net"
	"testing"

	"github.com/amirsober-cloud/sober-tunnel/internal/protocol"
)

func TestSessionStreamLifecycle(t *testing.T) {

	a, b := net.Pipe()

	defer a.Close()
	defer b.Close()

	session := NewSession(a)

	stream := &Stream{
		ID:      1,
		Session: session,
		Conn:    b,
	}

	session.AddStream(stream)

	found := session.GetStream(1)

	if found == nil {
		t.Fatal("stream not added")
	}

	session.RemoveStream(1)

	found = session.GetStream(1)

	if found != nil {
		t.Fatal("stream not removed")
	}
}

func TestFrameWriteReadThroughPipe(t *testing.T) {

	a, b := net.Pipe()

	defer a.Close()
	defer b.Close()

	frame := &protocol.Frame{

		Version: protocol.Version1,

		Command: protocol.CmdData,

		StreamID: 10,

		Payload: []byte("hello"),
	}

	go func() {

		err := protocol.WriteFrame(
			a,
			frame,
		)

		if err != nil {
			t.Error(err)
		}

	}()

	got, err := protocol.ReadFrame(b)

	if err != nil {
		t.Fatal(err)
	}

	if string(got.Payload) != "hello" {
		t.Fatal("payload mismatch")
	}
}
