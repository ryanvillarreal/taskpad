package tasks

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
)

var ErrAmbiguous = errors.New("ambiguous id prefix: multiple tasks match")

var nlp *when.Parser

func init() {
	nlp = when.New(nil)
	nlp.Add(en.All...)
}

type Service struct {
	store *Store
	clock func() time.Time
}

func NewService(store *Store) *Service {
	return &Service{store: store, clock: time.Now}
}

func NewServiceWithClock(store *Store, clock func() time.Time) *Service {
	return &Service{store: store, clock: clock}
}

func (s *Service) Create(title string) (*Task, error) {
	now := s.clock().UTC()

	id, err := newID()
	if err != nil {
		return nil, err
	}

	t := &Task{
		ID:        id,
		Title:     title,
		Status:    StatusActive,
		CreatedAt: now,
	}

	if r, _ := nlp.Parse(title, now); r != nil {
		t.DueAt = r.Time.UTC()
	}

	if err := s.store.Write(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) Get(id string) (*Task, error) {
	return s.store.Read(id)
}

func (s *Service) List() ([]*Task, error) {
	tasks, err := s.store.List()
	if err != nil {
		return nil, err
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})
	return tasks, nil
}

func (s *Service) SetStatus(id string, status Status) (*Task, error) {
	t, err := s.store.Read(id)
	if err != nil {
		return nil, err
	}
	t.Status = status
	if status == StatusClosed {
		t.ClosedAt = s.clock().UTC()
	}
	if err := s.store.Write(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) Due(now time.Time) ([]*Task, error) {
	all, err := s.store.List()
	if err != nil {
		return nil, err
	}
	var due []*Task
	for _, t := range all {
		if t.Status != StatusActive {
			continue
		}
		if !t.DueAt.IsZero() && !t.DueAt.After(now) {
			due = append(due, t)
		}
	}
	return due, nil
}

func (s *Service) Today(now time.Time) ([]*Task, error) {
	all, err := s.store.List()
	if err != nil {
		return nil, err
	}
	y, m, d := now.Date()
	var today []*Task
	for _, t := range all {
		if t.Status != StatusActive {
			continue
		}
		if t.DueAt.IsZero() {
			continue
		}
		ty, tm, td := t.DueAt.In(now.Location()).Date()
		if ty == y && tm == m && td == d {
			today = append(today, t)
		}
	}
	return today, nil
}

func (s *Service) MarkNotified(id string, at time.Time) (*Task, error) {
	t, err := s.store.Read(id)
	if err != nil {
		return nil, err
	}
	t.NotifiedAt = at.UTC()
	if err := s.store.Write(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) Delete(id string) error {
	return s.store.Delete(id)
}

func (s *Service) Resolve(prefix string) (string, error) {
	all, err := s.store.List()
	if err != nil {
		return "", err
	}
	var matches []string
	for _, t := range all {
		if strings.HasPrefix(t.ID, prefix) {
			matches = append(matches, t.ID)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("%w: %s", ErrNotFound, prefix)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("%w: %s", ErrAmbiguous, prefix)
	}
}

func newID() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
