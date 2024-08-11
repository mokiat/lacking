package asset

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// Storage represents a storage interface for assets.
type Storage interface {

	// OpenRegistryRead opens a reader for the registry data.
	OpenRegistryRead() (io.ReadCloser, error)

	// OpenRegistryWrite opens a writer for the registry data.
	OpenRegistryWrite() (io.WriteCloser, error)

	// OpenPreviewRead opens a reader for the data of the specified resource.
	OpenContentRead(id string) (io.ReadCloser, error)

	// OpenPreviewWrite opens a writer for the data of the specified resource.
	OpenContentWrite(id string) (io.WriteCloser, error)

	// DeleteContent removes the data of the specified resource.
	DeleteContent(id string) error
}

// NewFSStorage creates a new storage that uses the file system.
func NewFSStorage(assetsDir string) (Storage, error) {
	if err := os.MkdirAll(assetsDir, 0775); err != nil {
		return nil, fmt.Errorf("error creating assets dir: %w", err)
	}

	contentDir := filepath.Join(assetsDir, "content")
	if err := os.MkdirAll(contentDir, 0775); err != nil {
		return nil, fmt.Errorf("error creating content dir: %w", err)
	}

	return &fsStorage{
		assetsDir:   assetsDir,
		contentsDir: contentDir,
	}, nil
}

type fsStorage struct {
	assetsDir   string
	contentsDir string
}

func (s *fsStorage) OpenRegistryRead() (io.ReadCloser, error) {
	return s.openFile(s.registryFile())
}

func (s *fsStorage) OpenRegistryWrite() (io.WriteCloser, error) {
	return s.createFile(s.registryFile())
}

func (s *fsStorage) OpenContentRead(id string) (io.ReadCloser, error) {
	return s.openFile(s.contentFile(id))
}

func (s *fsStorage) OpenContentWrite(id string) (io.WriteCloser, error) {
	return s.createFile(s.contentFile(id))
}

func (s *fsStorage) DeleteContent(id string) error {
	return s.deleteFile(s.contentFile(id))
}

func (s *fsStorage) registryFile() string {
	return filepath.Join(s.assetsDir, "resources.dat")
}

func (s *fsStorage) contentFile(id string) string {
	return filepath.Join(s.contentsDir, fmt.Sprintf("%s.dat", id))
}

func (s *fsStorage) openFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	return file, nil
}

func (s *fsStorage) createFile(path string) (*os.File, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}
	return file, nil
}

func (s *fsStorage) deleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrNotFound
		}
		return fmt.Errorf("error deleting file: %w", err)
	}
	return nil
}

// NewWebStorage creates a new storage that uses the web.
func NewWebStorage(assetsURL string) (Storage, error) {
	return &webStorage{
		assetsURL:   assetsURL,
		contentsURL: path.Join(assetsURL, "content"),
	}, nil
}

type webStorage struct {
	assetsURL   string
	contentsURL string
}

func (s *webStorage) OpenRegistryRead() (io.ReadCloser, error) {
	return s.fetch(s.resourcesURL())
}

func (s *webStorage) OpenRegistryWrite() (io.WriteCloser, error) {
	return nil, errors.ErrUnsupported
}

func (s *webStorage) OpenContentRead(id string) (io.ReadCloser, error) {
	return s.fetch(s.contentURL(id))
}

func (s *webStorage) OpenContentWrite(id string) (io.WriteCloser, error) {
	return nil, errors.ErrUnsupported
}

func (s *webStorage) DeleteContent(id string) error {
	return errors.ErrUnsupported
}

func (s *webStorage) resourcesURL() string {
	return path.Join(s.assetsURL, "resources.dat")
}

func (s *webStorage) contentURL(id string) string {
	return path.Join(s.contentsURL, fmt.Sprintf("%s.dat", id))
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
