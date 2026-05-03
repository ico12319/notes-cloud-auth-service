package api_errors

import (
	"errors"
	"fmt"
)

var ErrWrongLoginCredentials = fmt.Errorf("wrong email or password provided")

func IsWrongLoginCredentialsError(err error) bool {
	return errors.Is(err, ErrWrongLoginCredentials)
}
