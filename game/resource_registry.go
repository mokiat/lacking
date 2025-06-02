package game

import (
	"fmt"
	"log/slog"
	"reflect"
	"sync"

	"github.com/mokiat/lacking/storage/chunked"
	"github.com/mokiat/lacking/util/async"
)

func newResourceRegistry(engine *Engine, storage chunked.Storage) *resourceRegistry {
	return &resourceRegistry{
		engine:  engine,
		storage: storage,

		resourceLoaders: make(map[reflect.Type]ResourceLoader[any]),
		resources:       make(map[string]*resourceHandle),
	}
}

type resourceRegistry struct {
	engine  *Engine
	storage chunked.Storage

	mu              sync.Mutex
	resourceLoaders map[reflect.Type]ResourceLoader[any]
	resources       map[string]*resourceHandle
}

func (r *resourceRegistry) RegisterResourceLoader(resourceLoader ResourceLoader[any]) {
	resourceType := resourceLoader.ApplicableType()
	r.resourceLoaders[resourceType] = resourceLoader
}

func (r *resourceRegistry) UnregisterResourceLoader(resourceLoader ResourceLoader[any]) {
	resourceType := resourceLoader.ApplicableType()
	delete(r.resourceLoaders, resourceType)
}

func (r *resourceRegistry) LoadResource(resourceSet *ResourceSet, path string, target any) async.Operation {
	reflValue := reflect.ValueOf(target)
	if reflValue.Kind() != reflect.Ptr || reflValue.IsNil() {
		return async.NewFailedOperation(fmt.Errorf("target must be a non-nil pointer, got %T", target))
	}
	reflValue = reflValue.Elem()

	r.mu.Lock()
	defer r.mu.Unlock()

	var promise async.Promise[any]
	if handle, ok := r.resources[path]; ok {
		handle.refCount++
		promise = handle.promise
	} else {
		resourceType := reflValue.Type()
		resourceLoader, ok := r.resourceLoaders[resourceType]
		if !ok {
			return async.NewFailedOperation(fmt.Errorf("no resource loader registered for type: %s", resourceType.String()))
		}
		promise = async.NewPromise[any]()
		r.resources[path] = &resourceHandle{
			resourceLoader: resourceLoader,
			promise:        promise,
			refCount:       1,
		}
		asset := chunked.NewAsset(r.storage, path)
		go func() {
			assetLoader := &AssetLoader{
				engine:      r.engine,
				resourceSet: resourceSet,
			}
			resource, err := resourceLoader.LoadResource(assetLoader, asset)
			if err != nil {
				promise.Fail(err)
			} else {
				promise.Deliver(resource)
			}
		}()
	}

	return async.NewFuncOperation(func() error {
		resource, err := promise.Wait()
		if err != nil {
			return err
		}
		reflValue.Set(reflect.ValueOf(resource))
		return nil
	})
}

func (r *resourceRegistry) UnloadResource(resourceSet *ResourceSet, path string, count int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	handle, ok := r.resources[path]
	if !ok {
		logger.Error("Trying to unload unknown resource", slog.String("path", path))
		return
	}

	if handle.refCount -= count; handle.refCount > 0 {
		return // still in use
	}
	delete(r.resources, path)

	resourceLoader := handle.resourceLoader
	promise := handle.promise
	go func() {
		resource, err := promise.Wait()
		if err != nil {
			return // cannot unload resource that failed to load
		}
		assetLoader := &AssetLoader{
			engine:      r.engine,
			resourceSet: resourceSet,
		}
		if err := resourceLoader.UnloadResource(assetLoader, resource); err != nil {
			logger.Error("Failed to unload resource", slog.String("path", path), slog.String("error", err.Error()))
		}
	}()
}
