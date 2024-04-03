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

// NewRegistry creates a new Registry that stores its as specified by the
// provided storage and formatter.
func NewRegistry(storage Storage, formatter Formatter) (*Registry, error) {
	registry := &Registry{
		storage:   storage,
		formatter: formatter,
	}
	if err := registry.open(); err != nil {
		return nil, fmt.Errorf("failed to open registry: %w", err)
	}
	return registry, nil
}

// Registry represents a managment interface for assets.
type Registry struct {
	storage   Storage
	formatter Formatter

	resources    []*Resource
	dependencies []fsDependency
}

// Resources returns a list of all resources in the registry.
func (r *Registry) Resources() []*Resource {
	return r.resources
}

// ResourceByID returns the resource with the specified ID.
// If no resource with the specified ID exists, nil is returned.
func (r *Registry) ResourceByID(id string) *Resource {
	resource, ok := gog.FindFunc(r.resources, func(res *Resource) bool {
		return res.id == id
	})
	if !ok {
		return nil
	}
	return resource
}

// ResourceByName returns the resource with the specified name.
// If no resource with the specified name exists, nil is returned.
func (r *Registry) ResourceByName(name string) *Resource {
	resource, ok := gog.FindFunc(r.resources, func(res *Resource) bool {
		return res.name == name
	})
	if !ok {
		return nil
	}
	return resource
}

// CreateResource creates a new resource with the specified name.
func (r *Registry) CreateResource(name string, content Model) (*Resource, error) {
	result := &Resource{
		registry: r,
		id:       uuid.NewString(),
		name:     name,
		preview:  nil,
	}
	r.resources = append(r.resources, result)
	if err := r.save(); err != nil {
		return nil, fmt.Errorf("failed to save registry: %w", err)
	}
	if err := r.saveContent(result.id, content); err != nil {
		return nil, fmt.Errorf("failed to save content: %w", err)
	}
	return result, nil
}

func (r *Registry) dependenciesOf(targetID string) []fsDependency {
	return gog.Select(r.dependencies, func(dep fsDependency) bool {
		return dep.TargetID == targetID
	})
}

func (r *Registry) dependentsOf(sourceID string) []fsDependency {
	return gog.Select(r.dependencies, func(dep fsDependency) bool {
		return dep.SourceID == sourceID
	})
}

func (r *Registry) open() error {
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

	r.resources = gog.Map(dtoRegistry.Resources, func(dto resourceDTO) *Resource {
		var preview image.Image
		if len(dto.PreviewData) > 0 {
			preview, err = png.Decode(bytes.NewReader(dto.PreviewData))
			if err != nil {
				log.Warn("Error decoding preview image: %v", err)
				preview = nil
			}
		}
		return &Resource{
			registry:     r,
			id:           dto.ID,
			name:         dto.Name,
			preview:      preview,
			sourceDigest: dto.SourceDigest,
		}
	})

	r.dependencies = gog.Map(dtoRegistry.Dependencies, func(dto dependencyDTO) fsDependency {
		return fsDependency(dto)
	})

	return nil
}

func (r *Registry) save() error {
	out, err := r.storage.OpenRegistryWrite()
	if err != nil {
		return fmt.Errorf("failed to create resources file: %w", err)
	}
	defer out.Close()

	dtoResources := gog.Map(r.resources, func(res *Resource) resourceDTO {
		var previewData bytes.Buffer
		if res.preview != nil {
			if err := png.Encode(&previewData, res.preview); err != nil {
				log.Warn("Error encoding preview image: %v", err)
				previewData.Reset()
			}
		}
		return resourceDTO{
			ID:           res.id,
			Name:         res.name,
			PreviewData:  previewData.Bytes(),
			SourceDigest: res.sourceDigest,
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

func (r *Registry) openContent(id string) (Model, error) {
	in, err := r.storage.OpenContentRead(id)
	if err != nil {
		return Model{}, fmt.Errorf("failed to open content file: %w", err)
	}
	defer in.Close()

	var content Model
	if err := r.formatter.Decode(in, &content); err != nil {
		return Model{}, fmt.Errorf("failed to decode content: %w", err)
	}
	return content, nil
}

func (r *Registry) saveContent(id string, content Model) error {
	newDependencies := gog.Map(content.ModelDefinitions, func(sourceID string) fsDependency {
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

func (r *Registry) deleteContent(id string) error {
	if err := r.storage.DeleteContent(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to delete content file: %w", err)
	}
	return nil
}

func (r *Registry) deleteResource(id string) error {
	r.resources = slices.DeleteFunc(r.resources, func(res *Resource) bool {
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

// Resource represents the generic aspects of an asset.
type Resource struct {
	registry     *Registry
	id           string
	name         string
	preview      image.Image
	sourceDigest string
}

// Registry returns the registry that manages the resource.
func (r *Resource) Registry() *Registry {
	return r.registry
}

// ID returns the unique identifier of the resource.
func (r *Resource) ID() string {
	return r.id
}

// Name returns the name of the resource.
func (r *Resource) Name() string {
	return r.name
}

// SetName changes the name of the resource. Two resources cannot have
// the same name.
func (r *Resource) SetName(name string) error {
	r.name = name
	return r.registry.save()
}

// Preview returns an image that represents the resource.
func (r *Resource) Preview() image.Image {
	return r.preview
}

// SetPreview changes the preview image of the resource.
func (r *Resource) SetPreview(preview image.Image) error {
	r.preview = preview
	return r.registry.save()
}

// SourceDigest returns the digest of the source content of the resource.
func (r *Resource) SourceDigest() string {
	return r.sourceDigest
}

// SetSourceDigest changes the digest of the source content of the resource.
func (r *Resource) SetSourceDigest(digest string) error {
	r.sourceDigest = digest
	return r.registry.save()
}

// OpenContent returns the content of the resource.
func (r *Resource) OpenContent() (Model, error) {
	return r.registry.openContent(r.id)
}

// SaveContent saves the content of the resource.
func (r *Resource) SaveContent(content Model) error {
	return r.registry.saveContent(r.id, content)
}

// Dependencies returns a list of resources that this resource depends on.
func (r *Resource) Dependencies() []*Resource {
	dependencies := r.registry.dependenciesOf(r.id)
	return gog.Map(dependencies, func(dep fsDependency) *Resource {
		return r.registry.ResourceByID(dep.SourceID)
	})
}

// Dependants returns a list of resources that depend on this resource.
func (r *Resource) Dependants() []*Resource {
	dependants := r.registry.dependentsOf(r.id)
	return gog.Map(dependants, func(dep fsDependency) *Resource {
		return r.registry.ResourceByID(dep.TargetID)
	})
}

// Delete removes the resource from the registry.
func (r *Resource) Delete() error {
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
