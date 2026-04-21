package search

import (
	"strings"
	"time"

	"github.com/ryanvillarreal/taskpad/internal/links"
	"github.com/ryanvillarreal/taskpad/internal/notes"
	"github.com/ryanvillarreal/taskpad/internal/tasks"
)

type Kind string

const (
	KindNote Kind = "note"
	KindLink Kind = "link"
	KindTask Kind = "task"
)

type Result struct {
	Kind    Kind
	ID      string
	Title   string
	Snippet string
	Time    time.Time
}

type Searcher struct {
	notes *notes.Service
	links *links.Service
	tasks *tasks.Service
}

func New(n *notes.Service, l *links.Service, t *tasks.Service) *Searcher {
	return &Searcher{notes: n, links: l, tasks: t}
}

func (s *Searcher) Query(q string, exact bool) ([]Result, error) {
	norm := q
	if !exact {
		norm = strings.ToLower(q)
	}

	var results []Result

	noteIDs, err := s.notes.List()
	if err != nil {
		return nil, err
	}
	for _, id := range noteIDs {
		n, err := s.notes.Get(id)
		if err != nil {
			continue
		}
		if matchesAny(norm, exact, id, n.Body) {
			results = append(results, Result{
				Kind:    KindNote,
				ID:      id,
				Title:   id,
				Snippet: excerpt(n.Body, norm, exact),
				Time:    n.UpdatedAt,
			})
		}
	}

	linkList, err := s.links.List()
	if err != nil {
		return nil, err
	}
	for _, l := range linkList {
		if matchesAny(norm, exact, l.Title, l.Description, l.URL, strings.Join(l.Tags, " ")) {
			results = append(results, Result{
				Kind:    KindLink,
				ID:      l.ID,
				Title:   l.Title,
				Snippet: l.URL,
				Time:    l.SavedAt,
			})
		}
	}

	taskList, err := s.tasks.List()
	if err != nil {
		return nil, err
	}
	for _, t := range taskList {
		if matchesAny(norm, exact, t.Title) {
			results = append(results, Result{
				Kind:    KindTask,
				ID:      t.ID,
				Title:   t.Title,
				Snippet: string(t.Status),
				Time:    t.CreatedAt,
			})
		}
	}

	return results, nil
}

func matchesAny(q string, exact bool, fields ...string) bool {
	for _, f := range fields {
		h := f
		if !exact {
			h = strings.ToLower(f)
		}
		if strings.Contains(h, q) {
			return true
		}
	}
	return false
}

func excerpt(body, q string, exact bool) string {
	search := body
	if !exact {
		search = strings.ToLower(body)
	}
	idx := strings.Index(search, q)
	if idx < 0 {
		if len(body) > 80 {
			return strings.TrimSpace(body[:80])
		}
		return strings.TrimSpace(body)
	}
	start := idx - 40
	if start < 0 {
		start = 0
	}
	end := idx + len(q) + 40
	if end > len(body) {
		end = len(body)
	}
	return strings.TrimSpace(body[start:end])
}
