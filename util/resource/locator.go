package resource

import "io"

// ReadLocator represents a logic by which resources can be opened for reading
// based off of a path.
type ReadLocator interface {

	// ReadResource opens the resource at the specified path for reading.
	ReadResource(path string) (io.ReadCloser, error)
}

// WriteLocator represents a logic by which resources can be opened for writing
// based off of a path.
type WriteLocator interface {

	// WriteResource opens the resource at the specified path for writing.
	WriteResource(path string) (io.WriteCloser, error)
}
