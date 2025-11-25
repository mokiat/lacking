package game

import (
	"reflect"

	"github.com/mokiat/lacking/storage/chunked"
)

// ResourceLoader represents a loader for a specific type of resource.
type ResourceLoader[T any] interface {

	// ApplicableType returns the reflect.Type of the resource that this loader
	// is capable of loading. This is used to determine which loader to use
	// when loading a resource.
	ApplicableType() reflect.Type

	// LoadResource loads the resource from the given asset.
	LoadResource(loader *AssetLoader, asset *chunked.Asset) (T, error)

	// UnloadResource unloads the given resource. This is called when the resource
	// is no longer needed and should be cleaned up.
	UnloadResource(loader *AssetLoader, resource T) error
}

// GenericResourceLoader allows a generic resource loader to be passed to the
// engine.
func GenericResourceLoader[T any](delegate ResourceLoader[T]) ResourceLoader[any] {
	return &genericResourceLoader[T]{
		delegate: delegate,
	}
}

type genericResourceLoader[T any] struct {
	delegate ResourceLoader[T]
}

func (l *genericResourceLoader[T]) ApplicableType() reflect.Type {
	return l.delegate.ApplicableType()
}

func (l *genericResourceLoader[T]) LoadResource(loader *AssetLoader, asset *chunked.Asset) (any, error) {
	return l.delegate.LoadResource(loader, asset)
}

func (l *genericResourceLoader[T]) UnloadResource(loader *AssetLoader, resource any) error {
	return l.delegate.UnloadResource(loader, resource.(T))
}
