package notes

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("note not found")

type Note struct {
	ID           string
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Frontmatter  map[string]any
}
