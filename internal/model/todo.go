package model

import "time"

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Todo struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Completed       bool       `json:"completed"`
	Priority        Priority   `json:"priority"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	CalendarEventID string     `json:"calendar_event_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateTodoRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	DueDate     *string  `json:"due_date,omitempty"` // RFC3339 string
}

type UpdateTodoRequest struct {
	Title           *string   `json:"title,omitempty"`
	Description     *string   `json:"description,omitempty"`
	Priority        *Priority `json:"priority,omitempty"`
	DueDate         *string   `json:"due_date,omitempty"`
	CalendarEventID *string   `json:"calendar_event_id,omitempty"`
}
