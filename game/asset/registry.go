package asset

import (
	"image"
)

type Registry interface {
	ReadResources() ([]Resource, error)
	WriteResources(resources []Resource) error

	ReadDependencies() ([]Dependency, error)
	WriteDependencies(dependencies []Dependency) error

	ReadPreview(guid string) (image.Image, error)
	WritePreview(guid string, img image.Image) error

	ReadContent(guid string, target interface{}) error
	WriteContent(guid string, target interface{}) error
}

type Resource struct {
	GUID string
	Kind string
	Name string
}

type Dependency struct {
	SourceGUID string
	TargetGUID string
}
