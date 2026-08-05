package protocol

import (
	"bytes"
	"testing"
)

func TestFrameCodec(t *testing.T) {

	original := &Frame{
		Version:  Version1,
		Command:  CmdData,
		StreamID: 10,
		Address:  "test.com",
		Port:     443,
		Payload:  []byte("hello"),
	}

	buf := new(bytes.Buffer)

	err := WriteFrame(
		buf,
		original,
	)

	if err != nil {
		t.Fatal(err)
	}

	decoded, err := ReadFrame(buf)

	if err != nil {
		t.Fatal(err)
	}

	if decoded.StreamID != original.StreamID {
		t.Fatal("stream mismatch")
	}

	if string(decoded.Payload) != "hello" {
		t.Fatal("payload mismatch")
	}
}
