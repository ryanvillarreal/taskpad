package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rvillarreal/taskpad/internal/model"
)

func TestMarkdownWriterWriteAndDelete(t *testing.T) {
	dir := t.TempDir()
	writer := NewMarkdownWriter(dir)

	note := &model.Note{
		ID:        "12345678-abcd-efgh-ijkl-1234567890ab",
		Title:     "Quick capture / today",
		Content:   "Pick up the dry cleaning.",
		Tags:      []string{"quick", "errands"},
		CreatedAt: time.Date(2026, 4, 3, 15, 0, 0, 0, time.UTC),
	}

	path, err := writer.Write(note)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if filepath.Ext(path) != ".md" {
		t.Fatalf("expected markdown file, got %q", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, `taskpad_id: "12345678-abcd-efgh-ijkl-1234567890ab"`) {
		t.Fatalf("missing taskpad_id frontmatter: %s", text)
	}
	if !strings.Contains(text, "# Quick capture / today") {
		t.Fatalf("missing markdown heading: %s", text)
	}
	if !strings.Contains(text, "Pick up the dry cleaning.") {
		t.Fatalf("missing note content: %s", text)
	}

	removedPath, err := writer.DeleteByNoteID(note.ID)
	if err != nil {
		t.Fatalf("DeleteByNoteID() error = %v", err)
	}
	if removedPath == "" {
		t.Fatal("expected removed path to be returned")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected note file to be removed, stat err = %v", err)
	}
}

func TestSanitizeFilename(t *testing.T) {
	got := sanitizeFilename(` ideas/brainstorm: april? `)
	want := "ideas-brainstorm- april-"
	if got != want {
		t.Fatalf("sanitizeFilename() = %q, want %q", got, want)
	}
}
