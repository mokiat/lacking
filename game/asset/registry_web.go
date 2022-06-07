package asset

import (
	"fmt"
	"image"
	"image/png"
	"net/http"
	"path"

	"github.com/mokiat/lacking/log"
	"gopkg.in/yaml.v2"
)

// NewWebRegistry creates a Registry implementation that reads content
// from the web. The provided assetsURL parameter needs to point to the
// web location of the assets.
//
// This registry does not support write or delete operations.
func NewWebRegistry(assetsURL string) (Registry, error) {
	registry := &webRegistry{
		assetsURL:         assetsURL,
		previewURL:        path.Join(assetsURL, "preview"),
		contentURL:        path.Join(assetsURL, "content"),
		resourcesFromID:   make(map[string]*webResource),
		resourcesFromName: make(map[string]*webResource),
	}
	dtoSet, err := registry.fetchResources()
	if err != nil {
		return nil, fmt.Errorf("error loading resources: %w", err)
	}
	registry.resources = make([]*webResource, len(dtoSet.Resources))
	for i, dtoResource := range dtoSet.Resources {
		resource := &webResource{
			registry:     registry,
			id:           dtoResource.GUID,
			kind:         dtoResource.Kind,
			name:         dtoResource.Name,
			dependants:   make(map[string]struct{}),
			dependencies: make(map[string]struct{}),
		}
		registry.resources[i] = resource
		registry.resourcesFromID[resource.id] = resource
		registry.resourcesFromName[resource.name] = resource
	}
	for _, dtoDependency := range dtoSet.Dependencies {
		sourceResource := registry.resourcesFromID[dtoDependency.SourceGUID]
		targetResource := registry.resourcesFromID[dtoDependency.TargetGUID]
		if sourceResource == nil || targetResource == nil {
			log.Warn("[registry] Dangling dependency detected")
			continue
		}
		sourceResource.dependencies[targetResource.id] = struct{}{}
		targetResource.dependants[sourceResource.id] = struct{}{}
	}
	return registry, nil
}

var _ Registry = (*webRegistry)(nil)

type webRegistry struct {
	assetsURL         string
	previewURL        string
	contentURL        string
	resources         []*webResource
	resourcesFromID   map[string]*webResource
	resourcesFromName map[string]*webResource
}

func (r *webRegistry) Resources() []Resource {
	result := make([]Resource, len(r.resources))
	for i, resource := range r.resources {
		result[i] = resource
	}
	return result
}

func (r *webRegistry) ResourceByID(id string) Resource {
	if resource, ok := r.resourcesFromID[id]; ok {
		return resource
	}
	return nil
}

func (r *webRegistry) ResourceByName(name string) Resource {
	return r.resourcesFromName[name]
}

func (r *webRegistry) ResourcesByName(name string) []Resource {
	var result []Resource
	for _, resource := range r.resources {
		if resource.name == name {
			result = append(result, resource)
		}
	}
	return result
}

func (r *webRegistry) CreateResource(kind, name string) Resource {
	panic("not supported")
}

func (r *webRegistry) Save() error {
	panic("not supported")
}

func (r *webRegistry) fetchResources() (*resourcesDTO, error) {
	uri := path.Join(r.assetsURL, "resources.yml")
	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusOK:
		var resourcesIn resourcesDTO
		if err := yaml.NewDecoder(resp.Body).Decode(&resourcesIn); err != nil {
			return nil, fmt.Errorf("failed to decode resources: %w", err)
		}
		return &resourcesIn, nil
	default:
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
}

func (r *webRegistry) fetchPreview(guid string) (image.Image, error) {
	uri := path.Join(r.previewURL, fmt.Sprintf("%s.png", guid))
	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusOK:
		img, err := png.Decode(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode png image: %w", err)
		}
		return img, nil
	default:
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
}

func (r *webRegistry) fetchContent(guid string, target Decodable) error {
	uri := path.Join(r.contentURL, fmt.Sprintf("%s.dat", guid))
	resp, err := http.Get(uri)
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusOK:
		if err := target.DecodeFrom(resp.Body); err != nil {
			return fmt.Errorf("failed to decode content file: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
}

var _ Resource = (*webResource)(nil)

type webResource struct {
	registry     *webRegistry
	id           string
	kind         string
	name         string
	dependencies map[string]struct{}
	dependants   map[string]struct{}
}

func (r *webResource) ID() string {
	return r.id
}

func (r *webResource) Kind() string {
	return r.kind
}

func (r *webResource) Name() string {
	return r.name
}

func (r *webResource) SetName(name string) {
	panic("not supported")
}

func (r *webResource) Dependants() []Resource {
	var result []Resource
	for id := range r.dependants {
		if resource := r.registry.ResourceByID(id); resource != nil {
			result = append(result, resource)
		}
	}
	return result
}

func (r *webResource) Dependencies() []Resource {
	var result []Resource
	for id := range r.dependencies {
		if resource := r.registry.ResourceByID(id); resource != nil {
			result = append(result, resource)
		}
	}
	return result
}

func (r *webResource) AddDependency(resource Resource) {
	panic("not supported")
}

func (r *webResource) RemoveDependency(resource Resource) {
	panic("not supported")
}

func (r *webResource) ReadPreview() (image.Image, error) {
	return r.registry.fetchPreview(r.id)
}

func (r *webResource) WritePreview(image.Image) error {
	panic("not supported")
}

func (r *webResource) DeletePreview() error {
	panic("not supported")
}

func (r *webResource) ReadContent(target Decodable) error {
	return r.registry.fetchContent(r.id, target)
}

func (r *webResource) WriteContent(source Encodable) error {
	panic("not supported")
}

func (r *webResource) DeleteContent() error {
	panic("not supported")
}

func (r *webResource) Delete() {
	panic("not supported")
}
