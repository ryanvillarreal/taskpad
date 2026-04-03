package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rvillarreal/taskpad/internal/model"
)

// NoteRepository defines the interface for note data access.
type NoteRepository interface {
	Create(note *model.Note) error
	GetByID(id string) (*model.Note, error)
	List(params model.ListParams, filters model.NoteFilters) (*model.ListResult[model.Note], error)
	Update(note *model.Note) error
	Delete(id string) error
	BulkDelete(ids []string) (int64, error)
}

type noteRepo struct {
	db *sql.DB
}

// NewNoteRepository creates a new NoteRepository backed by SQLite.
func NewNoteRepository(db *sql.DB) NoteRepository {
	return &noteRepo{db: db}
}

func (r *noteRepo) Create(note *model.Note) error {
	tagsJSON, err := json.Marshal(note.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}
	_, err = r.db.Exec(
		`INSERT INTO notes (id, title, content, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		note.ID, note.Title, note.Content, string(tagsJSON), note.CreatedAt, note.UpdatedAt,
	)
	return err
}

func (r *noteRepo) GetByID(id string) (*model.Note, error) {
	note := &model.Note{}
	var tagsJSON string
	err := r.db.QueryRow(
		`SELECT id, title, content, tags, created_at, updated_at
		 FROM notes WHERE id = ?`, id,
	).Scan(&note.ID, &note.Title, &note.Content, &tagsJSON, &note.CreatedAt, &note.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &note.Tags); err != nil {
		return nil, fmt.Errorf("unmarshal tags: %w", err)
	}
	if note.Tags == nil {
		note.Tags = []string{}
	}
	return note, nil
}

func (r *noteRepo) List(params model.ListParams, filters model.NoteFilters) (*model.ListResult[model.Note], error) {
	where, args := buildNoteWhere(filters)

	// Get total count.
	var total int
	countQuery := "SELECT COUNT(*) FROM notes" + where
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count notes: %w", err)
	}

	// Build and execute list query.
	orderClause := buildOrderClause(params, []string{"created_at", "updated_at", "title"})
	query := "SELECT id, title, content, tags, created_at, updated_at FROM notes" +
		where + orderClause + " LIMIT ? OFFSET ?"
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	notes := make([]model.Note, 0)
	for rows.Next() {
		var n model.Note
		var tagsJSON string
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &tagsJSON, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsJSON), &n.Tags); err != nil {
			return nil, fmt.Errorf("unmarshal tags: %w", err)
		}
		if n.Tags == nil {
			n.Tags = []string{}
		}
		notes = append(notes, n)
	}

	return &model.ListResult[model.Note]{
		Data:   notes,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func (r *noteRepo) Update(note *model.Note) error {
	note.UpdatedAt = time.Now().UTC()
	tagsJSON, err := json.Marshal(note.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}
	_, err = r.db.Exec(
		`UPDATE notes SET title=?, content=?, tags=?, updated_at=? WHERE id=?`,
		note.Title, note.Content, string(tagsJSON), note.UpdatedAt, note.ID,
	)
	return err
}

func (r *noteRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM notes WHERE id = ?", id)
	return err
}

func (r *noteRepo) BulkDelete(ids []string) (int64, error) {
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
		fmt.Sprintf("DELETE FROM notes WHERE id IN (%s)", placeholders),
		args...,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func buildNoteWhere(filters model.NoteFilters) (string, []any) {
	var conditions []string
	var args []any

	if filters.Tag != nil {
		// Search in JSON array of tags using LIKE.
		conditions = append(conditions, "tags LIKE ?")
		args = append(args, fmt.Sprintf("%%%s%%", *filters.Tag))
	}
	if filters.Search != nil {
		conditions = append(conditions, "(title LIKE ? OR content LIKE ?)")
		search := fmt.Sprintf("%%%s%%", *filters.Search)
		args = append(args, search, search)
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}
