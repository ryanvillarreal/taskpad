package notes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchAndView(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bettercap.md")
	content := `---
taskpad_id: "abc"
---

# Bettercap docs

## cmdline options
use caplets
set net.sniff.verbose true

## cheatsheet
ble.recon on
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := Search(dir, "bettercap", 10)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Bettercap docs" {
		t.Fatalf("unexpected title: %q", results[0].Title)
	}

	section, err := View(path, "cmdline options")
	if err != nil {
		t.Fatalf("View() error = %v", err)
	}
	if want := "set net.sniff.verbose true"; !contains(section, want) {
		t.Fatalf("expected section to contain %q, got %q", want, section)
	}
	if contains(section, "ble.recon on") {
		t.Fatalf("section should not include later heading content: %q", section)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || filepath.Base(needle) == needle && (len(haystack) > 0 && stringContains(haystack, needle)))
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
