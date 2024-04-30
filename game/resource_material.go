package game

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertMaterial(
	shaders []*graphics.Shader,
	textures []render.Texture,
	assetMaterial asset.Material,
) async.Promise[*graphics.Material] {

	materialInfo := graphics.MaterialInfo{
		Name: assetMaterial.Name,
		GeometryPasses: gog.Map(assetMaterial.GeometryPasses, func(pass asset.MaterialPass) graphics.MaterialPassInfo {
			return s.convertMaterialPass(shaders, pass)
		}),
		ShadowPasses: gog.Map(assetMaterial.ShadowPasses, func(pass asset.MaterialPass) graphics.MaterialPassInfo {
			return s.convertMaterialPass(shaders, pass)
		}),
		ForwardPasses: gog.Map(assetMaterial.ForwardPasses, func(pass asset.MaterialPass) graphics.MaterialPassInfo {
			return s.convertMaterialPass(shaders, pass)
		}),
		SkyPasses: gog.Map(assetMaterial.SkyPasses, func(pass asset.MaterialPass) graphics.MaterialPassInfo {
			return s.convertMaterialPass(shaders, pass)
		}),
		PostprocessingPasses: gog.Map(assetMaterial.PostprocessingPasses, func(pass asset.MaterialPass) graphics.MaterialPassInfo {
			return s.convertMaterialPass(shaders, pass)
		}),
	}

	promise := async.NewPromise[*graphics.Material]()
	s.gfxWorker.Schedule(func() {
		gfxEngine := s.engine.Graphics()
		material := gfxEngine.CreateMaterial(materialInfo)
		for _, binding := range assetMaterial.Textures {
			texture := textures[binding.TextureIndex]
			material.SetTexture(binding.BindingName, texture)
		}
		for _, binding := range assetMaterial.Textures {
			sampler := s.renderAPI.CreateSampler(render.SamplerInfo{
				Wrapping:   s.resolveWrapMode(binding.Wrapping),
				Filtering:  s.resolveFiltering(binding.Filtering),
				Mipmapping: binding.Mipmapping,
			})
			material.SetSampler(binding.BindingName, sampler)
		}
		for _, binding := range assetMaterial.Properties {
			material.SetProperty(binding.BindingName, binding.Data)
		}
		promise.Deliver(material)
	})
	return promise
}

func (s *ResourceSet) convertMaterialPass(shaders []*graphics.Shader, assetPass asset.MaterialPass) graphics.MaterialPassInfo {
	return graphics.MaterialPassInfo{
		Layer:           assetPass.Layer,
		Culling:         opt.V(s.resolveCullMode(assetPass.Culling)),
		FrontFace:       opt.V(s.resolveFaceOrientation(assetPass.FrontFace)),
		DepthTest:       opt.V(assetPass.DepthTest),
		DepthWrite:      opt.V(assetPass.DepthWrite),
		DepthComparison: opt.V(s.resolveComparison(assetPass.DepthComparison)),
		Blending:        opt.V(assetPass.Blending),
		Shader:          shaders[assetPass.ShaderIndex],
	}
}
