package notes

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type SearchResult struct {
	Path    string
	Title   string
	Snippet string
	Score   int
}

func Search(notesDir, query string, limit int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	results := make([]SearchResult, 0)
	err := filepath.WalkDir(notesDir, func(path string, d fs.DirEntry, err error) error {
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

		text := string(content)
		score, snippet := matchScore(text, query)
		if score == 0 {
			return nil
		}

		results = append(results, SearchResult{
			Path:    path,
			Title:   extractTitle(path, text),
			Snippet: snippet,
			Score:   score,
		})
		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Title < results[j].Title
		}
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func View(path, heading string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	text := string(content)
	if heading == "" {
		return text, nil
	}
	return extractHeadingSection(text, heading)
}

func OpenInEditor(editor, path string) error {
	if strings.TrimSpace(editor) == "" {
		return fmt.Errorf("$EDITOR is not set")
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractHeadingSection(text, heading string) (string, error) {
	var (
		lines        = strings.Split(text, "\n")
		target       = strings.ToLower(strings.TrimSpace(heading))
		start        = -1
		currentLevel int
	)

	for i, line := range lines {
		level, name := parseHeading(line)
		if level == 0 {
			continue
		}
		if strings.ToLower(strings.TrimSpace(name)) == target {
			start = i
			currentLevel = level
			break
		}
	}
	if start == -1 {
		return "", fmt.Errorf("heading not found: %s", heading)
	}

	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		level, _ := parseHeading(lines[i])
		if level != 0 && level <= currentLevel {
			end = i
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines[start:end], "\n")) + "\n", nil
}

func parseHeading(line string) (int, string) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "#") {
		return 0, ""
	}
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	name := strings.TrimSpace(line[level:])
	if name == "" {
		return 0, ""
	}
	return level, name
}

func extractTitle(path, text string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return base
}

func matchScore(text, query string) (int, string) {
	queryLower := strings.ToLower(query)
	lines := strings.Split(text, "\n")
	score := 0
	snippet := ""

	for _, line := range lines {
		lineLower := strings.ToLower(line)
		if !strings.Contains(lineLower, queryLower) {
			continue
		}
		if snippet == "" {
			snippet = strings.TrimSpace(line)
		}
		score++
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			score += 3
		}
	}

	return score, snippet
}
