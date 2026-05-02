package password

import "golang.org/x/crypto/bcrypt"

type service struct{}

func NewService() *service {
	return &service{}
}

func (s *service) GeneratePasswordHash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}
