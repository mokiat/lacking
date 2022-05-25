package resource

import (
	"io"
	"io/fs"
)

// NewFSLocator returns a new instance of *FSLocator that uses the specified
// fs.FS to access resources.
func NewFSLocator(filesys fs.FS) *FSLocator {
	return &FSLocator{
		filesys: filesys,
	}
}

var _ ReadLocator = (*FSLocator)(nil)

// FSLocator is an implementation of ReadLocator that uses the fs.FS abstraction
// to load resources, allowing the API to be used with embedded files.
type FSLocator struct {
	filesys fs.FS
}

func (l *FSLocator) ReadResource(path string) (io.ReadCloser, error) {
	return l.filesys.Open(path)
}
