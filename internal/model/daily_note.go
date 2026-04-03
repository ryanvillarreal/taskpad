package model

import "time"

// DailyNote represents a single day's note, keyed by ISO date (YYYY-MM-DD).
type DailyNote struct {
	Date      string    `json:"date"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpsertDailyNoteRequest is the body for creating or updating a daily note.
type UpsertDailyNoteRequest struct {
	Content string `json:"content"`
}
