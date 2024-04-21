package game

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertSkyDefinition(textures []render.Texture, skyShaders []*graphics.SkyShader, assetSky asset.Sky) async.Promise[*graphics.SkyDefinition] {
	skyDefinitionInfo := graphics.SkyDefinitionInfo{
		Layers: gog.Map(assetSky.Layers, func(layerAsset asset.SkyLayer) graphics.SkyLayerDefinitionInfo {
			return graphics.SkyLayerDefinitionInfo{
				Shader:   skyShaders[layerAsset.ShaderIndex],
				Blending: layerAsset.Blending,
			}
		}),
	}

	promise := async.NewPromise[*graphics.SkyDefinition]()
	s.gfxWorker.ScheduleVoid(func() {
		gfxEngine := s.engine.Graphics()
		skyDefinition := gfxEngine.CreateSkyDefinition(skyDefinitionInfo)
		for _, binding := range assetSky.Textures {
			texture := textures[binding.TextureIndex]
			skyDefinition.SetTexture(binding.BindingName, texture)
		}
		for _, binding := range assetSky.Textures {
			sampler := s.renderAPI.CreateSampler(render.SamplerInfo{
				Wrapping:   s.resolveWrapMode(binding.Wrapping),
				Filtering:  s.resolveFiltering(binding.Filtering),
				Mipmapping: binding.Mipmapping,
			})
			skyDefinition.SetSampler(binding.BindingName, sampler)
		}
		for _, binding := range assetSky.Properties {
			skyDefinition.SetProperty(binding.BindingName, binding.Data)
		}
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
