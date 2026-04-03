package repository

import (
	"database/sql"
	"time"

	"github.com/rvillarreal/taskpad/internal/model"
)

// DailyNoteRepository defines data access for daily notes.
type DailyNoteRepository interface {
	Get(date string) (*model.DailyNote, error)
	Upsert(date, content string) (*model.DailyNote, error)
}

type dailyNoteRepo struct {
	db *sql.DB
}

// NewDailyNoteRepository creates a DailyNoteRepository backed by SQLite.
func NewDailyNoteRepository(db *sql.DB) DailyNoteRepository {
	return &dailyNoteRepo{db: db}
}

func (r *dailyNoteRepo) Get(date string) (*model.DailyNote, error) {
	n := &model.DailyNote{}
	err := r.db.QueryRow(
		`SELECT date, content, created_at, updated_at FROM daily_notes WHERE date = ?`, date,
	).Scan(&n.Date, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *dailyNoteRepo) Upsert(date, content string) (*model.DailyNote, error) {
	now := time.Now().UTC()
	_, err := r.db.Exec(`
		INSERT INTO daily_notes (date, content, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE SET
			content    = excluded.content,
			updated_at = excluded.updated_at
	`, date, content, now, now)
	if err != nil {
		return nil, err
	}
	return r.Get(date)
}
