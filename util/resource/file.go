package resource

import (
	"io"
	"os"
	"path/filepath"
)

// NewFileLocator returns a new *FileLocator that is configured to access
// resources located on the local filesystem.
func NewFileLocator(dir string) *FileLocator {
	return &FileLocator{
		dir: dir,
	}
}

var _ ReadLocator = (*FileLocator)(nil)
var _ WriteLocator = (*FileLocator)(nil)

// FileLocator is an implementation of ReadLocator and WriteLocator that uses
// the local filesystem to access resources.
type FileLocator struct {
	dir string
}

func (l *FileLocator) ReadResource(path string) (io.ReadCloser, error) {
	path = filepath.FromSlash(path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(l.dir, path)
	}
	return os.Open(path)
}

func (l *FileLocator) WriteResource(path string) (io.WriteCloser, error) {
	path = filepath.FromSlash(path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(l.dir, path)
	}
	return os.Create(path)
}
