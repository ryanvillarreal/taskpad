package tasks

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Status string

const (
	StatusActive Status = "active"
	StatusPaused Status = "paused"
	StatusClosed Status = "closed"
)

type Task struct {
	ID         string
	Title      string
	Status     Status
	CreatedAt  time.Time
	DueAt      time.Time
	NotifiedAt time.Time
	ClosedAt   time.Time
}
