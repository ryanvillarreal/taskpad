package tasks

import (
	"bytes"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type taskFM struct {
	ID         string  `yaml:"id"`
	Title      string  `yaml:"title"`
	Status     string  `yaml:"status"`
	CreatedAt  string  `yaml:"created_at"`
	DueAt      *string `yaml:"due_at"`
	NotifiedAt *string `yaml:"notified_at"`
	ClosedAt   *string `yaml:"closed_at"`
}

func parseTask(id string, data []byte) (*Task, error) {
	t := &Task{ID: id}

	if !bytes.HasPrefix(data, []byte("---\n")) {
		t.Title = string(bytes.TrimSpace(data))
		t.Status = StatusActive
		return t, nil
	}

	rest := data[4:]
	end := bytes.Index(rest, []byte("\n---\n"))
	if end < 0 {
		t.Title = string(bytes.TrimSpace(data))
		t.Status = StatusActive
		return t, nil
	}

	var fm taskFM
	if err := yaml.Unmarshal(rest[:end], &fm); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	t.ID = fm.ID
	if t.ID == "" {
		t.ID = id
	}
	t.Title = fm.Title
	t.Status = Status(fm.Status)
	if t.Status == "" {
		t.Status = StatusActive
	}
	t.CreatedAt = parseOptTime(fm.CreatedAt)
	t.DueAt = parseOptTime(ptrVal(fm.DueAt))
	t.NotifiedAt = parseOptTime(ptrVal(fm.NotifiedAt))
	t.ClosedAt = parseOptTime(ptrVal(fm.ClosedAt))

	return t, nil
}

func serializeTask(t *Task) ([]byte, error) {
	fm := taskFM{
		ID:        t.ID,
		Title:     t.Title,
		Status:    string(t.Status),
		CreatedAt: t.CreatedAt.UTC().Format(time.RFC3339),
		DueAt:     fmtOptTime(t.DueAt),
		NotifiedAt: fmtOptTime(t.NotifiedAt),
		ClosedAt:  fmtOptTime(t.ClosedAt),
	}

	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(t.Title)
	buf.WriteString("\n")
	return buf.Bytes(), nil
}

func parseOptTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func fmtOptTime(t time.Time) *string {
	if t.IsZero() {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func ptrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
