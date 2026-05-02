package random

import "crypto/rand"

type service struct{}

func NewService() *service {
	return &service{}
}

func (s *service) Read(b []byte) (n int, err error) {
	return rand.Read(b)
}
