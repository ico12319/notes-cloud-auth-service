package api_errors

import (
	"errors"
	"fmt"
)

var ErrUserNotFound = fmt.Errorf("user not found")

func IsUserNotFoundError(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
