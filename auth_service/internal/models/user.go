package models

import "time"

type User struct {
	ID           string     `json:"id"`
	PasswordHash *string    `json:"passwordHash"`
	Name         string     `json:"displayName"`
	Email        string     `json:"email"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
}
