package asset

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/mokiat/lacking/log"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

// NewDirRegistry creates a Registry implementation that stores content
// on the filesystem. The provided dir parameter needs to point to the
// project root. A special assets directory will be created inside if one
// is not available already.
func NewDirRegistry(dir string) (Registry, error) {
	dirFile, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat assets dir: %w", err)
	}
	if !dirFile.IsDir() {
		return nil, fmt.Errorf("file at %q is not a dir", dir)
	}

	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0775); err != nil {
		return nil, fmt.Errorf("failed to create assets dir: %w", err)
	}

	previewDir := filepath.Join(assetsDir, "preview")
	if err := os.MkdirAll(previewDir, 0775); err != nil {
		return nil, fmt.Errorf("failed to create preview dir: %w", err)
	}

	contentDir := filepath.Join(assetsDir, "content")
	if err := os.MkdirAll(contentDir, 0775); err != nil {
		return nil, fmt.Errorf("failed to create content dir: %w", err)
	}

	registry := &dirRegistry{
		dir:               assetsDir,
		previewDir:        previewDir,
		contentDir:        contentDir,
		resourcesFromID:   make(map[string]*dirResource),
		resourcesFromName: make(map[string]*dirResource),
	}

	dtoSet, err := registry.readResources()
	if err != nil {
		return nil, fmt.Errorf("error loading resources: %w", err)
	}
	registry.resources = make([]*dirResource, len(dtoSet.Resources))
	for i, dtoResource := range dtoSet.Resources {
		resource := &dirResource{
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

var _ Registry = (*dirRegistry)(nil)

type dirRegistry struct {
	dir        string
	previewDir string
	contentDir string

	resources         []*dirResource
	resourcesFromID   map[string]*dirResource
	resourcesFromName map[string]*dirResource
}

func (r *dirRegistry) Resources() []Resource {
	result := make([]Resource, len(r.resources))
	for i, resource := range r.resources {
		result[i] = resource
	}
	return result
}

func (r *dirRegistry) ResourceByID(id string) Resource {
	if resource, ok := r.resourcesFromID[id]; ok {
		return resource
	}
	return nil
}

func (r *dirRegistry) ResourceByName(name string) Resource {
	if resource, ok := r.resourcesFromName[name]; ok {
		return resource
	}
	for _, resource := range r.resources {
		if resource.name == name {
			return resource
		}
	}
	return nil
}

func (r *dirRegistry) ResourcesByName(name string) []Resource {
	var result []Resource
	for _, resource := range r.resources {
		if resource.name == name {
			result = append(result, resource)
		}
	}
	return result
}

func (r *dirRegistry) CreateResource(kind, name string) Resource {
	result := &dirResource{
		registry: r,
		id:       uuid.NewString(),
		kind:     kind,
		name:     name,
	}
	r.resources = append(r.resources, result)
	r.resourcesFromID[result.id] = result
	r.resourcesFromName[result.name] = result
	return result
}

func (r *dirRegistry) Save() error {
	return r.writeResources()
}

func (r *dirRegistry) readResources() (*resourcesDTO, error) {
	file, err := os.Open(r.resourcesFile())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &resourcesDTO{}, nil
		}
		return nil, fmt.Errorf("failed to open resources file: %w", err)
	}
	defer file.Close()

	var resourcesIn resourcesDTO
	if err := yaml.NewDecoder(file).Decode(&resourcesIn); err != nil {
		return nil, fmt.Errorf("failed to decode resources: %w", err)
	}
	return &resourcesIn, nil
}

func (r *dirRegistry) writeResources() error {
	file, err := os.Create(r.resourcesFile())
	if err != nil {
		return fmt.Errorf("failed to create resources file: %w", err)
	}
	defer file.Close()

	resourcesOut := &resourcesDTO{
		Resources:    make([]resourceDTO, len(r.resources)),
		Dependencies: make([]dependencyDTO, 0),
	}
	for i, resource := range r.resources {
		resourcesOut.Resources[i] = resourceDTO{
			GUID: resource.id,
			Kind: resource.kind,
			Name: resource.name,
		}
		for dependency := range resource.dependencies {
			resourcesOut.Dependencies = append(resourcesOut.Dependencies, dependencyDTO{
				SourceGUID: resource.id,
				TargetGUID: dependency,
			})
		}
	}
	if err := yaml.NewEncoder(file).Encode(resourcesOut); err != nil {
		return fmt.Errorf("failed to encode resources: %w", err)
	}
	return nil
}

func (r *dirRegistry) readPreview(guid string) (image.Image, error) {
	file, err := os.Open(r.previewFile(guid))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to open preview file: %w", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode png image: %w", err)
	}
	return img, nil
}

func (r *dirRegistry) writePreview(guid string, img image.Image) error {
	file, err := os.Create(r.previewFile(guid))
	if err != nil {
		return fmt.Errorf("failed to create preview file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode png image: %w", err)
	}
	return nil
}

func (r *dirRegistry) deletePreview(guid string) error {
	if err := os.Remove(r.previewFile(guid)); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete preview file: %w", err)
	}
	return nil
}

func (r *dirRegistry) readContent(guid string, target Decodable) error {
	file, err := os.Open(r.contentFile(guid))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to open content file: %w", err)
	}
	defer file.Close()

	if err := target.DecodeFrom(file); err != nil {
		return fmt.Errorf("failed to decode content file: %w", err)
	}
	return nil
}

func (r *dirRegistry) writeContent(guid string, target Encodable) error {
	file, err := os.Create(r.contentFile(guid))
	if err != nil {
		return fmt.Errorf("failed to create content file: %w", err)
	}
	defer file.Close()

	if err := target.EncodeTo(file); err != nil {
		return fmt.Errorf("failed to encode content: %w", err)
	}
	return nil
}

func (r *dirRegistry) deleteContent(guid string) error {
	if err := os.Remove(r.contentFile(guid)); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete content file: %w", err)
	}
	return nil
}

func (r *dirRegistry) resourcesFile() string {
	return filepath.Join(r.dir, "resources.yml")
}

func (r *dirRegistry) dependenciesFile() string {
	return filepath.Join(r.dir, "dependencies.yml")
}

func (r *dirRegistry) previewFile(guid string) string {
	return filepath.Join(r.previewDir, fmt.Sprintf("%s.png", guid))
}

func (r *dirRegistry) contentFile(guid string) string {
	return filepath.Join(r.contentDir, fmt.Sprintf("%s.dat", guid))
}

type dirResource struct {
	registry     *dirRegistry
	id           string
	kind         string
	name         string
	dependants   map[string]struct{}
	dependencies map[string]struct{}
}

func (r *dirResource) ID() string {
	return r.id
}

func (r *dirResource) Kind() string {
	return r.kind
}

func (r *dirResource) Name() string {
	return r.name
}

func (r *dirResource) SetName(name string) {
	delete(r.registry.resourcesFromName, name)
	r.name = name
	r.registry.resourcesFromName[name] = r
}

func (r *dirResource) Dependants() []Resource {
	var result []Resource
	for id := range r.dependants {
		if resource := r.registry.ResourceByID(id); resource != nil {
			result = append(result, resource)
		}
	}
	return result
}

func (r *dirResource) Dependencies() []Resource {
	var result []Resource
	for id := range r.dependencies {
		if resource := r.registry.ResourceByID(id); resource != nil {
			result = append(result, resource)
		}
	}
	return result
}

func (r *dirResource) AddDependency(resource Resource) {
	target := resource.(*dirResource)
	r.dependencies[target.id] = struct{}{}
	target.dependants[r.id] = struct{}{}
}

func (r *dirResource) RemoveDependency(resource Resource) {
	target := resource.(*dirResource)
	delete(r.dependencies, target.id)
	delete(target.dependants, r.id)
}

func (r *dirResource) ReadPreview() (image.Image, error) {
	return r.registry.readPreview(r.id)
}

func (r *dirResource) WritePreview(img image.Image) error {
	return r.registry.writePreview(r.id, img)
}

func (r *dirResource) DeletePreview() error {
	return r.registry.deletePreview(r.id)
}

func (r *dirResource) ReadContent(target Decodable) error {
	return r.registry.readContent(r.id, target)
}

func (r *dirResource) WriteContent(source Encodable) error {
	return r.registry.writeContent(r.id, source)
}

func (r *dirResource) DeleteContent() error {
	return r.registry.deleteContent(r.id)
}

func (r *dirResource) Delete() {
	for dependant := range r.dependants {
		source := r.registry.resourcesFromID[dependant]
		delete(source.dependencies, r.id)
	}
	for dependency := range r.dependencies {
		target := r.registry.resourcesFromID[dependency]
		delete(target.dependants, r.id)
	}
	delete(r.registry.resourcesFromID, r.id)
	delete(r.registry.resourcesFromName, r.name)
	if index := slices.Index(r.registry.resources, r); index >= 0 {
		r.registry.resources = slices.Delete(r.registry.resources, index, index+1)
	}
}
