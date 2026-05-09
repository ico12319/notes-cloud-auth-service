package api_errors

import (
	"errors"
	"fmt"
)

var ErrEmailAlreadyExist = fmt.Errorf("email alrady exist")

func IsEmailAlreadyExist(err error) bool {
	return errors.Is(err, ErrEmailAlreadyExist)
}
