package game

import (
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertSkyDefinition(materials []*graphics.Material, assetSky asset.Sky) async.Promise[*graphics.SkyDefinition] {
	skyDefinitionInfo := graphics.SkyDefinitionInfo{
		Material: materials[assetSky.MaterialIndex],
	}

	promise := async.NewPromise[*graphics.SkyDefinition]()
	s.gfxWorker.Schedule(func() {
		gfxEngine := s.engine.Graphics()
		skyDefinition := gfxEngine.CreateSkyDefinition(skyDefinitionInfo)
		promise.Deliver(skyDefinition)
	})
	return promise
}

func (s *ResourceSet) convertSky(definitionIndex int, assetSky asset.Sky) skyInstance {
	return skyInstance{
		nodeIndex:       int(assetSky.NodeIndex),
		definitionIndex: definitionIndex,
	}
}
