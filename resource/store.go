package resource

import "io"

// Store represents a resource store that can manage multiple resources.
type Store interface {

	// Create opens a writer for the data of the specified asset.
	//
	// If the operation is not supported, an errors.ErrUnsupported is returned.
	Create(path string) (io.WriteCloser, error)

	// Open opens a reader for the data of the specified asset.
	Open(path string) (io.ReadCloser, error)

	// List returns all available assets.
	//
	// If the operation is not supported, an errors.ErrUnsupported is returned.
	List() ([]string, error)

	// Delete removes the specified asset.
	//
	// If the operation is not supported, an errors.ErrUnsupported is returned.
	Delete(path string) error
}
