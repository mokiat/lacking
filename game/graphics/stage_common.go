package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	renderutil "github.com/mokiat/lacking/render/util"
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

	commandBuffer render.CommandBuffer
	uniformBuffer *renderutil.UniformBlockBuffer
}

func (d *commonStageData) Allocate() {
	d.quadShape = internal.CreateQuadShape(d.api)
	d.cubeShape = internal.CreateCubeShape(d.api)
	d.sphereShape = internal.CreateSphereShape(d.api)
	d.coneShape = internal.CreateConeShape(d.api)

	d.commandBuffer = d.api.CreateCommandBuffer(commandBufferSize)
	d.uniformBuffer = renderutil.NewUniformBlockBuffer(d.api, uniformBufferSize)
}

func (d *commonStageData) Release() {
	defer d.quadShape.Release()
	defer d.cubeShape.Release()
	defer d.sphereShape.Release()
	defer d.coneShape.Release()

	defer d.uniformBuffer.Release()
}

func (d *commonStageData) CommandBuffer() render.CommandBuffer {
	return d.commandBuffer
}

func (d *commonStageData) UniformBuffer() *renderutil.UniformBlockBuffer {
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
