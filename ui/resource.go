package ui

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
func NewFileResourceLocator(dir string) *FileResourceLocator {
	return &FileResourceLocator{
		dir: dir,
	}
}

// FileResourceLocator is an implementation of ResourceLocator that
// uses the local filesystem to open resources.
type FileResourceLocator struct {
	dir string
}

// OpenResource opens the resource at the specified relative URI path.
func (l *FileResourceLocator) OpenResource(uri string) (io.ReadCloser, error) {
	if !filepath.IsAbs(uri) {
		uri = filepath.Join(l.dir, uri)
	}
	return os.Open(uri)
}

// NewFSResourceLocator returns a new instance of FSResourceLocator that
// uses the specified fs.FS to load resources.
func NewFSResourceLocator(filesys fs.FS) *FSResourceLocator {
	return &FSResourceLocator{
		filesys: filesys,
	}
}

// FileResourceLocator is an implementation of ResourceLocator that
// uses the fs.FS abstraction to load resources. This allows the
// API to be used with embedded files, for example.
type FSResourceLocator struct {
	filesys fs.FS
}

// OpenResource opens the resource at the specified URI location.
func (l *FSResourceLocator) OpenResource(uri string) (io.ReadCloser, error) {
	return l.filesys.Open(uri)
}
