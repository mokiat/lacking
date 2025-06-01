package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/storage/chunked"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) loadModel(asyncEngine *AsyncEngine, resource *chunked.Asset) (*ModelTemplate, error) {
	assetModelPromise := s.openResource(resource)
	assetModel, err := assetModelPromise.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}

	loader := &AssetLoader{
		resourceSet: s,
		asyncEngine: asyncEngine,
	}
	return loader.ResolveModelTemplate(assetModel)
}

func (s *ResourceSet) freeModel(model *ModelTemplate) {
	s.gfxWorker.Schedule(func() {
		for skyTemplate := range model.SkyTemplates.Values() {
			definition := skyTemplate.Definition
			definition.Delete()
		}
	})
	s.gfxWorker.Schedule(func() {
		for meshDefinition := range model.MeshDefinitions.Values() {
			meshDefinition.Delete()
		}
	})
	s.gfxWorker.Schedule(func() {
		for meshGeometry := range model.MeshGeometries.Values() {
			meshGeometry.Delete()
		}
	})
	s.gfxWorker.Schedule(func() {
		for texture := range model.Textures.Values() {
			texture.Release()
		}
	})
}

func (s *ResourceSet) openResource(resource *chunked.Asset) async.Promise[dto.Model] {
	promise := async.NewPromise[dto.Model]()
	s.ioWorker.Schedule(func() {
		var assetModel dto.Model
		if err := resource.Read(&assetModel); err != nil {
			promise.Fail(err)
		} else {
			promise.Deliver(assetModel)
		}
	})
	return promise
}
