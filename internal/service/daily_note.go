package service

import (
	"fmt"
	"regexp"

	"github.com/rvillarreal/taskpad/internal/model"
	"github.com/rvillarreal/taskpad/internal/repository"
)

var validDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// DailyNoteService defines business logic for daily notes.
type DailyNoteService interface {
	Get(date string) (*model.DailyNote, error)
	Upsert(date, content string) (*model.DailyNote, error)
}

type dailyNoteService struct {
	repo repository.DailyNoteRepository
}

// NewDailyNoteService creates a new DailyNoteService.
func NewDailyNoteService(repo repository.DailyNoteRepository) DailyNoteService {
	return &dailyNoteService{repo: repo}
}

func (s *dailyNoteService) Get(date string) (*model.DailyNote, error) {
	if err := validateDate(date); err != nil {
		return nil, err
	}
	n, err := s.repo.Get(date)
	if err != nil {
		return nil, fmt.Errorf("get daily note: %w", err)
	}
	if n == nil {
		return nil, ErrNotFound
	}
	return n, nil
}

func (s *dailyNoteService) Upsert(date, content string) (*model.DailyNote, error) {
	if err := validateDate(date); err != nil {
		return nil, err
	}
	n, err := s.repo.Upsert(date, content)
	if err != nil {
		return nil, fmt.Errorf("upsert daily note: %w", err)
	}
	return n, nil
}

func validateDate(date string) error {
	if !validDate.MatchString(date) {
		return fmt.Errorf("%w: date must be in YYYY-MM-DD format", ErrValidation)
	}
	return nil
}
