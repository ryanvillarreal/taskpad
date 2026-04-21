package notes

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Store struct {
	dir string
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) path(id string) string {
	t, err := time.Parse("01.02.2006", id)
	if err != nil {
		return filepath.Join(s.dir, id+".md")
	}
	return filepath.Join(s.dir, t.Format("06"), t.Format("01"), id+".md")
}

func (s *Store) Read(id string) (*Note, error) {
	data, err := os.ReadFile(s.path(id))
	if os.IsNotExist(err) {
		flat := filepath.Join(s.dir, id+".md")
		data, err = os.ReadFile(flat)
	}
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
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}

func (s *Store) Delete(id string) error {
	err := os.Remove(s.path(id))
	if os.IsNotExist(err) {
		err = os.Remove(filepath.Join(s.dir, id+".md"))
	}
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Store) List() ([]string, error) {
	var ids []string
	err := filepath.WalkDir(s.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, ".md") {
			ids = append(ids, strings.TrimSuffix(name, ".md"))
		}
		return nil
	})
	return ids, err
}

func (s *Store) Count() (int, error) {
	n := 0
	err := filepath.WalkDir(s.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			n++
		}
		return nil
	})
	return n, err
}
