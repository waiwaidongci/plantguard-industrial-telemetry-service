package domain

import (
	"crypto/rand"
	"encoding/hex"
)

func NewID(prefix string) string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand cannot realistically fail on supported platforms.
		panic(err)
	}
	return prefix + "_" + hex.EncodeToString(buf)
}

func NewToken() string {
	return NewID("tok")
}
