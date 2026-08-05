package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	MaxTokenLength = 1024
)

type Handshake struct {
	Version uint8
	Token   string
}

func (h *Handshake) Encode() ([]byte, error) {

	if h.Version == 0 {
		return nil, errors.New("invalid version")
	}

	if len(h.Token) == 0 {
		return nil, errors.New("empty token")
	}

	if len(h.Token) > MaxTokenLength {
		return nil, errors.New("token too long")
	}

	buf := new(bytes.Buffer)

	buf.WriteByte(h.Version)

	tokenLen := uint16(len(h.Token))

	if err := binary.Write(
		buf,
		binary.BigEndian,
		tokenLen,
	); err != nil {
		return nil, err
	}

	buf.WriteString(h.Token)

	return buf.Bytes(), nil
}

func DecodeHandshake(data []byte) (*Handshake, error) {

	buf := bytes.NewReader(data)

	h := &Handshake{}

	if err := binary.Read(
		buf,
		binary.BigEndian,
		&h.Version,
	); err != nil {
		return nil, err
	}

	if h.Version != Version1 {
		return nil, errors.New("unsupported version")
	}

	var tokenLen uint16

	if err := binary.Read(
		buf,
		binary.BigEndian,
		&tokenLen,
	); err != nil {
		return nil, err
	}

	if tokenLen == 0 {
		return nil, errors.New("empty token")
	}

	if tokenLen > MaxTokenLength {
		return nil, errors.New("token too long")
	}

	token := make([]byte, tokenLen)

	if _, err := buf.Read(token); err != nil {
		return nil, err
	}

	h.Token = string(token)

	return h, nil
}
