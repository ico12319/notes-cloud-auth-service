package models

import "time"

type ShareLinkView struct {
	ID        string    `json:"id"`
	NoteID    string    `json:"noteId"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type SharedNoteView struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Color     string    `json:"color"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}