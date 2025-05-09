package chunked

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"

	"github.com/mokiat/lacking/util/ioutil"
)

// ErrNotFound indicates that the specified content is not available.
var ErrNotFound = errors.New("not found")

// Storage represents a storage interface for assets.
type Storage interface {

	// Open opens a reader for the data of the specified asset.
	Open(path string) (io.ReadCloser, error)

	// Create opens a writer for the data of the specified asset.
	Create(path string) (io.WriteCloser, error)

	// Delete removes the specified asset.
	Delete(path string) error
}

// NewMemoryStorage creates a new storage that uses memory.
func NewMemoryStorage() Storage {
	return &memStorage{
		objects: make(map[string]*bytes.Buffer),
	}
}

type memStorage struct {
	objects map[string]*bytes.Buffer
}

func (s *memStorage) Open(path string) (io.ReadCloser, error) {
	path = cleanPath(path)
	buffer, ok := s.objects[path]
	if !ok {
		return nil, ErrNotFound
	}
	return io.NopCloser(buffer), nil
}

func (s *memStorage) Create(path string) (io.WriteCloser, error) {
	path = cleanPath(path)
	buffer := new(bytes.Buffer)
	s.objects[path] = buffer
	return ioutil.NopWriteCloser(buffer), nil
}

func (s *memStorage) Delete(path string) error {
	path = cleanPath(path)
	if _, ok := s.objects[path]; !ok {
		return ErrNotFound
	}
	delete(s.objects, path)
	return nil
}

func cleanPath(path string) string {
	return filepath.Clean(filepath.FromSlash(path))
}
