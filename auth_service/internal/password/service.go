package password

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

const (
	MinimumPasswordLength = 8
	MaximumPasswordLength = 128
)

type service struct{}

func NewService() *service {
	return &service{}
}

func (*service) GeneratePasswordHash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (*service) CompareHashAndPassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (*service) ValidatePassword(password *string) error {
	if password == nil {
		return fmt.Errorf("password must be at least %d symbols long", MinimumPasswordLength)
	}

	trimmedPassword := strings.TrimSpace(*password)
	if trimmedPassword != *password {
		return errors.New("password must not start or end with spaces")
	}

	if len(*password) < MinimumPasswordLength {
		return fmt.Errorf("password must be minumum %d symbols long", MinimumPasswordLength)
	}

	if len(*password) > MaximumPasswordLength {
		return fmt.Errorf("password must be maximum %d symbols long", MaximumPasswordLength)
	}

	passwordRunes := []rune(*password)
	firstPasswordRune := passwordRunes[0]

	if !strings.ContainsFunc(*password, func(r rune) bool {
		return r != firstPasswordRune
	}) {
		return errors.New("password must not be one repeated character")
	}

	return nil
}
