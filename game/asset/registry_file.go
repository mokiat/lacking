package asset

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// NewDirRegistry creates a Registry implementation that stores content
// on the filesystem. The provided dir parameter needs to point to the
// project root. A special assets directory will be created inside if one
// is not available already.
func NewDirRegistry(dir string) (*DirRegistry, error) {
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

	return &DirRegistry{
		dir:        assetsDir,
		previewDir: previewDir,
		contentDir: contentDir,
	}, nil
}

var _ Registry = (*DirRegistry)(nil)

// DirRegistry is an implementation of Registry that stores content
// on the local filesystem.
type DirRegistry struct {
	dir        string
	previewDir string
	contentDir string
}

func (r *DirRegistry) ReadResources() ([]Resource, error) {
	file, err := os.Open(r.resourcesFile())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Resource{}, nil
		}
		return nil, fmt.Errorf("failed to open resources file: %w", err)
	}
	defer file.Close()

	var resourcesIn resourcesDTO
	if err := yaml.NewDecoder(file).Decode(&resourcesIn); err != nil {
		return nil, fmt.Errorf("failed to decode resources: %w", err)
	}

	result := make([]Resource, len(resourcesIn.Resources))
	for i, resourceIn := range resourcesIn.Resources {
		result[i] = Resource(resourceIn)
	}
	return result, nil
}

func (r *DirRegistry) WriteResources(resources []Resource) error {
	file, err := os.Create(r.resourcesFile())
	if err != nil {
		return fmt.Errorf("failed to create resources file: %w", err)
	}
	defer file.Close()

	resourcesOut := &resourcesDTO{
		Resources: make([]resourceDTO, len(resources)),
	}
	for i, resource := range resources {
		resourcesOut.Resources[i] = resourceDTO(resource)
	}

	if err := yaml.NewEncoder(file).Encode(resourcesOut); err != nil {
		return fmt.Errorf("failed to encode resources: %w", err)
	}
	return nil
}

func (r *DirRegistry) ReadDependencies() ([]Dependency, error) {
	file, err := os.Open(r.dependenciesFile())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Dependency{}, nil
		}
		return nil, fmt.Errorf("failed to open dependencies file: %w", err)
	}
	defer file.Close()

	var dependenciesIn dependenciesDTO
	if err := yaml.NewDecoder(file).Decode(&dependenciesIn); err != nil {
		return nil, fmt.Errorf("failed to decode dependencies: %w", err)
	}

	result := make([]Dependency, len(dependenciesIn.Dependencies))
	for i, dependencyIn := range dependenciesIn.Dependencies {
		result[i] = Dependency(dependencyIn)
	}
	return result, nil
}

func (r *DirRegistry) WriteDependencies(dependencies []Dependency) error {
	file, err := os.Create(r.dependenciesFile())
	if err != nil {
		return fmt.Errorf("failed to create dependencies file: %w", err)
	}
	defer file.Close()

	dependenciesOut := &dependenciesDTO{
		Dependencies: make([]dependencyDTO, len(dependencies)),
	}
	for i, dependency := range dependencies {
		dependenciesOut.Dependencies[i] = dependencyDTO(dependency)
	}

	if err := yaml.NewEncoder(file).Encode(dependenciesOut); err != nil {
		return fmt.Errorf("failed to encode dependencies: %w", err)
	}
	return nil
}

func (r *DirRegistry) ReadPreview(guid string) (image.Image, error) {
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

func (r *DirRegistry) WritePreview(guid string, img image.Image) error {
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

func (r *DirRegistry) DeletePreview(guid string) error {
	if err := os.Remove(r.previewFile(guid)); err != nil {
		return fmt.Errorf("failed to delete preview file: %w", err)
	}
	return nil
}

func (r *DirRegistry) ReadContent(guid string, target Decodable) error {
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

func (r *DirRegistry) WriteContent(guid string, target Encodable) error {
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

func (r *DirRegistry) DeleteContent(guid string) error {
	if err := os.Remove(r.contentFile(guid)); err != nil {
		return fmt.Errorf("failed to delete content file: %w", err)
	}
	return nil
}

func (r *DirRegistry) resourcesFile() string {
	return filepath.Join(r.dir, "resources.yml")
}

func (r *DirRegistry) dependenciesFile() string {
	return filepath.Join(r.dir, "dependencies.yml")
}

func (r *DirRegistry) previewFile(guid string) string {
	return filepath.Join(r.previewDir, fmt.Sprintf("%s.png", guid))
}

func (r *DirRegistry) contentFile(guid string) string {
	return filepath.Join(r.contentDir, fmt.Sprintf("%s.dat", guid))
}

type resourcesDTO struct {
	Resources []resourceDTO `yaml:"resources"`
}

type resourceDTO struct {
	GUID string `yaml:"guid"`
	Kind string `yaml:"kind"`
	Name string `yaml:"name"`
}

type dependenciesDTO struct {
	Dependencies []dependencyDTO `yaml:"dependencies"`
}

type dependencyDTO struct {
	SourceGUID string `yaml:"source_guid"`
	TargetGUID string `yaml:"target_guid"`
}
