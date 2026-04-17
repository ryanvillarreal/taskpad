package notes

import (
	"sort"
	"time"
)

type Service struct {
	store *Store
	clock func() time.Time
}

func NewService(store *Store) *Service {
	return &Service{store: store, clock: time.Now}
}

func NewServiceWithClock(store *Store, clock func() time.Time) *Service {
	return &Service{store: store, clock: clock}
}

func (s *Service) Today() string {
	return s.clock().Format("01.02.2006")
}

func (s *Service) Save(id, rawBody string) (*Note, error) {
	now := s.clock().UTC()

	incoming, err := parseNote(id, []byte(rawBody))
	if err != nil {
		return nil, err
	}

	existing, err := s.store.Read(id)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	merged := &Note{
		ID:          id,
		Body:        incoming.Body,
		Frontmatter: incoming.Frontmatter,
		UpdatedAt:   now,
	}

	if existing != nil {
		merged.CreatedAt = existing.CreatedAt
		for k, v := range existing.Frontmatter {
			if _, present := merged.Frontmatter[k]; !present {
				merged.Frontmatter[k] = v
			}
		}
	}
	if merged.CreatedAt.IsZero() {
		merged.CreatedAt = extractTime(merged.Frontmatter["created_at"])
	}
	if merged.CreatedAt.IsZero() {
		merged.CreatedAt = now
	}

	if err := s.store.Write(merged); err != nil {
		return nil, err
	}
	return merged, nil
}

func (s *Service) Get(id string) (*Note, error) {
	return s.store.Read(id)
}

func (s *Service) Delete(id string) error {
	return s.store.Delete(id)
}

func (s *Service) Count() (int, error) {
	return s.store.Count()
}

func (s *Service) List() ([]string, error) {
	ids, err := s.store.List()
	if err != nil {
		return nil, err
	}
	sort.Slice(ids, func(i, j int) bool {
		ti, errI := time.Parse("01.02.2006", ids[i])
		tj, errJ := time.Parse("01.02.2006", ids[j])
		if errI != nil || errJ != nil {
			return ids[i] > ids[j]
		}
		return ti.After(tj)
	})
	return ids, nil
}

func (s *Service) Raw(id string) ([]byte, error) {
	n, err := s.store.Read(id)
	if err != nil {
		return nil, err
	}
	return serializeNote(n)
}
