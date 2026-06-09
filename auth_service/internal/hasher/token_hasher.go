package hasher

import (
	"crypto/sha256"
	"encoding/base64"
)

type tokenHasher struct{}

func NewTokenHasher() *tokenHasher {
	return &tokenHasher{}
}

func (*tokenHasher) Hash(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))

	return base64.RawURLEncoding.EncodeToString(hash[:])
}
