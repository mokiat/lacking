package graphics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
)

func newCommonStageData(api render.API, cfg *config) *commonStageData {
	return &commonStageData{
		api: api,

		cascadeShadowMapSize:  cfg.CascadeShadowMapSize,
		cascadeShadowMapCount: cfg.CascadeShadowMapCount,
		cascadeShadowMaps:     make([]internal.CascadeShadowMap, cfg.CascadeShadowMapCount),

		atlasShadowMapSize:    cfg.AtlasShadowMapSize,
		atlasShadowMapSectors: cfg.AtlasShadowMapSectors,
	}
}

type commonStageData struct {
	api render.API

	quadShape   *internal.Shape
	cubeShape   *internal.Shape
	sphereShape *internal.Shape
	coneShape   *internal.Shape

	nearestSampler render.Sampler
	linearSampler  render.Sampler
	depthSampler   render.Sampler

	commandBuffer render.CommandBuffer
	uniformBuffer *ubo.UniformBlockBuffer

	cascadeShadowMapSize  int
	cascadeShadowMapCount int
	cascadeShadowMaps     []internal.CascadeShadowMap

	atlasShadowMapSize    int
	atlasShadowMapSectors int
	atlasShadowMap        internal.AtlasShadowMap
}

func (d *commonStageData) Allocate() {
	d.quadShape = internal.CreateQuadShape(d.api)
	d.cubeShape = internal.CreateCubeShape(d.api)
	d.sphereShape = internal.CreateSphereShape(d.api)
	d.coneShape = internal.CreateConeShape(d.api)

	d.nearestSampler = d.api.CreateSampler(render.SamplerInfo{
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeNearest,
		Mipmapping: false,
	})
	d.linearSampler = d.api.CreateSampler(render.SamplerInfo{
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeLinear,
		Mipmapping: false,
	})
	d.depthSampler = d.api.CreateSampler(render.SamplerInfo{
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeLinear,
		Comparison: opt.V(render.ComparisonLess),
		Mipmapping: false,
	})

	d.commandBuffer = d.api.CreateCommandBuffer(commandBufferSize)
	d.uniformBuffer = ubo.NewUniformBlockBuffer(d.api, uniformBufferSize)

	for i := range d.cascadeShadowMaps {
		cascadeShadowMap := &d.cascadeShadowMaps[i]
		cascadeShadowMap.Texture = d.api.CreateDepthTexture2D(render.DepthTexture2DInfo{
			Width:      uint32(d.cascadeShadowMapSize),
			Height:     uint32(d.cascadeShadowMapSize),
			Comparable: true,
		})
		cascadeShadowMap.Framebuffer = d.api.CreateFramebuffer(render.FramebufferInfo{
			DepthAttachment: cascadeShadowMap.Texture,
		})
	}

	d.atlasShadowMap.Texture = d.api.CreateDepthTexture2D(render.DepthTexture2DInfo{
		Width:      uint32(d.atlasShadowMapSize),
		Height:     uint32(d.atlasShadowMapSize),
		Comparable: true,
	})
	d.atlasShadowMap.Framebuffer = d.api.CreateFramebuffer(render.FramebufferInfo{
		DepthAttachment: d.atlasShadowMap.Texture,
	})
}

func (d *commonStageData) Release() {
	defer d.quadShape.Release()
	defer d.cubeShape.Release()
	defer d.sphereShape.Release()
	defer d.coneShape.Release()

	defer d.nearestSampler.Release()
	defer d.linearSampler.Release()
	defer d.depthSampler.Release()

	defer d.uniformBuffer.Release()

	for i := range d.cascadeShadowMaps {
		cascadeShadowMap := &d.cascadeShadowMaps[i]
		defer cascadeShadowMap.Texture.Release()
		defer cascadeShadowMap.Framebuffer.Release()
	}

	defer d.atlasShadowMap.Texture.Release()
	defer d.atlasShadowMap.Framebuffer.Release()
}

func (d *commonStageData) CommandBuffer() render.CommandBuffer {
	return d.commandBuffer
}

func (d *commonStageData) UniformBuffer() *ubo.UniformBlockBuffer {
	return d.uniformBuffer
}

func (d *commonStageData) QuadShape() *internal.Shape {
	return d.quadShape
}

func (d *commonStageData) CubeShape() *internal.Shape {
	return d.cubeShape
}

func (d *commonStageData) SphereShape() *internal.Shape {
	return d.sphereShape
}

func (d *commonStageData) ConeShape() *internal.Shape {
	return d.coneShape
}

func (d *commonStageData) NearestSampler() render.Sampler {
	return d.nearestSampler
}

func (d *commonStageData) LinearSampler() render.Sampler {
	return d.linearSampler
}

func (d *commonStageData) DepthSampler() render.Sampler {
	return d.depthSampler
}
