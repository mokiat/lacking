package resource

import (
	"io"
	"os"
)

type Locator interface {
	Open(uri string) (io.ReadCloser, error)
}

type FileLocator struct{}

func (FileLocator) Open(uri string) (io.ReadCloser, error) {
	return os.Open(uri)
}
