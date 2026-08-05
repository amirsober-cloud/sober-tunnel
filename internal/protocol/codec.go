package protocol

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	MaxFrameSize = 16 * 1024 * 1024
)

func ReadFrame(
	r io.Reader,
) (*Frame, error) {

	var size uint32

	if err := binary.Read(
		r,
		binary.BigEndian,
		&size,
	); err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, errors.New("empty frame")
	}

	if size > MaxFrameSize {
		return nil, errors.New("frame too large")
	}

	data := make(
		[]byte,
		size,
	)

	if _, err := io.ReadFull(
		r,
		data,
	); err != nil {
		return nil, err
	}

	return Decode(data)
}

func WriteFrame(
	w io.Writer,
	f *Frame,
) error {

	data, err := f.Encode()

	if err != nil {
		return err
	}

	size := uint32(len(data))

	if size > MaxFrameSize {
		return errors.New("frame too large")
	}

	if err := binary.Write(
		w,
		binary.BigEndian,
		size,
	); err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}
