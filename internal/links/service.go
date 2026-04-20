package links

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
)

type Service struct {
	store      *Store
	clock      func() time.Time
	fetchLimit int
}

func NewService(store *Store, fetchLimit int) *Service {
	return &Service{store: store, clock: time.Now, fetchLimit: fetchLimit}
}

func (s *Service) Create(rawURL string, fetch bool) (*Link, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	l := &Link{
		ID:      id,
		URL:     rawURL,
		SavedAt: s.clock().UTC(),
	}

	var source string

	if fetch {
		article, err := readability.FromURL(rawURL, 15*time.Second)
		if err == nil {
			l.Title = article.Title()
			l.Description = article.Excerpt()
			l.Fetched = true
			var buf bytes.Buffer
			if err := article.RenderText(&buf); err == nil {
				source = truncate(buf.String(), s.fetchLimit)
			}
		}
	}

	if l.Title == "" {
		l.Title = rawURL
	}

	if err := s.store.Write(l, "", source); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Service) Get(id string) (*Link, error) {
	return s.store.Read(id)
}

func (s *Service) List() ([]*Link, error) {
	links, err := s.store.List()
	if err != nil {
		return nil, err
	}
	sort.Slice(links, func(i, j int) bool {
		return links[i].SavedAt.After(links[j].SavedAt)
	})
	return links, nil
}

func (s *Service) Resolve(prefix string) (string, error) {
	all, err := s.store.List()
	if err != nil {
		return "", err
	}
	var matches []string
	for _, l := range all {
		if strings.HasPrefix(l.ID, prefix) {
			matches = append(matches, l.ID)
		}
	}
	switch len(matches) {
	case 0:
		return "", ErrNotFound
	case 1:
		return matches[0], nil
	default:
		return "", ErrAmbiguous
	}
}

func (s *Service) Delete(id string) error {
	return s.store.Delete(id)
}

func truncate(s string, limit int) string {
	s = strings.TrimSpace(s)
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "\n[truncated at " + itoa(limit) + " chars]"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}

func newID() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
