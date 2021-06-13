package ui

import (
	"io"
	"os"
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

// FileResourceLocator is an implementation of ResourceLocator that
// uses the local filesystem to open resources.
type FileResourceLocator struct{}

func (l FileResourceLocator) OpenResource(uri string) (io.ReadCloser, error) {
	return os.Open(uri)
}
