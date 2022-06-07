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
	Resources() []Resource
	ResourceByID(id string) Resource
	ResourceByName(name string) Resource
	ResourcesByName(name string) []Resource
	CreateResource(kind, name string) Resource
	Save() error
}

// Resource represents the generic aspects of an asset.
type Resource interface {
	ID() string
	Kind() string
	Name() string
	SetName(name string)
	Dependants() []Resource
	Dependencies() []Resource
	AddDependency(resource Resource)
	RemoveDependency(resource Resource)
	ReadPreview() (image.Image, error)
	WritePreview(image.Image) error
	DeletePreview() error
	ReadContent(target Decodable) error
	WriteContent(source Encodable) error
	DeleteContent() error
	Delete()
}
