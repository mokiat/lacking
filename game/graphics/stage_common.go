package graphics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
)

func newCommonStageData(api render.API) *commonStageData {
	return &commonStageData{
		api: api,
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
