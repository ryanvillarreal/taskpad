package notes

import (
	"strings"
	"testing"
	"time"
)

func TestParseNote_NoFrontmatter(t *testing.T) {
	data := []byte("just some body content\nwith newlines\n")
	n, err := parseNote("04.16.2026", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Body != string(data) {
		t.Errorf("body mismatch:\ngot:  %q\nwant: %q", n.Body, string(data))
	}
	if len(n.Frontmatter) != 0 {
		t.Errorf("expected empty frontmatter, got %v", n.Frontmatter)
	}
}

func TestParseNote_StandardFrontmatter(t *testing.T) {
	data := []byte("---\nid: 04.16.2026\ncreated_at: 2026-04-16T10:00:00Z\nupdated_at: 2026-04-16T11:00:00Z\n---\n\nbody here\n")
	n, err := parseNote("04.16.2026", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Body != "body here\n" {
		t.Errorf("body = %q", n.Body)
	}
	want := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	if !n.CreatedAt.Equal(want) {
		t.Errorf("created_at = %v, want %v", n.CreatedAt, want)
	}
}

func TestParseNote_UnknownKeysPreserved(t *testing.T) {
	data := []byte("---\nid: 04.16.2026\naliases: [today, daily]\ncssclass: note\n---\n\nbody\n")
	n, err := parseNote("04.16.2026", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := n.Frontmatter["aliases"]; !ok {
		t.Error("aliases not preserved")
	}
	if v, ok := n.Frontmatter["cssclass"].(string); !ok || v != "note" {
		t.Errorf("cssclass = %v", n.Frontmatter["cssclass"])
	}
}

func TestParseNote_MalformedFrontmatter_TreatedAsBody(t *testing.T) {
	data := []byte("---\nid: 04.16.2026\nno closing marker\njust more text\n")
	n, err := parseNote("04.16.2026", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Body != string(data) {
		t.Error("malformed frontmatter should fall back to treating whole input as body")
	}
}

func TestSerializeNote_RoundTrip(t *testing.T) {
	orig := &Note{
		ID:          "04.16.2026",
		Body:        "some body\nwith multiple lines\n",
		CreatedAt:   time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 4, 16, 11, 0, 0, 0, time.UTC),
		Frontmatter: map[string]any{"tags": []any{"a", "b"}},
	}
	data, err := serializeNote(orig)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	got, err := parseNote(orig.ID, data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Body != orig.Body {
		t.Errorf("body mismatch:\ngot:  %q\nwant: %q", got.Body, orig.Body)
	}
	if !got.CreatedAt.Equal(orig.CreatedAt) {
		t.Errorf("created_at mismatch: %v vs %v", got.CreatedAt, orig.CreatedAt)
	}
	if !got.UpdatedAt.Equal(orig.UpdatedAt) {
		t.Errorf("updated_at mismatch: %v vs %v", got.UpdatedAt, orig.UpdatedAt)
	}
}

func TestSerializeNote_UnknownKeysSurvive(t *testing.T) {
	orig := &Note{
		ID:   "04.16.2026",
		Body: "body\n",
		Frontmatter: map[string]any{
			"aliases":  []any{"today"},
			"cssclass": "note",
		},
	}
	data, err := serializeNote(orig)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if !strings.Contains(string(data), "aliases") || !strings.Contains(string(data), "cssclass") {
		t.Errorf("unknown keys missing from output:\n%s", data)
	}
	parsed, err := parseNote(orig.ID, data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, ok := parsed.Frontmatter["aliases"]; !ok {
		t.Error("aliases lost in round-trip")
	}
	if _, ok := parsed.Frontmatter["cssclass"]; !ok {
		t.Error("cssclass lost in round-trip")
	}
}
