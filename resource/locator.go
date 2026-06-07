package resource

import (
	"errors"
	"io"
	"strings"
)

// ErrNotFound indicates that the specified content is not available.
var ErrNotFound = errors.New("not found")

// Locator represents a logic by which resources can be opened for reading
// based off of a path.
type Locator interface {

	// Open opens the resource at the specified path for reading.
	Open(path string) (io.ReadCloser, error)
}

// LocatorFunc is a function type that implements the Locator interface.
type LocatorFunc func(path string) (io.ReadCloser, error)

// Open opens the resource at the specified path for reading.
func (f LocatorFunc) Open(path string) (io.ReadCloser, error) {
	return f(path)
}

// OneOfLocator creates a Locator that tries each of the specified locators
// in order until one is able to successfully open the resource at the given
// path.
//
// If none of the locators are able to find the resource, an ErrNotFound
// error is returned.
//
// If any locator returns an error other than ErrNotFound, that error is
// returned immediately.
func OneOfLocator(locators ...Locator) Locator {
	return LocatorFunc(func(path string) (io.ReadCloser, error) {
		for _, locator := range locators {
			in, err := locator.Open(path)
			if errors.Is(err, ErrNotFound) {
				continue // try next locator
			}
			return in, err
		}
		return nil, ErrNotFound
	})
}

// SchemaLocator creates a Locator that delegates to the specified locator
// only if the path begins with the specified schema followed by ":///".
//
// If the path does not match the schema, an ErrNotFound error is returned.
func SchemaLocator(schema string, locator Locator) Locator {
	prefix := schema + ":///"
	return LocatorFunc(func(path string) (io.ReadCloser, error) {
		if !strings.HasPrefix(path, prefix) {
			return nil, ErrNotFound
		}
		return locator.Open(strings.TrimPrefix(path, prefix))
	})
}
