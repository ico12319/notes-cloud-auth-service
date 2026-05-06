package models

import "time"

type ReminderRequest struct {
	ID           string `json:"id,omitempty"`
	UserID       string `json:"userId,omitempty"`
	Heading      string `json:"heading"`
	Description  string `json:"description,omitempty"`
	ReminderDate string `json:"reminderDate"`
	ReminderTime string `json:"reminderTime"`
	Priority     string `json:"priority"`
	Status       string `json:"status,omitempty"`
	NotifyInApp  bool   `json:"notifyInApp"`
}

type ReminderView struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	Heading      string    `json:"heading"`
	Description  string    `json:"description"`
	ReminderDate string    `json:"reminderDate"`
	ReminderTime string    `json:"reminderTime"`
	Priority     string    `json:"priority"`
	Status       string    `json:"status"`
	NotifyInApp  bool      `json:"notifyInApp"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
