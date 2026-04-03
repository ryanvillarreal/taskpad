package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DailyNoteDir returns the directory where daily notes are stored.
func DailyNoteDir(notesDir string) string {
	return filepath.Join(notesDir, "daily")
}

// DailyNoteFilePath returns the full path for a given date's note file.
func DailyNoteFilePath(notesDir, date string) string {
	return filepath.Join(DailyNoteDir(notesDir), date+".md")
}

// WriteLocalDailyNote renders and writes a daily note file to disk.
// The content parameter is just the body text (no frontmatter or heading).
// Returns the path written.
func WriteLocalDailyNote(notesDir, date, content string) (string, error) {
	dir := DailyNoteDir(notesDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create daily notes directory: %w", err)
	}

	path := DailyNoteFilePath(notesDir, date)

	// Preserve the original added date if the file exists.
	added := date
	if existing, err := os.ReadFile(path); err == nil {
		if a := parseFrontmatterField(string(existing), "added"); a != "" {
			added = a
		}
	}

	rendered := renderDailyNote(date, added, content)
	if err := os.WriteFile(path, []byte(rendered), 0o644); err != nil {
		return "", fmt.Errorf("write daily note: %w", err)
	}
	return path, nil
}

// ParseDailyNoteBody reads a daily note file and returns only the body text
// (everything after the frontmatter and the H1 heading line).
func ParseDailyNoteBody(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read daily note: %w", err)
	}
	return extractBody(string(raw)), nil
}

// renderDailyNote produces the full Markdown file content.
func renderDailyNote(date, added, content string) string {
	edited := time.Now().UTC().Format(time.RFC3339)

	// Parse the date for a human-readable heading.
	heading := date
	if t, err := time.Parse("2006-01-02", date); err == nil {
		heading = t.Format("January 2, 2006")
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("added: %s\n", added))
	b.WriteString(fmt.Sprintf("edited: %s\n", edited))
	b.WriteString("---\n\n")
	b.WriteString(fmt.Sprintf("# Daily Note - %s\n\n", heading))
	if strings.TrimSpace(content) != "" {
		body := strings.TrimSpace(content)
		b.WriteString(body)
		b.WriteString("\n")
	}
	return b.String()
}

// extractBody strips the YAML frontmatter and H1 heading, returning the rest.
func extractBody(text string) string {
	lines := strings.Split(text, "\n")
	i := 0

	// Skip frontmatter block (--- ... ---).
	if i < len(lines) && strings.TrimSpace(lines[i]) == "---" {
		i++
		for i < len(lines) {
			if strings.TrimSpace(lines[i]) == "---" {
				i++
				break
			}
			i++
		}
	}

	// Skip blank lines after frontmatter.
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	// Skip the H1 heading line.
	if i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "# ") {
		i++
	}

	// Skip blank lines after heading.
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	body := strings.Join(lines[i:], "\n")
	return strings.TrimSpace(body)
}

// parseFrontmatterField extracts a simple scalar value from YAML frontmatter.
func parseFrontmatterField(text, key string) string {
	prefix := key + ": "
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}
