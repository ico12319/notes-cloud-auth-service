package models

import "time"

type AccessToken struct {
	Token     string
	TokenType string
	ExpiresIn int
	ExpiresAt time.Time
}
