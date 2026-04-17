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
	for _, layout := range []string{"1.2.2006", "1.2.06"} {
		if t, err := time.Parse(layout, input); err == nil {
			return t.Format("01.02.2006"), nil
		}
	}
	return "", fmt.Errorf("invalid date %q (expected MM.DD.YYYY or MM.DD.YY)", input)
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

	if err := saveBody(c, svc, id, newBody); err != nil {
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

func saveBody(c *Client, svc *notes.Service, id string, body []byte) error {
	err := c.Save(id, body)
	switch err {
	case nil:
		return nil
	case ErrUnreachable:
		slog.Warn("server unreachable, writing to local filesystem")
	default:
		return err
	}
	_, err = svc.Save(id, string(body))
	return err
}

func DeleteNote(rawID string) error {
	id, err := NormalizeID(rawID)
	if err != nil {
		return err
	}
	cfg := config.Load()
	svc := notes.NewService(notes.NewStore(cfg.NotesDir))
	c := New(cfg.BaseURL)

	slog.Debug("deleting note", "id", id)

	switch err := c.Delete(id); err {
	case nil:
		slog.Info("note deleted", "id", id, "source", "server")
		return nil
	case ErrNotFound:
		return ErrMissingNote
	case ErrUnreachable:
		slog.Warn("server unreachable, deleting from local filesystem")
	default:
		return err
	}

	if err := svc.Delete(id); err != nil {
		if err == notes.ErrNotFound {
			return ErrMissingNote
		}
		return err
	}
	slog.Info("note deleted", "id", id, "source", "filesystem")
	return nil
}

func writeTmp(id string, body []byte) (string, error) {
	path := filepath.Join(os.TempDir(), "taskpad-"+id+".md")
	if err := os.WriteFile(path, body, 0o600); err != nil {
		return "", err
	}
	slog.Debug("tmpfile written", "path", path)
	return path, nil
}
