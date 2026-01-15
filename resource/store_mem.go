package resource

import (
	"bytes"
	"io"
	"maps"
	"slices"

	"github.com/mokiat/lacking/util/ioutil"
)

// NewMemStore creates a new Store that uses memory.
func NewMemStore() Store {
	return &memStore{
		objects: make(map[string]*bytes.Buffer),
	}
}

type memStore struct {
	objects map[string]*bytes.Buffer
}

var _ Store = (*memStore)(nil)

func (s *memStore) Create(path string) (io.WriteCloser, error) {
	path = cleanFilePath(path)
	buffer := new(bytes.Buffer)
	s.objects[path] = buffer
	return ioutil.NopWriteCloser(buffer), nil
}

func (s *memStore) Open(path string) (io.ReadCloser, error) {
	path = cleanFilePath(path)
	buffer, ok := s.objects[path]
	if !ok {
		return nil, ErrNotFound
	}
	return io.NopCloser(buffer), nil
}

func (s *memStore) List() ([]string, error) {
	return slices.Collect(maps.Keys(s.objects)), nil
}

func (s *memStore) Delete(path string) error {
	path = cleanFilePath(path)
	if _, ok := s.objects[path]; !ok {
		return ErrNotFound
	}
	delete(s.objects, path)
	return nil
}
