package client

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/ryanvillarreal/taskpad/internal/config"
	"github.com/ryanvillarreal/taskpad/internal/editor"
	"github.com/ryanvillarreal/taskpad/internal/notes"
)

var ErrMissingNote = errors.New("note does not exist (pass --new to create)")

func NormalizeID(input string) (string, error) {
	t, err := time.Parse("1.2.2006", input)
	if err != nil {
		return "", fmt.Errorf("invalid date %q (expected MM.DD.YYYY)", input)
	}
	return t.Format("01.02.2006"), nil
}

func EditNote(rawID string, allowCreate bool) error {
	cfg := config.Load()
	svc := notes.NewService(notes.NewStore(cfg.NotesDir))
	c := New(cfg.BaseURL)

	id := svc.Today()
	if rawID != "" {
		norm, err := NormalizeID(rawID)
		if err != nil {
			return err
		}
		id = norm
	}
	slog.Debug("editing note", "id", id, "allow_create", allowCreate)

	body, source, err := loadBody(c, svc, id, allowCreate)
	if err != nil {
		return err
	}
	slog.Info("note loaded", "id", id, "source", source, "bytes", len(body))

	tmp, err := writeTmp(id, body)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	if err := editor.Run(tmp); err != nil {
		slog.Error("editor failed", "err", err)
		return err
	}

	newBody, err := os.ReadFile(tmp)
	if err != nil {
		return err
	}

	if bytes.Equal(body, newBody) {
		slog.Info("no changes")
		return nil
	}

	if err := saveBody(c, svc, newBody); err != nil {
		return err
	}
	slog.Info("note saved", "id", id, "bytes", len(newBody))
	return nil
}

func loadBody(c *Client, svc *notes.Service, id string, allowCreate bool) ([]byte, string, error) {
	body, err := c.Get(id)
	switch err {
	case nil:
		return body, "server", nil
	case ErrNotFound:
		if !allowCreate {
			return nil, "", ErrMissingNote
		}
		slog.Debug("no existing note, using template")
		return []byte(notes.TodayTemplate(id)), "template", nil
	case ErrUnreachable:
		slog.Warn("server unreachable, using local filesystem")
	default:
		return nil, "", err
	}

	raw, err := svc.Raw(id)
	if err == notes.ErrNotFound {
		if !allowCreate {
			return nil, "", ErrMissingNote
		}
		return []byte(notes.TodayTemplate(id)), "template", nil
	}
	if err != nil {
		return nil, "", err
	}
	return raw, "filesystem", nil
}

func saveBody(c *Client, svc *notes.Service, body []byte) error {
	err := c.Save(body)
	switch err {
	case nil:
		return nil
	case ErrUnreachable:
		slog.Warn("server unreachable, writing to local filesystem")
	default:
		return err
	}
	_, err = svc.Save(string(body))
	return err
}

func writeTmp(id string, body []byte) (string, error) {
	path := filepath.Join(os.TempDir(), "taskpad-"+id+".md")
	if err := os.WriteFile(path, body, 0o600); err != nil {
		return "", err
	}
	slog.Debug("tmpfile written", "path", path)
	return path, nil
}
