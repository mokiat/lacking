package game

import (
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertSkyDefinition(materials map[uint32]*graphics.Material, assetSky dto.Sky) async.Promise[*graphics.SkyDefinition] {
	skyDefinitionInfo := graphics.SkyDefinitionInfo{
		Material: materials[assetSky.MaterialID],
	}

	promise := async.NewPromise[*graphics.SkyDefinition]()
	s.gfxWorker.Schedule(func() {
		gfxEngine := s.engine.Graphics()
		skyDefinition := gfxEngine.CreateSkyDefinition(skyDefinitionInfo)
		promise.Deliver(skyDefinition)
	})
	return promise
}

func (s *ResourceSet) convertSky(definitionIndex int, assetSky dto.Sky) skyInstance {
	return skyInstance{
		nodeID:          assetSky.NodeID,
		definitionIndex: definitionIndex,
	}
}
