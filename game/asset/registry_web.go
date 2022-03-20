package asset

import (
	"fmt"
	"image"
	"image/png"
	"net/http"
	"path"

	"gopkg.in/yaml.v2"
)

// NewWebRegistry creates a Registry implementation that reads content
// from the web. The provided assetsURL parameter needs to point to the
// web location of the assets.
//
// This registry does not support write or delete operations.
func NewWebRegistry(assetsURL string) *WebRegistry {
	return &WebRegistry{
		assetsURL:  assetsURL,
		previewURL: path.Join(assetsURL, "preview"),
		contentURL: path.Join(assetsURL, "content"),
	}
}

var _ Registry = (*WebRegistry)(nil)

// WebRegistry is an implementation of Registry that reads content
// from the web.
type WebRegistry struct {
	assetsURL  string
	previewURL string
	contentURL string
}

func (r *WebRegistry) ReadResources() ([]Resource, error) {
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
		result := make([]Resource, len(resourcesIn.Resources))
		for i, resourceIn := range resourcesIn.Resources {
			result[i] = Resource(resourceIn)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
}

func (r *WebRegistry) WriteResources(resources []Resource) error {
	return fmt.Errorf("NOT SUPPORTED")
}

func (r *WebRegistry) ReadDependencies() ([]Dependency, error) {
	uri := path.Join(r.assetsURL, "dependencies.yml")
	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusOK:
		var dependenciesIn dependenciesDTO
		if err := yaml.NewDecoder(resp.Body).Decode(&dependenciesIn); err != nil {
			return nil, fmt.Errorf("failed to decode dependencies: %w", err)
		}
		result := make([]Dependency, len(dependenciesIn.Dependencies))
		for i, dependencyIn := range dependenciesIn.Dependencies {
			result[i] = Dependency(dependencyIn)
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
}

func (r *WebRegistry) WriteDependencies(dependencies []Dependency) error {
	return fmt.Errorf("NOT SUPPORTED")
}

func (r *WebRegistry) ReadPreview(guid string) (image.Image, error) {
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

func (r *WebRegistry) WritePreview(guid string, img image.Image) error {
	return fmt.Errorf("NOT SUPPORTED")
}

func (r *WebRegistry) DeletePreview(guid string) error {
	return fmt.Errorf("NOT SUPPORTED")
}

func (r *WebRegistry) ReadContent(guid string, target Decodable) error {
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

func (r *WebRegistry) WriteContent(guid string, target Encodable) error {
	return fmt.Errorf("NOT SUPPORTED")
}

func (r *WebRegistry) DeleteContent(guid string) error {
	return fmt.Errorf("NOT SUPPORTED")
}
