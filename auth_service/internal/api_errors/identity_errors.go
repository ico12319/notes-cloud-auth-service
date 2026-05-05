package api_errors

import (
	"errors"
	"fmt"
)

var ErrIdentityNotFound = fmt.Errorf("identity not found")

func IsIdentityNotFoundError(err error) bool {
	return errors.Is(err, ErrIdentityNotFound)
}
