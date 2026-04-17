package notes

import (
	"bytes"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

func parseNote(id string, data []byte) (*Note, error) {
	n := &Note{ID: id, Frontmatter: map[string]any{}}

	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		n.Body = string(data)
		return n, nil
	}

	rest := data[4:]
	end := bytes.Index(rest, []byte("\n---\n"))
	if end < 0 {
		end = bytes.Index(rest, []byte("\n---\r\n"))
	}
	if end < 0 {
		n.Body = string(data)
		return n, nil
	}
	fm := rest[:end]
	body := rest[end:]
	body = bytes.TrimPrefix(body, []byte("\n---\n"))
	body = bytes.TrimPrefix(body, []byte("\n---\r\n"))
	body = bytes.TrimLeft(body, "\n")

	if err := yaml.Unmarshal(fm, &n.Frontmatter); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	if n.Frontmatter == nil {
		n.Frontmatter = map[string]any{}
	}

	n.CreatedAt = extractTime(n.Frontmatter["created_at"])
	n.UpdatedAt = extractTime(n.Frontmatter["updated_at"])

	n.Body = string(body)
	return n, nil
}

func extractTime(v any) time.Time {
	switch x := v.(type) {
	case time.Time:
		return x
	case string:
		if t, err := time.Parse(time.RFC3339, x); err == nil {
			return t
		}
	}
	return time.Time{}
}

func serializeNote(n *Note) ([]byte, error) {
	fm := n.Frontmatter
	if fm == nil {
		fm = map[string]any{}
	}
	fm["id"] = n.ID
	if !n.CreatedAt.IsZero() {
		fm["created_at"] = n.CreatedAt.UTC().Format(time.RFC3339)
	}
	if !n.UpdatedAt.IsZero() {
		fm["updated_at"] = n.UpdatedAt.UTC().Format(time.RFC3339)
	}

	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(n.Body)
	return buf.Bytes(), nil
}
