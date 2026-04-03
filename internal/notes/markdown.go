package notes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rvillarreal/taskpad/internal/model"
)

// Writer syncs taskpad notes to local Markdown files.
type Writer interface {
	Write(note *model.Note) (string, error)
	DeleteByNoteID(noteID string) (string, error)
}

type markdownWriter struct {
	notesDir string
}

// NewMarkdownWriter returns a writer rooted at notesDir.
func NewMarkdownWriter(notesDir string) Writer {
	return &markdownWriter{
		notesDir: notesDir,
	}
}

func (w *markdownWriter) Write(note *model.Note) (string, error) {
	if note == nil {
		return "", fmt.Errorf("note is required")
	}
	if note.ID == "" {
		return "", fmt.Errorf("note ID is required")
	}
	if err := os.MkdirAll(w.notesDir, 0o755); err != nil {
		return "", fmt.Errorf("create notes directory: %w", err)
	}

	path, err := w.findPathByNoteID(note.ID)
	if err != nil {
		return "", err
	}
	if path == "" {
		path, err = w.nextPathForTitle(note.Title, note.ID)
		if err != nil {
			return "", err
		}
	}

	if err := os.WriteFile(path, []byte(renderMarkdown(note)), 0o644); err != nil {
		return "", fmt.Errorf("write note file: %w", err)
	}
	return path, nil
}

func (w *markdownWriter) DeleteByNoteID(noteID string) (string, error) {
	if noteID == "" {
		return "", fmt.Errorf("note ID is required")
	}
	path, err := w.findPathByNoteID(noteID)
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := os.Remove(path); err != nil {
		return "", fmt.Errorf("remove note file: %w", err)
	}
	return path, nil
}

func (w *markdownWriter) findPathByNoteID(noteID string) (string, error) {
	var match string
	needle := fmt.Sprintf("taskpad_id: %q", noteID)

	err := filepath.WalkDir(w.notesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), needle) {
			match = path
			return fs.SkipAll
		}
		return nil
	})
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil && err != fs.SkipAll {
		return "", fmt.Errorf("scan notes directory: %w", err)
	}
	return match, nil
}

func (w *markdownWriter) nextPathForTitle(title, noteID string) (string, error) {
	base := sanitizeFilename(title)
	if base == "" {
		base = "note"
	}

	primary := filepath.Join(w.notesDir, base+".md")
	if _, err := os.Stat(primary); os.IsNotExist(err) {
		return primary, nil
	}

	fallback := filepath.Join(w.notesDir, fmt.Sprintf("%s-%s.md", base, shortID(noteID)))
	if _, err := os.Stat(fallback); os.IsNotExist(err) {
		return fallback, nil
	}

	for i := 2; ; i++ {
		candidate := filepath.Join(w.notesDir, fmt.Sprintf("%s-%s-%d.md", base, shortID(noteID), i))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
}

func renderMarkdown(note *model.Note) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("taskpad_id: %q\n", note.ID))
	b.WriteString(fmt.Sprintf("taskpad_created_at: %q\n", note.CreatedAt.Format("2006-01-02T15:04:05Z07:00")))
	if len(note.Tags) > 0 {
		b.WriteString("tags:\n")
		for _, tag := range note.Tags {
			b.WriteString(fmt.Sprintf("  - %q\n", escapeYAML(tag)))
		}
	}
	b.WriteString("---\n\n")
	b.WriteString("# ")
	b.WriteString(note.Title)
	b.WriteString("\n\n")
	if note.Content != "" {
		b.WriteString(note.Content)
		if !strings.HasSuffix(note.Content, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func sanitizeFilename(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|':
			return '-'
		case r < 32:
			return -1
		default:
			return r
		}
	}, s)

	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.Trim(s, ". ")
}

func shortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

func escapeYAML(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}
