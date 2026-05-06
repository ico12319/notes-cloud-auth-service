package models

import "time"

type NotificationView struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	ReminderID string     `json:"reminderId"`
	Heading    string     `json:"heading"`
	Message    string     `json:"message"`
	Priority   string     `json:"priority"`
	Read       bool       `json:"read"`
	ReadAt     *time.Time `json:"readAt,omitempty"`
	FiredAt    time.Time  `json:"firedAt"`
}
