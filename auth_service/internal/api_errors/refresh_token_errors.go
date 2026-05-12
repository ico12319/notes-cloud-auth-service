package api_errors

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidRefreshToken = fmt.Errorf("refresh token provided in the request is invalid")
	ErrRefreshTokenRevoked = fmt.Errorf("refresh token provided in the request is revoked")
	ErrExpiredRefreshToken = fmt.Errorf("refresh token provided in the request is expired")
	ErrTokenNotFound       = fmt.Errorf("refresh token does not exist")
)

func IsInvalidRefreshTokenError(err error) bool {
	return errors.Is(err, ErrRefreshTokenRevoked) || errors.Is(err, ErrExpiredRefreshToken) || errors.Is(err, ErrInvalidRefreshToken)
}

func IsRefreshTokenNotFound(err error) bool {
	return errors.Is(err, ErrInvalidRefreshToken)
}
