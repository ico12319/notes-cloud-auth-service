package encoder

import "encoding/base64"

type service struct{}

func NewService() *service {
	return &service{}
}

func (s *service) EncodeToString(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
