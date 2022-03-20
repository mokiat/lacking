package asset

import (
	"fmt"
	"image"
	"io"
)

// ErrNotFound indicates that the specified content is not available.
var ErrNotFound = fmt.Errorf("not found")

// Encodable represents an asset that can be serialized.
type Encodable interface {
	EncodeTo(out io.Writer) error
}

// Decodable represents an asset that can be deserialized.
type Decodable interface {
	DecodeFrom(in io.Reader) error
}

// Registry represents a managment interface for assets.
type Registry interface {
	ReadResources() ([]Resource, error)
	WriteResources(resources []Resource) error

	ReadDependencies() ([]Dependency, error)
	WriteDependencies(dependencies []Dependency) error

	ReadPreview(guid string) (image.Image, error)
	WritePreview(guid string, img image.Image) error
	DeletePreview(guid string) error

	ReadContent(guid string, target Decodable) error
	WriteContent(guid string, target Encodable) error
	DeleteContent(guid string) error
}

// Resource represents the generic aspects of an asset.
type Resource struct {
	GUID string
	Kind string
	Name string
}

// Dependency describes the dependency of a source asset to
// a target asset.
type Dependency struct {
	SourceGUID string
	TargetGUID string
}
