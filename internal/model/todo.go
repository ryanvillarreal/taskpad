package model

import "time"

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type TodoStatus string

const (
	TodoStatusActive TodoStatus = "active"
	TodoStatusPaused TodoStatus = "paused"
	TodoStatusDone   TodoStatus = "done"
)

type TodoUrgency string

const (
	TodoUrgencyNow        TodoUrgency = "now"
	TodoUrgencyHigh       TodoUrgency = "high"
	TodoUrgencyNormal     TodoUrgency = "normal"
	TodoUrgencyLow        TodoUrgency = "low"
	TodoUrgencyBackburner TodoUrgency = "backburner"
)

type Todo struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	Description     string      `json:"description"`
	Completed       bool        `json:"completed"`
	Priority        Priority    `json:"priority"`
	Status          TodoStatus  `json:"status"`
	Urgency         TodoUrgency `json:"urgency"`
	Tags            []string    `json:"tags"`
	DueDate         *time.Time  `json:"due_date,omitempty"`
	CalendarEventID string      `json:"calendar_event_id,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type CreateTodoRequest struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      TodoStatus  `json:"status"`
	Urgency     TodoUrgency `json:"urgency"`
	Tags        []string    `json:"tags"`
	DueDate     *string     `json:"due_date,omitempty"` // RFC3339 string
}

type UpdateTodoRequest struct {
	Title           *string      `json:"title,omitempty"`
	Description     *string      `json:"description,omitempty"`
	Status          *TodoStatus  `json:"status,omitempty"`
	Urgency         *TodoUrgency `json:"urgency,omitempty"`
	Tags            []string     `json:"tags,omitempty"`
	DueDate         *string      `json:"due_date,omitempty"`
	CalendarEventID *string      `json:"calendar_event_id,omitempty"`
}
