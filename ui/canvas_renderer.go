package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

const (
	maxLayerDepth = 256

	maxVertexCount = 524288
)

func newCanvasRenderer(api render.API, shaders ShaderCollection) *canvasRenderer {
	state := newCanvasState()
	return &canvasRenderer{
		canvasPath: newCanvasPath(),
		api:        api,
		state:      state,

		shapeMesh:          newShapeMesh(maxVertexCount),
		shapeShadeMaterial: newMaterial(shaders.ShapeMaterial),
		shapeBlankMaterial: newMaterial(shaders.ShapeBlankMaterial),

		contour: newContour(state, shaders),
		text:    newText(state, shaders),
	}
}

type canvasRenderer struct {
	*canvasPath
	api          render.API
	state        *canvasState
	commandQueue render.CommandQueue

	shapeMesh            *ShapeMesh
	shapeShadeMaterial   *Material
	shapeBlankMaterial   *Material
	shapeMaskPipeline    render.Pipeline
	shapeSimplePipeline  render.Pipeline
	shapeNonZeroPipeline render.Pipeline
	shapeOddPipeline     render.Pipeline

	contour *Contour
	text    *Text
}

func (c *canvasRenderer) onCreate() {
	c.commandQueue = c.api.CreateCommandQueue()
	c.state.onCreate(c.api)

	c.shapeMesh.Allocate(c.api)
	c.shapeShadeMaterial.Allocate(c.api)
	c.shapeBlankMaterial.Allocate(c.api)
	c.shapeMaskPipeline = c.api.CreatePipeline(render.PipelineInfo{
		Program:         c.shapeBlankMaterial.program,
		VertexArray:     c.shapeMesh.vertexArray,
		Topology:        render.TopologyTriangleFan,
		Culling:         render.CullModeNone,
		FrontFace:       render.FaceOrientationCCW,
		LineWidth:       1.0,
		DepthTest:       false,
		DepthWrite:      false,
		DepthComparison: render.ComparisonAlways,
		StencilTest:     true,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationIncreaseWrap,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationDecreaseWrap,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		ColorWrite:                  render.ColorMaskFalse,
		BlendEnabled:                false,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
	c.shapeSimplePipeline = c.api.CreatePipeline(render.PipelineInfo{
		Program:                     c.shapeShadeMaterial.program,
		VertexArray:                 c.shapeMesh.vertexArray,
		Topology:                    render.TopologyTriangleFan,
		Culling:                     render.CullModeNone,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		StencilTest:                 false,
		ColorWrite:                  render.ColorMaskTrue,
		BlendEnabled:                true,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
	c.shapeNonZeroPipeline = c.api.CreatePipeline(render.PipelineInfo{
		Program:         c.shapeShadeMaterial.program,
		VertexArray:     c.shapeMesh.vertexArray,
		Topology:        render.TopologyTriangleFan,
		Culling:         render.CullModeNone,
		FrontFace:       render.FaceOrientationCCW,
		LineWidth:       1.0,
		DepthTest:       false,
		DepthWrite:      false,
		DepthComparison: render.ComparisonAlways,
		StencilTest:     true,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationReplace,
			Comparison:     render.ComparisonNotEqual,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationReplace,
			Comparison:     render.ComparisonNotEqual,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		ColorWrite:                  render.ColorMaskTrue,
		BlendEnabled:                true,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
	c.shapeOddPipeline = c.api.CreatePipeline(render.PipelineInfo{
		Program:     c.shapeShadeMaterial.program,
		VertexArray: c.shapeMesh.vertexArray,
		Topology:    render.TopologyTriangleFan,
		Culling:     render.CullModeNone,
		FrontFace:   render.FaceOrientationCCW,
		DepthTest:   false,
		DepthWrite:  false,
		StencilTest: true,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationReplace,
			Comparison:     render.ComparisonNotEqual,
			ComparisonMask: 0x01,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationReplace,
			Comparison:     render.ComparisonNotEqual,
			ComparisonMask: 0x01,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		ColorWrite:                  render.ColorMaskTrue,
		BlendEnabled:                true,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})

	c.contour.onCreate(c.api, c.commandQueue)
	c.text.onCreate(c.api, c.commandQueue)
}

func (c *canvasRenderer) onDestroy() {
	defer c.commandQueue.Release()
	defer c.state.onDestroy()

	defer c.shapeMesh.Release()
	defer c.shapeShadeMaterial.Release()
	defer c.shapeBlankMaterial.Release()
	defer c.shapeMaskPipeline.Release()
	defer c.shapeSimplePipeline.Release()
	defer c.shapeNonZeroPipeline.Release()
	defer c.shapeOddPipeline.Release()

	defer c.contour.onDestroy()
	defer c.text.onDestroy()
}

func (c *canvasRenderer) onBegin(size Size) {
	c.state.currentLayer = c.state.topLayer
	c.state.currentLayer.ClipBounds.Position = NewPosition(0, 0)
	c.state.currentLayer.ClipBounds.Size = size
	c.state.currentLayer.Transform = sprec.IdentityMat4()
	c.state.projectionMatrix = sprec.OrthoMat4(
		0.0, float32(size.Width),
		0.0, float32(size.Height),
		0.0, 1.0,
	)

	c.shapeMesh.Reset()

	c.contour.onBegin()
	c.text.onBegin()
}

func (c *canvasRenderer) onEnd() {
	c.shapeMesh.Update()

	c.contour.onEnd()
	c.text.onEnd()
	c.api.SubmitQueue(c.commandQueue)
}

// Push records the current state and creates a new state layer. Changes done
// in the new layer will not affect the parent layer.
//
// You may create up to 256 layers including the starting one after which the
// method panics.
func (c *canvasRenderer) Push() {
	c.state.currentLayer = c.state.currentLayer.Next()
}

// Pop restores the drawing state based on the parent layer. If this is the
// first layer, then this method panics.
func (c *canvasRenderer) Pop() {
	c.state.currentLayer = c.state.currentLayer.Previous()
}

// ResetTransform restores the transform to the value it had
// after the last Push. If this is the first layer, then it is
// set to the identity matrix.
func (c *canvasRenderer) ResetTransform() {
	if c.state.currentLayer == c.state.topLayer {
		c.state.currentLayer.Transform = sprec.IdentityMat4()
	} else {
		c.state.currentLayer.Transform = c.state.currentLayer.previous.Transform
	}
}

// SetTransform changes the transform relative to the former layer transform.
func (c *canvasRenderer) SetTransform(transform sprec.Mat4) {
	if c.state.currentLayer == c.state.topLayer {
		c.state.currentLayer.Transform = transform
	} else {
		c.state.currentLayer.Transform = sprec.Mat4Prod(
			c.state.currentLayer.previous.Transform,
			transform,
		)
	}
}

// Translate moves the drawing position by the specified delta amount.
func (c *canvasRenderer) Translate(delta sprec.Vec2) {
	c.state.currentLayer.Transform = sprec.Mat4Prod(
		c.state.currentLayer.Transform,
		sprec.TranslationMat4(delta.X, delta.Y, 0.0),
	)
}

func (c *canvasRenderer) SetClipBounds(left, right, top, bottom float32) {
	c.state.currentLayer.ClipBounds = NewBounds(
		int(left),
		int(top),
		int(right-left),
		int(bottom-top),
	)
}

// Clip sets new clipping bounds. Pixels from draw operations
// that are outside the clipping bounds will not be drawn.
//
// Initially the clipping bounds are equal to the window size.
func (c *canvasRenderer) Clip(bounds Bounds) {
	// FIXME: This no longer works correctly
	c.state.currentLayer.ClipBounds = bounds.Translate(
		NewPosition(
			int(c.state.currentLayer.Transform.Translation().X),
			int(c.state.currentLayer.Transform.Translation().Y),
		),
	)
}

// Contour returns the contour rendering module.
func (c *canvasRenderer) Contour() *Contour {
	return c.contour
}

// Text returns the text rendering module.
func (c *canvasRenderer) Text() *Text {
	return c.text
}

// DrawSurface renders the specified surface. The surface's Render
// method will be called when needed with the UI framebuffer bound.
func (c *canvasRenderer) DrawSurface(surface Surface, position Position, size Size) {
	texture := surface.Render(size.Width, size.Height)

	c.Reset()
	c.Rectangle(
		sprec.NewVec2(float32(position.X), float32(position.Y)),
		sprec.NewVec2(float32(size.Width), float32(size.Height)),
	)
	c.Fill(Fill{
		Rule: FillRuleSimple,
		Image: &Image{ // TODO: Don't allocate
			texture: texture,
			size:    size,
		},
		Color:       White(),
		ImageOffset: sprec.NewVec2(0.0, float32(size.Height)),
		ImageSize:   sprec.NewVec2(float32(size.Width), -float32(size.Height)),
	})
}

func (c *canvasRenderer) Fill(fill Fill) {
	c.fillPath(c.canvasPath, fill)
}

func (c *canvasRenderer) Stroke() {
	c.strokePath(c.canvasPath)
}

func (c *canvasRenderer) fillPath(path *canvasPath, fill Fill) {
	currentLayer := c.state.currentLayer
	clipBounds := sprec.NewVec4(
		float32(currentLayer.ClipBounds.X),
		float32(currentLayer.ClipBounds.X+currentLayer.ClipBounds.Width),
		float32(currentLayer.ClipBounds.Y),
		float32(currentLayer.ClipBounds.Y+currentLayer.ClipBounds.Height),
	)
	transformMatrix := currentLayer.Transform
	textureTransformMatrix := sprec.Mat4MultiProd(
		sprec.ScaleMat4(
			1.0/fill.ImageSize.X,
			1.0/fill.ImageSize.Y,
			1.0,
		),
		sprec.TranslationMat4(
			-fill.ImageOffset.X,
			-fill.ImageOffset.Y,
			0.0,
		),
	)

	vertexOffset := c.shapeMesh.Offset()
	for _, point := range c.points {
		c.shapeMesh.Append(ShapeVertex{
			position: point.coords,
		})
	}

	// draw stencil mask for all sub-shapes
	if fill.Rule != FillRuleSimple {
		c.commandQueue.BindPipeline(c.shapeMaskPipeline)
		c.commandQueue.Uniform4f(c.shapeBlankMaterial.clipDistancesLocation, clipBounds.Array())
		c.commandQueue.UniformMatrix4f(c.shapeBlankMaterial.projectionMatrixLocation, c.state.projectionMatrix.ColumnMajorArray())
		c.commandQueue.UniformMatrix4f(c.shapeBlankMaterial.transformMatrixLocation, transformMatrix.ColumnMajorArray())

		for i, pointOffset := range c.subPathOffsets {
			pointCount := len(c.points) - pointOffset
			if i+1 < len(c.subPathOffsets) {
				pointCount = c.subPathOffsets[i+1] - pointOffset
			}
			c.commandQueue.Draw(vertexOffset+pointOffset, pointCount, 1)
		}
	}

	// render color shader shape and clear stencil buffer
	switch fill.Rule {
	case FillRuleSimple:
		c.commandQueue.BindPipeline(c.shapeSimplePipeline)
	case FillRuleNonZero:
		c.commandQueue.BindPipeline(c.shapeNonZeroPipeline)
	case FillRuleEvenOdd:
		c.commandQueue.BindPipeline(c.shapeOddPipeline)
	default:
		c.commandQueue.BindPipeline(c.shapeSimplePipeline)
	}

	texture := c.state.whiteMask
	if fill.Image != nil {
		texture = fill.Image.texture
	}
	color := uiColorToVec(fill.Color)

	c.commandQueue.Uniform4f(c.shapeShadeMaterial.clipDistancesLocation, clipBounds.Array())
	c.commandQueue.Uniform4f(c.shapeShadeMaterial.colorLocation, color.Array())
	c.commandQueue.UniformMatrix4f(c.shapeShadeMaterial.projectionMatrixLocation, c.state.projectionMatrix.ColumnMajorArray())
	c.commandQueue.UniformMatrix4f(c.shapeShadeMaterial.transformMatrixLocation, transformMatrix.ColumnMajorArray())
	c.commandQueue.UniformMatrix4f(c.shapeShadeMaterial.textureTransformMatrixLocation, textureTransformMatrix.ColumnMajorArray())
	c.commandQueue.TextureUnit(0, texture)
	c.commandQueue.Uniform1i(c.shapeShadeMaterial.textureLocation, 0)

	for i, pointOffset := range c.subPathOffsets {
		pointCount := len(c.points) - pointOffset
		if i+1 < len(c.subPathOffsets) {
			pointCount = c.subPathOffsets[i+1] - pointOffset
		}
		c.commandQueue.Draw(vertexOffset+pointOffset, pointCount, 1)
	}
}

func (c *canvasRenderer) strokePath(path *canvasPath) {
	// TODO: Implement directly and remove old API
	c.Contour().begin()
	for i := 0; i < len(path.subPathOffsets); i++ {
		offset := path.subPathOffsets[i]
		nextOffset := len(path.points)
		if i+1 < len(path.subPathOffsets) {
			nextOffset = path.subPathOffsets[i+1]
		}
		for j, point := range path.points[offset:nextOffset] {
			if j == 0 {
				c.Contour().moveTo(point.coords, Stroke{
					InnerSize: point.innerSize,
					OuterSize: point.outerSize,
					Color:     point.color,
				})
			} else {
				c.Contour().lineTo(point.coords, Stroke{
					InnerSize: point.innerSize,
					OuterSize: point.outerSize,
					Color:     point.color,
				})
			}
		}
	}
	c.Contour().end()
}

const (
	// FillRuleSimple is the fastest approach and should be used
	// with non-overlapping concave shapes.
	FillRuleSimple FillRule = iota

	// FillRuleNonZero will fill areas that are covered by the
	// shape, regardless if it overlaps.
	FillRuleNonZero

	// FillRuleEvenOdd will fill areas that are covered by the
	// shape and it does not overlap or overlaps an odd number
	// of times.
	FillRuleEvenOdd
)

// FillRule represents the mechanism through which it is determined
// which point is part of the shape in an overlapping or concave
// polygon.
type FillRule int

// Fill configures how a solid shape is to be drawn.
type Fill struct {

	// Rule specifies the mechanism through which it is determined
	// which point is part of the shape in an overlapping or concave
	// polygon.
	Rule FillRule

	// Color specifies the color to use to fill the shape.
	Color Color

	// Image specifies an optional image to be used for filling
	// the shape.
	Image *Image

	// ImageOffset determines the offset of the origin of the
	// image relative to the current translation context.
	ImageOffset sprec.Vec2

	// ImageSize determines the size of the drawn image. In
	// essence, this size performs scaling.
	ImageSize sprec.Vec2
}

// Surface represents an auxiliary drawer.
type Surface interface {
	Render(width, height int) render.Texture
}
