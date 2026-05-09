package models

import "time"

type User struct {
	ID            string     `json:"id"`
	PasswordHash  *string    `json:"passwordHash"`
	EmailVerified bool       `json:"emailVerified"`
	Name          string     `json:"displayName"`
	Email         string     `json:"email"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     *time.Time `json:"updatedAt,omitempty"`
}

type UserPatch struct {
	DisplayName   *string `json:"displayName,omitempty"`
	Email         *string `json:"email,omitempty"`
	EmailVerified *bool   `json:"emailVerified,omitempty"`
	PasswordHash  *string `json:"-"`
}
