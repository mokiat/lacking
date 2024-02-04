package asset

import "io"

// Storage represents a storage interface for assets.
type Storage interface {

	// OpenRegistryRead opens a reader for the registry data.
	OpenRegistryRead() (io.ReadCloser, error)

	// OpenRegistryWrite opens a writer for the registry data.
	OpenRegistryWrite() (io.WriteCloser, error)

	// OpenPreviewRead opens a reader for the data of the specified resource.
	OpenContentRead(id string) (io.ReadCloser, error)

	// OpenPreviewWrite opens a writer for the data of the specified resource.
	OpenContentWrite(id string) (io.WriteCloser, error)

	// DeleteContent removes the data of the specified resource.
	DeleteContent(id string) error
}

// TODO: FS storage implementation

// TODO: Web storage implementation
