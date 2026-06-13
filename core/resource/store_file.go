package resource

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// NewFileStore creates a new Store that uses the file system.
func NewFileStore(baseDir string) (Store, error) {
	root, err := os.OpenRoot(baseDir)
	if err != nil {
		return nil, fmt.Errorf("error opening base dir: %w", err)
	}
	return &fileStore{
		root: root,
	}, nil
}

type fileStore struct {
	root *os.Root
}

var _ Store = (*fileStore)(nil)

func (s *fileStore) Create(path string) (io.WriteCloser, error) {
	return s.root.Create(cleanFilePath(path))
}

func (s *fileStore) Open(path string) (io.ReadCloser, error) {
	file, err := s.root.Open(cleanFilePath(path))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	return file, err
}

func (s *fileStore) List() ([]string, error) {
	var result []string
	err := fs.WalkDir(s.root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

func (s *fileStore) Delete(path string) error {
	err := s.root.Remove(cleanFilePath(path))
	if errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	}
	return err
}
