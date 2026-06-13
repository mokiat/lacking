package resource

import (
	"errors"
	"io"
	"io/fs"
)

// NewFSStore creates a new Store that uses an abstract filesystem.
func NewFSStore(fileSystem fs.FS) Store {
	return &fsStore{
		fileSystem: fileSystem,
	}
}

type fsStore struct {
	fileSystem fs.FS
}

var _ Store = (*fsStore)(nil)

func (s *fsStore) Create(path string) (io.WriteCloser, error) {
	return nil, errors.ErrUnsupported
}

func (s *fsStore) Open(path string) (io.ReadCloser, error) {
	in, err := s.fileSystem.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return in, err
}

func (s *fsStore) List() ([]string, error) {
	return nil, errors.ErrUnsupported
}

func (s *fsStore) Delete(path string) error {
	return errors.ErrUnsupported
}
