package protocol

import (
	"crypto/subtle"
)

func VerifyToken(
	token string,
	expected string,
) bool {

	if token == "" || expected == "" {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(token),
		[]byte(expected),
	) == 1
}
