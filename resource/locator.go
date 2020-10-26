package resource

import (
	"io"
	"os"
	"path/filepath"
)

type Locator interface {
	Open(segments ...string) (io.ReadCloser, error)
}

type FileLocator struct{}

func (FileLocator) Open(segments ...string) (io.ReadCloser, error) {
	path := filepath.Join(segments...) + ".dat"
	return os.Open(path)
}
