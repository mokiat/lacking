package asset

import (
	"bytes"
	"fmt"
	"io"

	"github.com/mokiat/lacking/util/resource"
)

// NewRegistryLocator returns a new *FileLocator that is configured to access
// resources located on the local filesystem.
func NewRegistryLocator(registry Registry) *RegistryLocator {
	return &RegistryLocator{
		registry: registry,
	}
}

var _ resource.ReadLocator = (*RegistryLocator)(nil)

// RegistryLocator is an implementation of ReadLocator that uses a Registry to
// access resources.
type RegistryLocator struct {
	registry Registry
}

func (l *RegistryLocator) ReadResource(path string) (io.ReadCloser, error) {
	resource := l.registry.ResourceByName(path)
	if resource == nil {
		return nil, fmt.Errorf("resource %q not found", path)
	}
	var blob Binary
	if err := resource.ReadContent(&blob); err != nil {
		return nil, fmt.Errorf("error reading binary data: %w", err)
	}
	return io.NopCloser(bytes.NewReader(blob.Data)), nil
}
