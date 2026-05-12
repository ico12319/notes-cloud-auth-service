package models

import "time"

type RefreshToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userID"`
	RawToken  string     `json:"refreshToken"`
	TokenHash string     `json:"tokenHash"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}
