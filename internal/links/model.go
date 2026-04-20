package links

import (
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("link not found")
	ErrAmbiguous = errors.New("ambiguous id prefix: multiple links match")
)

type Link struct {
	ID          string
	URL         string
	Title       string
	Description string
	Tags        []string
	SavedAt     time.Time
	Fetched     bool
}
