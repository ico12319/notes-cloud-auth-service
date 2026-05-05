package random

import "crypto/rand"

type encoder interface {
	EncodeToString(b []byte) string
}
type service struct {
	encoder encoder
}

func NewService(encoder encoder) *service {
	return &service{
		encoder: encoder,
	}
}

func (s *service) Read(b []byte) (n int, err error) {
	return rand.Read(b)
}

func (s *service) GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := s.Read(b); err != nil {
		return "", err
	}

	return s.encoder.EncodeToString(b), nil
}
