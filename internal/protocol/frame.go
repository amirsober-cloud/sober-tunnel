package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	Version1 uint8 = 1
)

type Frame struct {
	Version uint8

	Command uint8

	StreamID uint32

	Address string

	Port uint16

	Payload []byte
}

func (f *Frame) Encode() ([]byte, error) {

	if len(f.Address) > 4096 {
		return nil, errors.New("address too long")
	}

	if len(f.Payload) > 16*1024*1024 {
		return nil, errors.New("payload too large")
	}

	buf := new(bytes.Buffer)

	buf.WriteByte(f.Version)

	buf.WriteByte(f.Command)

	if err := binary.Write(
		buf,
		binary.BigEndian,
		f.StreamID,
	); err != nil {
		return nil, err
	}

	addrLen := uint16(len(f.Address))

	if err := binary.Write(
		buf,
		binary.BigEndian,
		addrLen,
	); err != nil {
		return nil, err
	}

	buf.WriteString(f.Address)

	if err := binary.Write(
		buf,
		binary.BigEndian,
		f.Port,
	); err != nil {
		return nil, err
	}

	payloadLen := uint32(len(f.Payload))

	if err := binary.Write(
		buf,
		binary.BigEndian,
		payloadLen,
	); err != nil {
		return nil, err
	}

	buf.Write(f.Payload)

	return buf.Bytes(), nil
}

func Decode(data []byte) (*Frame, error) {

	buf := bytes.NewReader(data)

	f := &Frame{}

	if err := binary.Read(
		buf,
		binary.BigEndian,
		&f.Version,
	); err != nil {
		return nil, err
	}

	if err := binary.Read(
		buf,
		binary.BigEndian,
		&f.Command,
	); err != nil {
		return nil, err
	}

	if err := binary.Read(
		buf,
		binary.BigEndian,
		&f.StreamID,
	); err != nil {
		return nil, err
	}

	var addrLen uint16

	if err := binary.Read(
		buf,
		binary.BigEndian,
		&addrLen,
	); err != nil {
		return nil, err
	}

	if addrLen > 4096 {
		return nil, errors.New("address too long")
	}

	addr := make([]byte, addrLen)

	if _, err := io.ReadFull(
		buf,
		addr,
	); err != nil {
		return nil, err
	}

	f.Address = string(addr)

	if err := binary.Read(
		buf,
		binary.BigEndian,
		&f.Port,
	); err != nil {
		return nil, err
	}

	var payloadLen uint32

	if err := binary.Read(
		buf,
		binary.BigEndian,
		&payloadLen,
	); err != nil {
		return nil, err
	}

	if payloadLen > 16*1024*1024 {
		return nil, errors.New("payload too large")
	}

	if payloadLen > 0 {

		f.Payload = make(
			[]byte,
			payloadLen,
		)

		if _, err := io.ReadFull(
			buf,
			f.Payload,
		); err != nil {
			return nil, err
		}

	} else {

		f.Payload = nil
	}

	return f, nil
}
