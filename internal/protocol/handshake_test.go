package protocol

import "testing"

func TestHandshakeEncodeDecode(t *testing.T) {

	original := &Handshake{
		Version: Version1,
		Token:   "secret-token",
	}

	data, err := original.Encode()

	if err != nil {
		t.Fatal(err)
	}

	decoded, err := DecodeHandshake(data)

	if err != nil {
		t.Fatal(err)
	}

	if decoded.Version != original.Version {
		t.Fatal("version mismatch")
	}

	if decoded.Token != original.Token {
		t.Fatal("token mismatch")
	}
}

func TestHandshakeRejectEmptyToken(t *testing.T) {

	h := &Handshake{
		Version: Version1,
		Token:   "",
	}

	_, err := h.Encode()

	if err == nil {
		t.Fatal("expected empty token rejection")
	}
}

func TestHandshakeRejectWrongVersion(t *testing.T) {

	data := []byte{
		99,
		0,
		5,
		"hello"[0],
		"hello"[1],
		"hello"[2],
		"hello"[3],
		"hello"[4],
	}

	_, err := DecodeHandshake(data)

	if err == nil {
		t.Fatal("expected version rejection")
	}
}
