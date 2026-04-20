package links

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type linkFM struct {
	ID          string   `yaml:"id"`
	URL         string   `yaml:"url"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	SavedAt     string   `yaml:"saved_at"`
	Fetched     bool     `yaml:"fetched"`
}

func parseLink(id string, data []byte) (*Link, error) {
	l := &Link{ID: id}

	if !bytes.HasPrefix(data, []byte("---\n")) {
		l.URL = strings.TrimSpace(string(data))
		return l, nil
	}

	rest := data[4:]
	end := bytes.Index(rest, []byte("\n---\n"))
	if end < 0 {
		l.URL = strings.TrimSpace(string(data))
		return l, nil
	}

	var fm linkFM
	if err := yaml.Unmarshal(rest[:end], &fm); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	l.ID = fm.ID
	if l.ID == "" {
		l.ID = id
	}
	l.URL = fm.URL
	l.Title = fm.Title
	l.Description = fm.Description
	l.Tags = fm.Tags
	l.Fetched = fm.Fetched
	if fm.SavedAt != "" {
		if t, err := time.Parse(time.RFC3339, fm.SavedAt); err == nil {
			l.SavedAt = t
		}
	}

	return l, nil
}

func serializeLink(l *Link, notes, source string) ([]byte, error) {
	tags := l.Tags
	if tags == nil {
		tags = []string{}
	}

	fm := linkFM{
		ID:          l.ID,
		URL:         l.URL,
		Title:       l.Title,
		Description: l.Description,
		Tags:        tags,
		SavedAt:     l.SavedAt.UTC().Format(time.RFC3339),
		Fetched:     l.Fetched,
	}

	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(notes)
	if source != "" {
		if !strings.HasSuffix(notes, "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("\n```source\n")
		buf.WriteString(source)
		buf.WriteString("\n```\n")
	}

	return buf.Bytes(), nil
}
