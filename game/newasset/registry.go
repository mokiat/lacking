package asset

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"slices"

	"github.com/google/uuid"
	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/debug/log"
)

// ErrNotFound indicates that the specified content is not available.
var ErrNotFound = fs.ErrNotExist

// Registry represents a managment interface for assets.
type Registry interface {

	// Resources returns a list of all resources in the registry.
	Resources() []Resource

	// ResourceByID returns the resource with the specified ID.
	// If no resource with the specified ID exists, nil is returned.
	ResourceByID(id string) Resource

	// ResourceByName returns the resource with the specified name.
	// If no resource with the specified name exists, nil is returned.
	ResourceByName(name string) Resource

	// CreateResource creates a new resource with the specified name.
	CreateResource(name string) (Resource, error)
}

// Resource represents the generic aspects of an asset.
type Resource interface {

	// ID returns the unique identifier of the resource.
	ID() string

	// Dependencies returns a list of resources that this resource depends on.
	Dependencies() []Resource

	// Dependants returns a list of resources that depend on this resource.
	Dependants() []Resource

	// Name returns the name of the resource.
	Name() string

	// SetName changes the name of the resource. Two resources cannot have
	// the same name.
	SetName(newName string) error

	// Preview returns an image that represents the resource.
	Preview() image.Image

	// SetPreview changes the preview image of the resource.
	SetPreview(image.Image) error

	// OpenContent returns the content of the resource.
	OpenContent() (Fragment, error)

	// SaveContent saves the content of the resource.
	SaveContent(Fragment) error

	// Delete removes the resource from the registry.
	Delete() error
}

// NewRegistry creates a new Registry that stores its as specified by the
// provided storage and formatter.
func NewRegistry(storage Storage, formatter Formatter) (Registry, error) {
	registry := &fsRegistry{
		storage:   storage,
		formatter: formatter,
	}
	if err := registry.open(); err != nil {
		return nil, fmt.Errorf("failed to open registry: %w", err)
	}
	return registry, nil
}

type fsRegistry struct {
	storage   Storage
	formatter Formatter

	resources    []*fsResource
	dependencies []fsDependency
}

func (r *fsRegistry) Resources() []Resource {
	return gog.Map(r.resources, func(res *fsResource) Resource {
		return res
	})
}

func (r *fsRegistry) ResourceByID(id string) Resource {
	resource, ok := gog.FindFunc(r.resources, func(res *fsResource) bool {
		return res.id == id
	})
	if !ok {
		return nil
	}
	return resource
}

func (r *fsRegistry) ResourceByName(name string) Resource {
	resource, ok := gog.FindFunc(r.resources, func(res *fsResource) bool {
		return res.name == name
	})
	if !ok {
		return nil
	}
	return resource
}

func (r *fsRegistry) CreateResource(name string) (Resource, error) {
	result := &fsResource{
		registry: r,
		id:       uuid.NewString(),
		name:     name,
		preview:  nil,
	}
	r.resources = append(r.resources, result)
	if err := r.save(); err != nil {
		return nil, fmt.Errorf("failed to save registry: %w", err)
	}
	return result, nil
}

func (r *fsRegistry) dependenciesOf(targetID string) []fsDependency {
	return gog.Select(r.dependencies, func(dep fsDependency) bool {
		return dep.TargetID == targetID
	})
}

func (r *fsRegistry) dependentsOf(sourceID string) []fsDependency {
	return gog.Select(r.dependencies, func(dep fsDependency) bool {
		return dep.SourceID == sourceID
	})
}

func (r *fsRegistry) open() error {
	in, err := r.storage.OpenRegistryRead()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to open registry file: %w", err)
	}
	defer in.Close()

	var dtoRegistry registryDTO
	if err := r.formatter.Decode(in, &dtoRegistry); err != nil {
		return fmt.Errorf("failed to decode registry: %w", err)
	}

	r.resources = gog.Map(dtoRegistry.Resources, func(dto resourceDTO) *fsResource {
		preview, err := png.Decode(bytes.NewReader(dto.PreviewData))
		if err != nil {
			log.Warn("Error decoding preview image: %v", err)
			preview = nil
		}
		return &fsResource{
			registry: r,
			id:       dto.ID,
			name:     dto.Name,
			preview:  preview,
		}
	})

	r.dependencies = gog.Map(dtoRegistry.Dependencies, func(dto dependencyDTO) fsDependency {
		return fsDependency(dto)
	})

	return nil
}

func (r *fsRegistry) save() error {
	out, err := r.storage.OpenRegistryWrite()
	if err != nil {
		return fmt.Errorf("failed to create resources file: %w", err)
	}
	defer out.Close()

	dtoResources := gog.Map(r.resources, func(res *fsResource) resourceDTO {
		var previewData bytes.Buffer
		if err := png.Encode(&previewData, res.preview); err != nil {
			log.Warn("Error encoding preview image: %v", err)
			previewData.Reset()
		}
		return resourceDTO{
			ID:          res.id,
			Name:        res.name,
			PreviewData: previewData.Bytes(),
		}
	})
	dtoDependencies := gog.Map(r.dependencies, func(dep fsDependency) dependencyDTO {
		return dependencyDTO(dep)
	})
	dtoRegistry := registryDTO{
		Resources:    dtoResources,
		Dependencies: dtoDependencies,
	}

	if err := r.formatter.Encode(out, dtoRegistry); err != nil {
		return fmt.Errorf("failed to encode registry: %w", err)
	}
	return nil
}

func (r *fsRegistry) openContent(id string) (Fragment, error) {
	in, err := r.storage.OpenContentRead(id)
	if err != nil {
		return Fragment{}, fmt.Errorf("failed to open content file: %w", err)
	}
	defer in.Close()

	var fragment Fragment
	if err := r.formatter.Decode(in, &fragment); err != nil {
		return Fragment{}, fmt.Errorf("failed to decode content: %w", err)
	}
	return fragment, nil
}

func (r *fsRegistry) saveContent(id string, content Fragment) error {
	newDependencies := gog.Map(content.Dependencies, func(sourceID string) fsDependency {
		return fsDependency{
			TargetID: id,
			SourceID: sourceID,
		}
	})
	r.dependencies = slices.DeleteFunc(r.dependencies, func(dep fsDependency) bool {
		return dep.TargetID == id
	})
	r.dependencies = append(r.dependencies, newDependencies...)

	if err := r.save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	out, err := r.storage.OpenContentWrite(id)
	if err != nil {
		return fmt.Errorf("failed to create content file: %w", err)
	}
	defer out.Close()

	if err := r.formatter.Encode(out, content); err != nil {
		return fmt.Errorf("failed to encode content: %w", err)
	}
	return nil
}

func (r *fsRegistry) deleteContent(id string) error {
	if err := r.storage.DeleteContent(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to delete content file: %w", err)
	}
	return nil
}

func (r *fsRegistry) deleteResource(id string) error {
	r.resources = slices.DeleteFunc(r.resources, func(res *fsResource) bool {
		return res.id == id
	})
	r.dependencies = slices.DeleteFunc(r.dependencies, func(dep fsDependency) bool {
		return dep.TargetID == id || dep.SourceID == id
	})
	if err := r.save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}
	return nil
}

type fsResource struct {
	registry *fsRegistry
	id       string
	name     string
	preview  image.Image
}

func (r *fsResource) ID() string {
	return r.id
}

func (r *fsResource) Name() string {
	return r.name
}

func (r *fsResource) SetName(name string) error {
	r.name = name
	return r.registry.save()
}

func (r *fsResource) Preview() image.Image {
	return r.preview
}

func (r *fsResource) SetPreview(preview image.Image) error {
	r.preview = preview
	return r.registry.save()
}

func (r *fsResource) OpenContent() (Fragment, error) {
	return r.registry.openContent(r.id)
}

func (r *fsResource) SaveContent(content Fragment) error {
	return r.registry.saveContent(r.id, content)
}

func (r *fsResource) Dependencies() []Resource {
	dependencies := r.registry.dependenciesOf(r.id)
	return gog.Map(dependencies, func(dep fsDependency) Resource {
		return r.registry.ResourceByID(dep.SourceID)
	})
}

func (r *fsResource) Dependants() []Resource {
	dependants := r.registry.dependentsOf(r.id)
	return gog.Map(dependants, func(dep fsDependency) Resource {
		return r.registry.ResourceByID(dep.TargetID)
	})
}

func (r *fsResource) Delete() error {
	if err := r.registry.deleteContent(r.id); err != nil {
		return fmt.Errorf("failed to delete content: %w", err)
	}
	if err := r.registry.deleteResource(r.id); err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}
	return nil
}

type fsDependency struct {
	TargetID string
	SourceID string
}
