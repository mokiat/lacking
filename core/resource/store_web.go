package resource

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// NewWebStore creates a new Store that uses HTTP requests.
func NewWebStore(baseURL string) Store {
	return &webStore{
		baseURL: baseURL,
	}
}

type webStore struct {
	baseURL string
}

var _ Store = (*webStore)(nil)

func (s *webStore) Create(path string) (io.WriteCloser, error) {
	return nil, errors.ErrUnsupported
}

func (s *webStore) Open(path string) (io.ReadCloser, error) {
	baseURL := strings.TrimSuffix(s.baseURL, "/")
	return s.fetch(fmt.Sprintf("%s/%s", baseURL, path))
}

func (s *webStore) List() ([]string, error) {
	return nil, errors.ErrUnsupported
}

func (s *webStore) Delete(path string) error {
	return errors.ErrUnsupported
}

func (s *webStore) fetch(url string) (io.ReadCloser, error) {
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
