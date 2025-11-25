package chunked

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/http"
	"os"
	urlpath "path"
	"slices"

	"github.com/mokiat/lacking/util/ioutil"
)

// ErrNotFound indicates that the specified content is not available.
var ErrNotFound = errors.New("not found")

// Storage represents a storage interface for assets.
type Storage interface {

	// List returns all available assets.
	List() ([]string, error)

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

func (s *memStorage) List() ([]string, error) {
	return slices.Collect(maps.Keys(s.objects)), nil
}

func (s *memStorage) Open(path string) (io.ReadCloser, error) {
	path = cleanFilePath(path)
	buffer, ok := s.objects[path]
	if !ok {
		return nil, ErrNotFound
	}
	return io.NopCloser(buffer), nil
}

func (s *memStorage) Create(path string) (io.WriteCloser, error) {
	path = cleanFilePath(path)
	buffer := new(bytes.Buffer)
	s.objects[path] = buffer
	return ioutil.NopWriteCloser(buffer), nil
}

func (s *memStorage) Delete(path string) error {
	path = cleanFilePath(path)
	if _, ok := s.objects[path]; !ok {
		return ErrNotFound
	}
	delete(s.objects, path)
	return nil
}

// NewFileStorage creates a new storage that uses the file system.
func NewFileStorage(baseDir string) (Storage, error) {
	root, err := os.OpenRoot(baseDir)
	if err != nil {
		return nil, fmt.Errorf("error opening base dir: %w", err)
	}
	return &fileStorage{
		root: root,
	}, nil
}

type fileStorage struct {
	root *os.Root
}

func (s *fileStorage) List() ([]string, error) {
	var result []string
	err := fs.WalkDir(s.root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

func (s *fileStorage) Open(path string) (io.ReadCloser, error) {
	file, err := s.root.Open(cleanFilePath(path))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	return file, err
}

func (s *fileStorage) Create(path string) (io.WriteCloser, error) {
	return s.root.Create(cleanFilePath(path))
}

func (s *fileStorage) Delete(path string) error {
	err := s.root.Remove(cleanFilePath(path))
	if errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	}
	return err
}

// NewWebStorage creates a new storage that uses HTTP requests.
func NewWebStorage(baseURL string) (Storage, error) {
	return &webStorage{
		baseURL: baseURL,
	}, nil
}

type webStorage struct {
	baseURL string
}

func (s *webStorage) List() ([]string, error) {
	return nil, errors.ErrUnsupported
}

func (s *webStorage) Open(path string) (io.ReadCloser, error) {
	return s.fetch(urlpath.Join(s.baseURL, path))
}

func (s *webStorage) Create(path string) (io.WriteCloser, error) {
	return nil, errors.ErrUnsupported
}

func (s *webStorage) Delete(path string) error {
	return errors.ErrUnsupported
}

func (s *webStorage) fetch(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	if resp.StatusCode == http.StatusOK {
		return resp.Body, nil
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
}
