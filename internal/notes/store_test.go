package notes

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	return NewStore(t.TempDir())
}

func TestStore_WriteThenRead(t *testing.T) {
	s := newTestStore(t)
	orig := &Note{
		ID:          "04.16.2026",
		Body:        "hello\n",
		CreatedAt:   time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC),
		Frontmatter: map[string]any{},
	}
	if err := s.Write(orig); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := s.Read("04.16.2026")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Body != orig.Body {
		t.Errorf("body: got %q want %q", got.Body, orig.Body)
	}
}

func TestStore_ReadMissing_ReturnsErrNotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Read("99.99.9999")
	if err != ErrNotFound {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestStore_DeleteMissing_ReturnsErrNotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.Delete("99.99.9999")
	if err != ErrNotFound {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestStore_Count(t *testing.T) {
	s := newTestStore(t)

	n, err := s.Count()
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Errorf("empty dir count = %d, want 0", n)
	}

	for _, id := range []string{"04.16.2026", "04.17.2026", "04.18.2026"} {
		note := &Note{ID: id, Body: "x", Frontmatter: map[string]any{}}
		if err := s.Write(note); err != nil {
			t.Fatal(err)
		}
	}
	n, err = s.Count()
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 3 {
		t.Errorf("count = %d, want 3", n)
	}
}

func TestStore_CountIgnoresNonMarkdown(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	note := &Note{ID: "04.16.2026", Body: "x", Frontmatter: map[string]any{}}
	if err := s.Write(note); err != nil {
		t.Fatal(err)
	}

	n, err := s.Count()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("count = %d, want 1", n)
	}
}

func TestStore_AtomicWrite_NoTmpLeftBehind(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	note := &Note{ID: "04.16.2026", Body: "x", Frontmatter: map[string]any{}}
	if err := s.Write(note); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("leftover tmp file: %s", e.Name())
		}
	}
}
