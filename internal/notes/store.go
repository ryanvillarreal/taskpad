package notes

import (
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	dir string
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) path(id string) string {
	return filepath.Join(s.dir, id+".md")
}

func (s *Store) Read(id string) (*Note, error) {
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return parseNote(id, data)
}

func (s *Store) Write(n *Note) error {
	data, err := serializeNote(n)
	if err != nil {
		return err
	}
	target := s.path(n.ID)
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}

func (s *Store) Delete(id string) error {
	err := os.Remove(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Store) Count() (int, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			n++
		}
	}
	return n, nil
}

