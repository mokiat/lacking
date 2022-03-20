package pack

import (
	"fmt"
	"io"
	"os"
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
