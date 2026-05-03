package models

import "time"

type AccessToken struct {
	Token     string    `json:"token"`
	TokenType string    `json:"tokenType"`
	ExpiresIn int       `json:"expiresIn"`
	ExpiresAt time.Time `json:"expiresAt"`
}
