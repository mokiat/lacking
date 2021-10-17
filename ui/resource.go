package ui

import (
	"io"
	"io/fs"
)

// ResourceLocator represents a logic by which resources
// can be opened based off of a URI address.
//
// This allows resources (e.g. Images, Fonts) to be accessed
// from the filesystem, from the network, or from a custom
// bundle file.
type ResourceLocator interface {

	// OpenResource opens the resource at the specified
	// URI address.
	OpenResource(uri string) (io.ReadCloser, error)
}

// NewFileResourceLocator returns a new FileResourceLocator that is
// configured to search for resources relative to dir.
func NewFileResourceLocator(dir fs.FS) *FileResourceLocator {
	return &FileResourceLocator{
		dir: dir,
	}
}

// FileResourceLocator is an implementation of ResourceLocator that
// uses the local filesystem to open resources.
type FileResourceLocator struct {
	dir fs.FS
}

// OpenResource opens the resource at the specified
// relative address path.
func (l *FileResourceLocator) OpenResource(uri string) (io.ReadCloser, error) {
	return l.dir.Open(uri)
}
