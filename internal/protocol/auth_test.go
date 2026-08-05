package protocol

import "testing"

func TestVerifyTokenValid(t *testing.T) {

	ok := VerifyToken(
		"secret",
		"secret",
	)

	if !ok {
		t.Fatal("expected token to be valid")
	}
}

func TestVerifyTokenInvalid(t *testing.T) {

	ok := VerifyToken(
		"wrong",
		"secret",
	)

	if ok {
		t.Fatal("expected token rejection")
	}
}

func TestVerifyTokenEmpty(t *testing.T) {

	ok := VerifyToken(
		"",
		"secret",
	)

	if ok {
		t.Fatal("expected empty token rejection")
	}
}
