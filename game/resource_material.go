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
	geometryShaders []*graphics.GeometryShader,
	shadowShaders []*graphics.ShadowShader,
	forwardShaders []*graphics.ForwardShader,
	textures []render.Texture,
	assetMaterial asset.Material,
) async.Promise[*graphics.Material] {

	materialInfo := graphics.MaterialInfo{
		Name: assetMaterial.Name,
		GeometryPasses: gog.Map(assetMaterial.GeometryPasses, func(pass asset.GeometryPass) graphics.GeometryRenderPassInfo {
			return graphics.GeometryRenderPassInfo{
				Layer:           pass.Layer,
				Culling:         opt.V(s.resolveCullMode(pass.Culling)),
				FrontFace:       opt.V(s.resolveFaceOrientation(pass.FrontFace)),
				DepthTest:       opt.V(pass.DepthTest),
				DepthWrite:      opt.V(pass.DepthWrite),
				DepthComparison: opt.V(s.resolveComparison(pass.DepthComparison)),
				Shader:          geometryShaders[pass.ShaderIndex],
			}
		}),
		ShadowPasses: gog.Map(assetMaterial.ShadowPasses, func(pass asset.ShadowPass) graphics.ShadowRenderPassInfo {
			return graphics.ShadowRenderPassInfo{
				Culling:   opt.V(s.resolveCullMode(pass.Culling)),
				FrontFace: opt.V(s.resolveFaceOrientation(pass.FrontFace)),
				Shader:    shadowShaders[pass.ShaderIndex],
			}
		}),
		ForwardPasses: gog.Map(assetMaterial.ForwardPasses, func(pass asset.ForwardPass) graphics.ForwardRenderPassInfo {
			return graphics.ForwardRenderPassInfo{
				Layer:           pass.Layer,
				Culling:         opt.V(s.resolveCullMode(pass.Culling)),
				FrontFace:       opt.V(s.resolveFaceOrientation(pass.FrontFace)),
				DepthTest:       opt.V(pass.DepthTest),
				DepthWrite:      opt.V(pass.DepthWrite),
				DepthComparison: opt.V(s.resolveComparison(pass.DepthComparison)),
				Blending:        opt.V(pass.Blending),
				Shader:          forwardShaders[pass.ShaderIndex],
			}
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
