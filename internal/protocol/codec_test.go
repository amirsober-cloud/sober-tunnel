package protocol

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
)

func TestCodecPipe(t *testing.T) {

	a, b := net.Pipe()

	defer a.Close()
	defer b.Close()

	go func() {

		t.Log("before write")

		err := WriteFrame(
			a,
			&Frame{
				Version:  Version1,
				Command:  1,
				StreamID: 1,
				Address:  "example.com",
				Port:     443,
			},
		)

		t.Log("after write", err)

	}()

	t.Log("before read")

	var size uint32

	err := binary.Read(
		b,
		binary.BigEndian,
		&size,
	)

	t.Log("read size", size, err)

	if err != nil {
		t.Fatal(err)
	}

	data := make([]byte, size)

	_, err = io.ReadFull(
		b,
		data,
	)

	t.Log("read data", len(data), err)

	if err != nil {
		t.Fatal(err)
	}

	frame, err := Decode(data)

	t.Log("decode", frame, err)

	if err != nil {
		t.Fatal(err)
	}
}
