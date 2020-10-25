package pack

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ResourceLocator interface {
	Open(uri string) (io.ReadCloser, error)
}

type FileResourceLocator struct{}

func (FileResourceLocator) Open(uri string) (io.ReadCloser, error) {
	file, err := os.Open(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", uri, err)
	}
	return file, nil
}

type AssetLocator interface {
	Create(uri string) (io.WriteCloser, error)
}

type FileAssetLocator struct{}

func (FileAssetLocator) Create(uri string) (io.WriteCloser, error) {
	dirname := filepath.Dir(uri)
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return nil, fmt.Errorf("failed to create dir %q: %w", dirname, err)
	}

	file, err := os.Create(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %q: %w", uri, err)
	}
	return file, nil
}
