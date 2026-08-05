package protocol

import (
	"strings"
	"testing"
)

func TestFrameEncodeDecode(t *testing.T) {

	original := &Frame{
		Version: Version1,
		Command: CmdOpen,

		StreamID: 100,

		Address: "example.com",
		Port:    443,

		Payload: []byte("hello sober tunnel"),
	}

	data, err := original.Encode()

	if err != nil {
		t.Fatalf(
			"encode failed: %v",
			err,
		)
	}

	decoded, err := Decode(data)

	if err != nil {
		t.Fatalf(
			"decode failed: %v",
			err,
		)
	}

	if decoded.Version != original.Version {
		t.Fatal("version mismatch")
	}

	if decoded.Command != original.Command {
		t.Fatal("command mismatch")
	}

	if decoded.StreamID != original.StreamID {
		t.Fatal("stream id mismatch")
	}

	if decoded.Address != original.Address {
		t.Fatal("address mismatch")
	}

	if decoded.Port != original.Port {
		t.Fatal("port mismatch")
	}

	if string(decoded.Payload) != string(original.Payload) {
		t.Fatal("payload mismatch")
	}
}

func TestFrameRejectLargePayload(t *testing.T) {

	frame := &Frame{
		Version: Version1,

		Command: CmdData,

		Address: "localhost",

		Port: 80,

		Payload: make(
			[]byte,
			17*1024*1024,
		),
	}

	_, err := frame.Encode()

	if err == nil {
		t.Fatal(
			"expected large payload rejection",
		)
	}
}

func TestFrameRejectLargeAddress(t *testing.T) {

	frame := &Frame{
		Version: Version1,

		Command: CmdOpen,

		Address: strings.Repeat(
			"a",
			5000,
		),

		Port: 443,
	}

	_, err := frame.Encode()

	if err == nil {
		t.Fatal(
			"expected address rejection",
		)
	}
}
