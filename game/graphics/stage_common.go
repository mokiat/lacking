package graphics

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
)

func newCommonStageData(api render.API, cfg *config) *commonStageData {
	return &commonStageData{
		api: api,

		directionalShadowMapCount:        cfg.DirectionalShadowMapCount,
		directionalShadowMapSize:         cfg.DirectionalShadowMapSize,
		directionalShadowMapCascadeCount: cfg.DirectionalShadowMapCascadeCount,
		directionalShadowMaps:            make([]internal.DirectionalShadowMap, cfg.DirectionalShadowMapCount),
		directionalShadowMapAssignments:  make(map[*DirectionalLight]*internal.DirectionalShadowMap, cfg.DirectionalShadowMapCount),

		spotShadowMapCount:       cfg.SpotShadowMapCount,
		spotShadowMapSize:        cfg.SpotShadowMapSize,
		spotShadowMaps:           make([]internal.SpotShadowMap, cfg.SpotShadowMapCount),
		spotShadowMapAssignments: make(map[*SpotLight]*internal.SpotShadowMap, cfg.SpotShadowMapCount),

		pointShadowMapCount:       cfg.PointShadowMapCount,
		pointShadowMapSize:        cfg.PointShadowMapSize,
		pointShadowMaps:           make([]internal.PointShadowMap, cfg.PointShadowMapCount),
		pointShadowMapAssignments: make(map[*PointLight]*internal.PointShadowMap, cfg.PointShadowMapCount),
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

	directionalShadowMapCount        int
	directionalShadowMapSize         int
	directionalShadowMapCascadeCount int
	directionalShadowMaps            []internal.DirectionalShadowMap
	directionalShadowMapAssignments  map[*DirectionalLight]*internal.DirectionalShadowMap

	spotShadowMapCount       int
	spotShadowMapSize        int
	spotShadowMaps           []internal.SpotShadowMap
	spotShadowMapAssignments map[*SpotLight]*internal.SpotShadowMap

	pointShadowMapCount       int
	pointShadowMapSize        int
	pointShadowMaps           []internal.PointShadowMap
	pointShadowMapAssignments map[*PointLight]*internal.PointShadowMap
}

func (d *commonStageData) Allocate() {
	d.quadShape = internal.CreateQuadShape(d.api)
	d.cubeShape = internal.CreateCubeShape(d.api)
	d.sphereShape = internal.CreateSphereShape(d.api)
	d.coneShape = internal.CreateConeShape(d.api)

	d.nearestSampler = d.api.CreateSampler(render.SamplerInfo{
		Label:      "Nearest Sampler",
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeNearest,
		Mipmapping: false,
	})
	d.linearSampler = d.api.CreateSampler(render.SamplerInfo{
		Label:      "Linear Sampler",
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeLinear,
		Mipmapping: false,
	})
	d.depthSampler = d.api.CreateSampler(render.SamplerInfo{
		Label:      "Depth Sampler",
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeLinear,
		Comparison: opt.V(render.ComparisonLess),
		Mipmapping: false,
	})

	d.commandBuffer = d.api.CreateCommandBuffer(commandBufferSize)
	d.uniformBuffer = ubo.NewUniformBlockBuffer(d.api, uniformBufferSize)

	gog.Mutate(d.directionalShadowMaps, func(shadowMap *internal.DirectionalShadowMap) {
		shadowMap.ArrayTexture = d.api.CreateDepthTexture2DArray(render.DepthTexture2DArrayInfo{
			Label:      "Directional Shadow Map Texture",
			Width:      uint32(d.directionalShadowMapSize),
			Height:     uint32(d.directionalShadowMapSize),
			Layers:     uint32(d.directionalShadowMapCascadeCount),
			Comparable: true,
		})
		shadowMap.Cascades = make([]internal.DirectionalShadowMapCascade, d.directionalShadowMapCascadeCount)
		gog.MutateIndex(shadowMap.Cascades, func(j int, cascade *internal.DirectionalShadowMapCascade) {
			cascade.Framebuffer = d.api.CreateFramebuffer(render.FramebufferInfo{
				Label: "Directional Shadow Map Framebuffer",
				DepthAttachment: opt.V(render.TextureAttachment{
					Texture: shadowMap.ArrayTexture,
					Depth:   uint32(j),
				}),
			})
		})
	})

	gog.Mutate(d.spotShadowMaps, func(shadowMap *internal.SpotShadowMap) {
		shadowMap.Texture = d.api.CreateDepthTexture2D(render.DepthTexture2DInfo{
			Label:      "Spot Shadow Map Texture",
			Width:      uint32(d.spotShadowMapSize),
			Height:     uint32(d.spotShadowMapSize),
			Comparable: true,
		})
		shadowMap.Framebuffer = d.api.CreateFramebuffer(render.FramebufferInfo{
			Label:           "Spot Shadow Map Framebuffer",
			DepthAttachment: opt.V(render.PlainTextureAttachment(shadowMap.Texture)),
		})
	})

	gog.Mutate(d.pointShadowMaps, func(shadowMap *internal.PointShadowMap) {
		shadowMap.ArrayTexture = d.api.CreateDepthTexture2DArray(render.DepthTexture2DArrayInfo{
			Label:      "Point Shadow Map Texture",
			Width:      uint32(d.pointShadowMapSize),
			Height:     uint32(d.pointShadowMapSize),
			Layers:     6,
			Comparable: true,
		})
		for i := range 6 {
			shadowMap.Framebuffers[i] = d.api.CreateFramebuffer(render.FramebufferInfo{
				Label: "Point Shadow Map Framebuffer",
				DepthAttachment: opt.V(render.TextureAttachment{
					Texture: shadowMap.ArrayTexture,
					Depth:   uint32(i),
				}),
			})
		}
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

	gog.Mutate(d.directionalShadowMaps, func(shadowMap *internal.DirectionalShadowMap) {
		defer shadowMap.ArrayTexture.Release()
		for _, cascade := range shadowMap.Cascades {
			defer cascade.Framebuffer.Release()
		}
	})

	gog.Mutate(d.spotShadowMaps, func(shadowMap *internal.SpotShadowMap) {
		defer shadowMap.Texture.Release()
		defer shadowMap.Framebuffer.Release()
	})

	gog.Mutate(d.pointShadowMaps, func(shadowMap *internal.PointShadowMap) {
		defer shadowMap.ArrayTexture.Release()
		for _, framebuffer := range shadowMap.Framebuffers {
			defer framebuffer.Release()
		}
	})
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

func (d *commonStageData) ResetDirectionalShadowMapAssignments() {
	clear(d.directionalShadowMapAssignments)
}

func (d *commonStageData) AssignDirectionalShadowMap(light *DirectionalLight) *internal.DirectionalShadowMap {
	freeIndex := len(d.directionalShadowMapAssignments)
	if freeIndex >= len(d.directionalShadowMaps) {
		return nil
	}
	shadowMap := &d.directionalShadowMaps[freeIndex]
	d.directionalShadowMapAssignments[light] = shadowMap
	return shadowMap
}

func (d *commonStageData) GetDirectionalShadowMap(light *DirectionalLight) *internal.DirectionalShadowMap {
	return d.directionalShadowMapAssignments[light]
}

func (d *commonStageData) ResetSpotShadowMapAssignments() {
	clear(d.spotShadowMapAssignments)
}

func (d *commonStageData) AssignSpotShadowMap(light *SpotLight) *internal.SpotShadowMap {
	freeIndex := len(d.spotShadowMapAssignments)
	if freeIndex >= len(d.spotShadowMaps) {
		return nil
	}
	shadowMap := &d.spotShadowMaps[freeIndex]
	d.spotShadowMapAssignments[light] = shadowMap
	return shadowMap
}

func (d *commonStageData) GetSpotShadowMap(light *SpotLight) *internal.SpotShadowMap {
	return d.spotShadowMapAssignments[light]
}

func (d *commonStageData) ResetPointShadowMapAssignments() {
	clear(d.pointShadowMapAssignments)
}

func (d *commonStageData) AssignPointShadowMap(light *PointLight) *internal.PointShadowMap {
	freeIndex := len(d.pointShadowMapAssignments)
	if freeIndex >= len(d.pointShadowMaps) {
		return nil
	}
	shadowMap := &d.pointShadowMaps[freeIndex]
	d.pointShadowMapAssignments[light] = shadowMap
	return shadowMap
}

func (d *commonStageData) GetPointShadowMap(light *PointLight) *internal.PointShadowMap {
	return d.pointShadowMapAssignments[light]
}
