package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rvillarreal/taskpad/internal/model"
	"github.com/rvillarreal/taskpad/internal/repository"
)

// NoteService defines the interface for note business logic.
type NoteService interface {
	Create(req model.CreateNoteRequest) (*model.Note, error)
	GetByID(id string) (*model.Note, error)
	List(params model.ListParams, filters model.NoteFilters) (*model.ListResult[model.Note], error)
	Update(id string, req model.UpdateNoteRequest) (*model.Note, error)
	Delete(id string) error
	BulkDelete(ids []string) (int64, error)
}

type noteService struct {
	repo repository.NoteRepository
}

// NewNoteService creates a new NoteService.
func NewNoteService(repo repository.NoteRepository) NoteService {
	return &noteService{repo: repo}
}

func (s *noteService) Create(req model.CreateNoteRequest) (*model.Note, error) {
	if err := validateTitle(req.Title); err != nil {
		return nil, err
	}
	if err := validateNoteContent(req.Content); err != nil {
		return nil, err
	}
	if err := validateTags(req.Tags); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	note := &model.Note{
		ID:        uuid.New().String(),
		Title:     req.Title,
		Content:   req.Content,
		Tags:      req.Tags,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if note.Tags == nil {
		note.Tags = []string{}
	}

	if err := s.repo.Create(note); err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}
	return note, nil
}

func (s *noteService) GetByID(id string) (*model.Note, error) {
	note, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get note: %w", err)
	}
	if note == nil {
		return nil, ErrNotFound
	}
	return note, nil
}

func (s *noteService) List(params model.ListParams, filters model.NoteFilters) (*model.ListResult[model.Note], error) {
	params = sanitizeListParams(params)
	return s.repo.List(params, filters)
}

func (s *noteService) Update(id string, req model.UpdateNoteRequest) (*model.Note, error) {
	note, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get note: %w", err)
	}
	if note == nil {
		return nil, ErrNotFound
	}

	if req.Title != nil {
		if err := validateTitle(*req.Title); err != nil {
			return nil, err
		}
		note.Title = *req.Title
	}
	if req.Content != nil {
		if err := validateNoteContent(*req.Content); err != nil {
			return nil, err
		}
		note.Content = *req.Content
	}
	if req.Tags != nil {
		if err := validateTags(req.Tags); err != nil {
			return nil, err
		}
		note.Tags = req.Tags
	}

	if err := s.repo.Update(note); err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}
	return note, nil
}

func (s *noteService) Delete(id string) error {
	note, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("get note: %w", err)
	}
	if note == nil {
		return ErrNotFound
	}
	return s.repo.Delete(id)
}

func (s *noteService) BulkDelete(ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("%w: ids must not be empty", ErrValidation)
	}
	if len(ids) > 100 {
		return 0, fmt.Errorf("%w: cannot bulk operate on more than 100 items", ErrValidation)
	}
	return s.repo.BulkDelete(ids)
}

func validateNoteContent(content string) error {
	if len(content) > 50000 {
		return fmt.Errorf("%w: content must be 50000 characters or less", ErrValidation)
	}
	return nil
}

func validateTags(tags []string) error {
	if len(tags) > 20 {
		return fmt.Errorf("%w: maximum 20 tags allowed", ErrValidation)
	}
	for _, tag := range tags {
		if len(tag) > 100 {
			return fmt.Errorf("%w: each tag must be 100 characters or less", ErrValidation)
		}
	}
	return nil
}
