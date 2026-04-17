package notes

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func newTestService(t *testing.T, now time.Time) *Service {
	t.Helper()
	return NewServiceWithClock(NewStore(t.TempDir()), fixedClock(now))
}

func TestService_Today(t *testing.T) {
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	s := newTestService(t, now)
	if got := s.Today(); got != "04.16.2026" {
		t.Errorf("Today() = %q, want 04.16.2026", got)
	}
}

func TestService_Save_FreshNote(t *testing.T) {
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	s := newTestService(t, now)

	n, err := s.Save("04.16.2026", "hello world\n")
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if n.ID != "04.16.2026" {
		t.Errorf("id = %q", n.ID)
	}
	if !n.CreatedAt.Equal(now) {
		t.Errorf("created_at = %v, want %v", n.CreatedAt, now)
	}
	if !n.UpdatedAt.Equal(now) {
		t.Errorf("updated_at = %v, want %v", n.UpdatedAt, now)
	}
}

func TestService_Save_PreservesCreatedAtOnSecondSave(t *testing.T) {
	first := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	second := time.Date(2026, 4, 16, 14, 0, 0, 0, time.UTC)

	store := NewStore(t.TempDir())
	s := NewServiceWithClock(store, fixedClock(first))
	if _, err := s.Save("04.16.2026", "first body\n"); err != nil {
		t.Fatal(err)
	}

	s.clock = fixedClock(second)
	n, err := s.Save("04.16.2026", "second body\n")
	if err != nil {
		t.Fatal(err)
	}
	if !n.CreatedAt.Equal(first) {
		t.Errorf("created_at changed: got %v, want %v", n.CreatedAt, first)
	}
	if !n.UpdatedAt.Equal(second) {
		t.Errorf("updated_at = %v, want %v", n.UpdatedAt, second)
	}
}

func TestService_Save_MergesFrontmatter_PreservesExisting(t *testing.T) {
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	s := newTestService(t, now)

	if _, err := s.Save("04.16.2026", "---\ntags: [work]\n---\n\nfirst\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Save("04.16.2026", "just new body, no frontmatter\n"); err != nil {
		t.Fatal(err)
	}

	got, err := s.Get("04.16.2026")
	if err != nil {
		t.Fatal(err)
	}
	tags, ok := got.Frontmatter["tags"]
	if !ok {
		t.Fatal("tags missing after merge")
	}
	list, _ := tags.([]any)
	if len(list) != 1 || list[0] != "work" {
		t.Errorf("tags = %v, want [work]", tags)
	}
}

func TestService_Save_NewFrontmatterOverridesExisting(t *testing.T) {
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	s := newTestService(t, now)

	if _, err := s.Save("04.16.2026", "---\ntags: [a]\n---\n\nfirst\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Save("04.16.2026", "---\ntags: [b]\n---\n\nsecond\n"); err != nil {
		t.Fatal(err)
	}

	got, err := s.Get("04.16.2026")
	if err != nil {
		t.Fatal(err)
	}
	list, _ := got.Frontmatter["tags"].([]any)
	if len(list) != 1 || list[0] != "b" {
		t.Errorf("tags = %v, want [b]", got.Frontmatter["tags"])
	}
}

func TestService_Delete(t *testing.T) {
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	s := newTestService(t, now)

	if _, err := s.Save("04.16.2026", "body\n"); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("04.16.2026"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.Get("04.16.2026"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Get_Missing(t *testing.T) {
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	s := newTestService(t, now)

	if _, err := s.Get("99.99.9999"); err != ErrNotFound {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}
