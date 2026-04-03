package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rvillarreal/taskpad/internal/model"
)

// TodoRepository defines the interface for todo data access.
type TodoRepository interface {
	Create(todo *model.Todo) error
	GetByID(id string) (*model.Todo, error)
	List(params model.ListParams, filters model.TodoFilters) (*model.ListResult[model.Todo], error)
	Update(todo *model.Todo) error
	Delete(id string) error
	BulkUpdateCompleted(ids []string, completed bool) (int64, error)
	BulkDelete(ids []string) (int64, error)
}

type todoRepo struct {
	db *sql.DB
}

// NewTodoRepository creates a new TodoRepository backed by SQLite.
func NewTodoRepository(db *sql.DB) TodoRepository {
	return &todoRepo{db: db}
}

func (r *todoRepo) Create(todo *model.Todo) error {
	tagsJSON, err := json.Marshal(todo.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}
	_, err = r.db.Exec(
		`INSERT INTO todos (id, title, description, completed, priority, status, urgency, tags, due_date, calendar_event_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		todo.ID, todo.Title, todo.Description, todo.Completed, todo.Priority, todo.Status, todo.Urgency,
		string(tagsJSON), todo.DueDate, todo.CalendarEventID, todo.CreatedAt, todo.UpdatedAt,
	)
	return err
}

func (r *todoRepo) GetByID(id string) (*model.Todo, error) {
	todo := &model.Todo{}
	var tagsJSON string
	err := r.db.QueryRow(
		`SELECT id, title, description, completed, priority, status, urgency, tags, due_date, calendar_event_id, created_at, updated_at
		 FROM todos WHERE id = ?`, id,
	).Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.Priority, &todo.Status, &todo.Urgency,
		&tagsJSON, &todo.DueDate, &todo.CalendarEventID, &todo.CreatedAt, &todo.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(tagsJSON), &todo.Tags); err != nil {
		return nil, fmt.Errorf("unmarshal tags: %w", err)
	}
	if todo.Tags == nil {
		todo.Tags = []string{}
	}
	return todo, nil
}

func (r *todoRepo) List(params model.ListParams, filters model.TodoFilters) (*model.ListResult[model.Todo], error) {
	where, args := buildTodoWhere(filters)

	// Get total count.
	var total int
	countQuery := "SELECT COUNT(*) FROM todos" + where
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count todos: %w", err)
	}

	// Build and execute list query.
	orderClause := buildOrderClause(params, []string{"created_at", "updated_at", "title", "urgency", "due_date", "status"})
	query := "SELECT id, title, description, completed, priority, status, urgency, tags, due_date, calendar_event_id, created_at, updated_at FROM todos" +
		where + orderClause + " LIMIT ? OFFSET ?"
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	defer rows.Close()

	todos := make([]model.Todo, 0)
	for rows.Next() {
		var t model.Todo
		var tagsJSON string
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.Priority, &t.Status, &t.Urgency,
			&tagsJSON, &t.DueDate, &t.CalendarEventID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsJSON), &t.Tags); err != nil {
			return nil, fmt.Errorf("unmarshal tags: %w", err)
		}
		if t.Tags == nil {
			t.Tags = []string{}
		}
		todos = append(todos, t)
	}

	return &model.ListResult[model.Todo]{
		Data:   todos,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func (r *todoRepo) Update(todo *model.Todo) error {
	todo.UpdatedAt = time.Now().UTC()
	tagsJSON, err := json.Marshal(todo.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}
	_, err = r.db.Exec(
		`UPDATE todos SET title=?, description=?, completed=?, priority=?, status=?, urgency=?, tags=?, due_date=?, calendar_event_id=?, updated_at=?
		 WHERE id=?`,
		todo.Title, todo.Description, todo.Completed, todo.Priority, todo.Status, todo.Urgency, string(tagsJSON),
		todo.DueDate, todo.CalendarEventID, todo.UpdatedAt, todo.ID,
	)
	return err
}

func (r *todoRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM todos WHERE id = ?", id)
	return err
}

func (r *todoRepo) BulkUpdateCompleted(ids []string, completed bool) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]any, 0, len(ids)+3)
	args = append(args, completed, statusFromCompleted(completed), time.Now().UTC())
	for _, id := range ids {
		args = append(args, id)
	}

	result, err := r.db.Exec(
		fmt.Sprintf("UPDATE todos SET completed=?, status=?, updated_at=? WHERE id IN (%s)", placeholders),
		args...,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *todoRepo) BulkDelete(ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	result, err := r.db.Exec(
		fmt.Sprintf("DELETE FROM todos WHERE id IN (%s)", placeholders),
		args...,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func buildTodoWhere(filters model.TodoFilters) (string, []any) {
	var conditions []string
	var args []any

	if filters.Completed != nil {
		if *filters.Completed {
			conditions = append(conditions, "status = ?")
			args = append(args, string(model.TodoStatusDone))
		} else {
			conditions = append(conditions, "status != ?")
			args = append(args, string(model.TodoStatusDone))
		}
	}
	if filters.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, string(*filters.Status))
	}
	if filters.Urgency != nil {
		conditions = append(conditions, "urgency = ?")
		args = append(args, string(*filters.Urgency))
	}
	if filters.Tag != nil {
		conditions = append(conditions, "tags LIKE ?")
		args = append(args, fmt.Sprintf("%%%s%%", *filters.Tag))
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func buildOrderClause(params model.ListParams, allowed []string) string {
	sortField := "created_at"
	for _, a := range allowed {
		if params.Sort == a {
			sortField = params.Sort
			break
		}
	}

	order := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		order = "ASC"
	}

	return fmt.Sprintf(" ORDER BY %s %s", sortField, order)
}

func statusFromCompleted(completed bool) string {
	if completed {
		return string(model.TodoStatusDone)
	}
	return string(model.TodoStatusActive)
}
