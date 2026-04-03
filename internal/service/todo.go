package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rvillarreal/taskpad/internal/model"
	"github.com/rvillarreal/taskpad/internal/repository"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation error")
)

// TodoService defines the interface for todo business logic.
type TodoService interface {
	Create(req model.CreateTodoRequest) (*model.Todo, error)
	GetByID(id string) (*model.Todo, error)
	List(params model.ListParams, filters model.TodoFilters) (*model.ListResult[model.Todo], error)
	Update(id string, req model.UpdateTodoRequest) (*model.Todo, error)
	Delete(id string) error
	SetCompleted(id string, completed bool) (*model.Todo, error)
	BulkComplete(ids []string) (int64, error)
	BulkDelete(ids []string) (int64, error)
}

type todoService struct {
	repo repository.TodoRepository
}

// NewTodoService creates a new TodoService.
func NewTodoService(repo repository.TodoRepository) TodoService {
	return &todoService{repo: repo}
}

func (s *todoService) Create(req model.CreateTodoRequest) (*model.Todo, error) {
	if err := validateTitle(req.Title); err != nil {
		return nil, err
	}
	if err := validateStatus(req.Status); err != nil {
		return nil, err
	}
	if err := validateUrgency(req.Urgency); err != nil {
		return nil, err
	}
	if err := validateTodoTags(req.Tags); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	todo := &model.Todo{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Urgency:     req.Urgency,
		Tags:        req.Tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if todo.Status == "" {
		todo.Status = model.TodoStatusActive
	}
	if todo.Urgency == "" {
		todo.Urgency = model.TodoUrgencyNormal
	}
	syncLegacyFields(todo)

	if req.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("%w: due_date must be valid RFC3339 format", ErrValidation)
		}
		todo.DueDate = &t
	}

	if err := s.repo.Create(todo); err != nil {
		return nil, fmt.Errorf("create todo: %w", err)
	}
	return todo, nil
}

func (s *todoService) GetByID(id string) (*model.Todo, error) {
	todo, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get todo: %w", err)
	}
	if todo == nil {
		return nil, ErrNotFound
	}
	return todo, nil
}

func (s *todoService) List(params model.ListParams, filters model.TodoFilters) (*model.ListResult[model.Todo], error) {
	params = sanitizeListParams(params)
	return s.repo.List(params, filters)
}

func (s *todoService) Update(id string, req model.UpdateTodoRequest) (*model.Todo, error) {
	todo, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get todo: %w", err)
	}
	if todo == nil {
		return nil, ErrNotFound
	}

	if req.Title != nil {
		if err := validateTitle(*req.Title); err != nil {
			return nil, err
		}
		todo.Title = *req.Title
	}
	if req.Description != nil {
		todo.Description = *req.Description
	}
	if req.Status != nil {
		if err := validateStatus(*req.Status); err != nil {
			return nil, err
		}
		todo.Status = *req.Status
	}
	if req.Urgency != nil {
		if err := validateUrgency(*req.Urgency); err != nil {
			return nil, err
		}
		todo.Urgency = *req.Urgency
	}
	if req.Tags != nil {
		if err := validateTodoTags(req.Tags); err != nil {
			return nil, err
		}
		todo.Tags = req.Tags
	}
	if req.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("%w: due_date must be valid RFC3339 format", ErrValidation)
		}
		todo.DueDate = &t
	}
	if req.CalendarEventID != nil {
		todo.CalendarEventID = *req.CalendarEventID
	}

	syncLegacyFields(todo)

	if err := s.repo.Update(todo); err != nil {
		return nil, fmt.Errorf("update todo: %w", err)
	}
	return todo, nil
}

func (s *todoService) Delete(id string) error {
	todo, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("get todo: %w", err)
	}
	if todo == nil {
		return ErrNotFound
	}
	return s.repo.Delete(id)
}

func (s *todoService) SetCompleted(id string, completed bool) (*model.Todo, error) {
	todo, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get todo: %w", err)
	}
	if todo == nil {
		return nil, ErrNotFound
	}

	if completed {
		todo.Status = model.TodoStatusDone
	} else {
		todo.Status = model.TodoStatusActive
	}
	syncLegacyFields(todo)

	if err := s.repo.Update(todo); err != nil {
		return nil, fmt.Errorf("update todo: %w", err)
	}
	return todo, nil
}

func (s *todoService) BulkComplete(ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("%w: ids must not be empty", ErrValidation)
	}
	if len(ids) > 100 {
		return 0, fmt.Errorf("%w: cannot bulk operate on more than 100 items", ErrValidation)
	}
	return s.repo.BulkUpdateCompleted(ids, true)
}

func (s *todoService) BulkDelete(ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("%w: ids must not be empty", ErrValidation)
	}
	if len(ids) > 100 {
		return 0, fmt.Errorf("%w: cannot bulk operate on more than 100 items", ErrValidation)
	}
	return s.repo.BulkDelete(ids)
}

func validateTitle(title string) error {
	if title == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if len(title) > 500 {
		return fmt.Errorf("%w: title must be 500 characters or less", ErrValidation)
	}
	return nil
}

func validateStatus(status model.TodoStatus) error {
	if status == "" {
		return nil
	}
	switch status {
	case model.TodoStatusActive, model.TodoStatusPaused, model.TodoStatusDone:
		return nil
	default:
		return fmt.Errorf("%w: status must be active, paused, or done", ErrValidation)
	}
}

func validateUrgency(urgency model.TodoUrgency) error {
	if urgency == "" {
		return nil
	}
	switch urgency {
	case model.TodoUrgencyNow, model.TodoUrgencyHigh, model.TodoUrgencyNormal, model.TodoUrgencyLow, model.TodoUrgencyBackburner:
		return nil
	default:
		return fmt.Errorf("%w: urgency must be now, high, normal, low, or backburner", ErrValidation)
	}
}

func validateTodoTags(tags []string) error {
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

func sanitizeListParams(p model.ListParams) model.ListParams {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}

func syncLegacyFields(todo *model.Todo) {
	todo.Completed = todo.Status == model.TodoStatusDone
	switch todo.Urgency {
	case model.TodoUrgencyNow, model.TodoUrgencyHigh:
		todo.Priority = model.PriorityHigh
	case model.TodoUrgencyLow, model.TodoUrgencyBackburner:
		todo.Priority = model.PriorityLow
	default:
		todo.Priority = model.PriorityMedium
	}
	if todo.Tags == nil {
		todo.Tags = []string{}
	}
}
