package api_errors

import (
	"errors"
	"fmt"
)

var ErrEmailNotVerified = fmt.Errorf("email not verified")

func IsEmailNotVerified(err error) bool {
	return errors.Is(err, ErrEmailNotVerified)
}
