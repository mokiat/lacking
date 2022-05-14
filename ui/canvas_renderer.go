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
	return &canvasRenderer{
		canvasPath: newCanvasPath(),

		api: api,

		shapeMesh:          newShapeMesh(maxVertexCount),
		shapeShadeMaterial: newMaterial(shaders.ShapeMaterial),
		shapeBlankMaterial: newMaterial(shaders.ShapeBlankMaterial),

		contourMesh:     newContourMesh(maxVertexCount),
		contourMaterial: newMaterial(shaders.ContourMaterial),

		textMesh:     newTextMesh(maxVertexCount),
		textMaterial: newMaterial(shaders.TextMaterial),

		topLayer:         &canvasLayer{},
		projectionMatrix: sprec.IdentityMat4(),
	}
}

type canvasRenderer struct {
	*canvasPath

	api          render.API
	commandQueue render.CommandQueue

	whiteMask render.Texture

	shapeMesh            *shapeMesh
	shapeShadeMaterial   *material
	shapeBlankMaterial   *material
	shapeMaskPipeline    render.Pipeline
	shapeSimplePipeline  render.Pipeline
	shapeNonZeroPipeline render.Pipeline
	shapeOddPipeline     render.Pipeline

	contourMesh     *contourMesh
	contourMaterial *material
	contourPipeline render.Pipeline

	textMesh     *textMesh
	textMaterial *material
	textPipeline render.Pipeline

	topLayer         *canvasLayer
	currentLayer     *canvasLayer
	projectionMatrix sprec.Mat4
}

func (c *canvasRenderer) onCreate() {
	c.commandQueue = c.api.CreateCommandQueue()

	c.whiteMask = c.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           1,
		Height:          1,
		Filtering:       render.FilterModeNearest,
		Wrapping:        render.WrapModeClamp,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
		Data:            []byte{0xFF, 0xFF, 0xFF, 0xFF},
	})

	c.shapeMesh.Allocate(c.api)
	c.shapeShadeMaterial.Allocate(c.api)
	c.shapeBlankMaterial.Allocate(c.api)
	c.shapeMaskPipeline = c.api.CreatePipeline(render.PipelineInfo{
		Program:     c.shapeBlankMaterial.program,
		VertexArray: c.shapeMesh.vertexArray,
		Topology:    render.TopologyTriangleFan,
		Culling:     render.CullModeNone,
		FrontFace:   render.FaceOrientationCCW,
		LineWidth:   1.0,
		DepthTest:   false,
		DepthWrite:  false,
		StencilTest: true,
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
		ColorWrite:   render.ColorMaskFalse,
		BlendEnabled: false,
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
		Program:     c.shapeShadeMaterial.program,
		VertexArray: c.shapeMesh.vertexArray,
		Topology:    render.TopologyTriangleFan,
		Culling:     render.CullModeNone,
		FrontFace:   render.FaceOrientationCCW,
		LineWidth:   1.0,
		DepthTest:   false,
		DepthWrite:  false,
		StencilTest: true,
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

	c.contourMesh.Allocate(c.api)
	c.contourMaterial.Allocate(c.api)
	c.contourPipeline = c.api.CreatePipeline(render.PipelineInfo{
		Program:                     c.contourMaterial.program,
		VertexArray:                 c.contourMesh.vertexArray,
		Topology:                    render.TopologyTriangles,
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

	c.textMesh.Allocate(c.api)
	c.textMaterial.Allocate(c.api)
	c.textPipeline = c.api.CreatePipeline(render.PipelineInfo{
		Program:                     c.textMaterial.program,
		VertexArray:                 c.textMesh.vertexArray,
		Topology:                    render.TopologyTriangles,
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
}

func (c *canvasRenderer) onDestroy() {
	defer c.commandQueue.Release()

	defer c.whiteMask.Release()

	defer c.shapeMesh.Release()
	defer c.shapeShadeMaterial.Release()
	defer c.shapeBlankMaterial.Release()
	defer c.shapeMaskPipeline.Release()
	defer c.shapeSimplePipeline.Release()
	defer c.shapeNonZeroPipeline.Release()
	defer c.shapeOddPipeline.Release()

	defer c.contourMesh.Release()
	defer c.contourMaterial.Release()
	defer c.contourPipeline.Release()

	defer c.textMesh.Release()
	defer c.textMaterial.Release()
	defer c.textPipeline.Release()
}

func (c *canvasRenderer) onBegin(size Size) {
	c.currentLayer = c.topLayer
	c.SetTransform(sprec.IdentityMat4())
	c.SetClipRect(0, float32(size.Width), 0, float32(size.Height))
	c.projectionMatrix = sprec.OrthoMat4(
		0.0, float32(size.Width),
		0.0, float32(size.Height),
		0.0, 1.0,
	)

	c.shapeMesh.Reset()
	c.contourMesh.Reset()
	c.textMesh.Reset()
}

func (c *canvasRenderer) onEnd() {
	c.shapeMesh.Update()
	c.contourMesh.Update()
	c.textMesh.Update()
	c.api.SubmitQueue(c.commandQueue)
}

// Push records the current state and creates a new state layer. Changes done
// in the new layer will not affect the parent layer.
//
// You may create up to 256 layers including the starting one after which the
// method panics.
func (c *canvasRenderer) Push() {
	c.currentLayer = c.currentLayer.Next()
}

// Pop restores the drawing state based on the parent layer. If this is the
// first layer, then this method panics.
func (c *canvasRenderer) Pop() {
	c.currentLayer = c.currentLayer.Previous()
}

// ResetTransform restores the transform to the value it had
// after the last Push. If this is the first layer, then it is
// set to the identity matrix.
func (c *canvasRenderer) ResetTransform() {
	if c.currentLayer == c.topLayer {
		c.currentLayer.Transform = sprec.IdentityMat4()
	} else {
		c.currentLayer.Transform = c.currentLayer.previous.Transform
	}
}

// SetTransform changes the transform relative to the former layer transform.
func (c *canvasRenderer) SetTransform(transform sprec.Mat4) {
	if c.currentLayer == c.topLayer {
		c.currentLayer.Transform = transform
	} else {
		c.currentLayer.Transform = sprec.Mat4Prod(
			c.currentLayer.previous.Transform,
			transform,
		)
	}
}

// Translate moves the drawing position by the specified delta amount.
func (c *canvasRenderer) Translate(delta sprec.Vec2) {
	c.currentLayer.Transform = sprec.Mat4Prod(
		c.currentLayer.Transform,
		sprec.TranslationMat4(delta.X, delta.Y, 0.0),
	)
}

// Rotate rotates the drawing coordinate system by the specified angle.
func (c *canvasRenderer) Rotate(angle sprec.Angle) {
	c.currentLayer.Transform = sprec.Mat4Prod(
		c.currentLayer.Transform,
		sprec.RotationMat4(angle, 0.0, 0.0, 1.0),
	)
}

// Scale scales the drawing coordinate system by the specified amount in
// both directions.
func (c *canvasRenderer) Scale(amount sprec.Vec2) {
	c.currentLayer.Transform = sprec.Mat4Prod(
		c.currentLayer.Transform,
		sprec.ScaleMat4(amount.X, amount.Y, 1.0),
	)
}

// SetClipRect creates a clipping rectangle region. This clipping mechanism
// is slighly faster than using Clip with a Path and is used by the UI framework
// for clipping Element contents.
//
// Note: This clipping model does not nest, hence you can escape the boundaries
// of your Element depending on the provided values. In most cases, the Clip
// method should be used instead.
func (c *canvasRenderer) SetClipRect(left, right, top, bottom float32) {
	c.currentLayer.ClipTransform = sprec.Mat4Prod(
		sprec.NewMat4(
			1.0, 0.0, 0.0, -left,
			-1.0, 0.0, 0.0, right,
			0.0, 1.0, 0.0, -top,
			0.0, -1.0, 0.0, bottom,
		),
		sprec.InverseMat4(c.currentLayer.Transform),
	)
}

// DrawSurface renders the specified surface. The surface's Render
// method will be called when needed and is expected to return a texture
// representing the rendered frame.
func (c *canvasRenderer) DrawSurface(surface Surface, position, size sprec.Vec2) {
	texture, texSize := surface.Render()

	c.Reset()
	c.Rectangle(
		position,
		size,
	)
	c.Fill(Fill{
		Rule: FillRuleSimple,
		Image: &Image{ // TODO: Don't allocate
			texture: texture,
			size:    texSize,
		},
		Color:       White(),
		ImageOffset: sprec.NewVec2(0.0, size.Y),
		ImageSize:   sprec.NewVec2(size.X, -size.Y),
	})
}

// FillText draws a solid text at the specified position using the provided
// typography settings.
func (c *canvasRenderer) FillText(text string, position sprec.Vec2, typography Typography) {
	currentLayer := c.currentLayer
	transformMatrix := currentLayer.Transform
	clipMatrix := currentLayer.ClipTransform

	font := typography.Font
	fontSize := typography.Size
	color := uiColorToVec(typography.Color)

	vertexOffset := c.textMesh.Offset()
	offset := position
	lastGlyph := (*fontGlyph)(nil)

	for _, ch := range text {
		lineHeight := font.lineHeight * fontSize
		lineAscent := font.lineAscent * fontSize
		if ch == '\r' {
			offset.X = position.X
			lastGlyph = nil
			continue
		}
		if ch == '\n' {
			offset.X = position.X
			offset.Y += lineHeight
			lastGlyph = nil
			continue
		}

		if glyph, ok := font.glyphs[ch]; ok {
			advance := glyph.advance * fontSize
			leftBearing := glyph.leftBearing * fontSize
			rightBearing := glyph.rightBearing * fontSize
			ascent := glyph.ascent * fontSize
			descent := glyph.descent * fontSize

			vertTopLeft := textVertex{
				position: sprec.Vec2Sum(
					sprec.NewVec2(
						leftBearing,
						lineAscent-ascent,
					),
					offset,
				),
				texCoord: sprec.NewVec2(glyph.leftU, glyph.topV),
			}
			vertTopRight := textVertex{
				position: sprec.Vec2Sum(
					sprec.NewVec2(
						advance-rightBearing,
						lineAscent-ascent,
					),
					offset,
				),
				texCoord: sprec.NewVec2(glyph.rightU, glyph.topV),
			}
			vertBottomLeft := textVertex{
				position: sprec.Vec2Sum(
					sprec.NewVec2(
						leftBearing,
						lineAscent+descent,
					),
					offset,
				),
				texCoord: sprec.NewVec2(glyph.leftU, glyph.bottomV),
			}
			vertBottomRight := textVertex{
				position: sprec.Vec2Sum(
					sprec.NewVec2(
						advance-rightBearing,
						lineAscent+descent,
					),
					offset,
				),
				texCoord: sprec.NewVec2(glyph.rightU, glyph.bottomV),
			}

			c.textMesh.Append(vertTopLeft)
			c.textMesh.Append(vertBottomLeft)
			c.textMesh.Append(vertBottomRight)

			c.textMesh.Append(vertTopLeft)
			c.textMesh.Append(vertBottomRight)
			c.textMesh.Append(vertTopRight)

			offset.X += advance
			if lastGlyph != nil {
				offset.X += lastGlyph.kerns[ch] * fontSize
			}
			lastGlyph = glyph
		}
	}
	vertexCount := c.textMesh.Offset() - vertexOffset

	c.commandQueue.BindPipeline(c.textPipeline)
	c.commandQueue.Uniform4f(c.textMaterial.colorLocation, color.Array())
	c.commandQueue.UniformMatrix4f(c.textMaterial.projectionMatrixLocation, c.projectionMatrix.ColumnMajorArray())
	c.commandQueue.UniformMatrix4f(c.textMaterial.transformMatrixLocation, transformMatrix.ColumnMajorArray())
	c.commandQueue.UniformMatrix4f(c.textMaterial.clipMatrixLocation, clipMatrix.ColumnMajorArray())
	c.commandQueue.TextureUnit(0, font.texture)
	c.commandQueue.Uniform1i(c.textMaterial.textureLocation, 0)
	c.commandQueue.Draw(vertexOffset, vertexCount, 1)
}

// Fill fills the currently constructed Path according to the fill settings.
func (c *canvasRenderer) Fill(fill Fill) {
	c.fillPath(c.canvasPath, fill)
}

// Stroke outlines the currently constructed Path.
func (c *canvasRenderer) Stroke() {
	c.strokePath(c.canvasPath)
}

// Clip creates a new clipping area according to the currently constructed Path
// and the clip area of parent layers.
func (c *canvasRenderer) Clip() {
	c.clipPath(c.canvasPath)
}

func (c *canvasRenderer) fillPath(path *canvasPath, fill Fill) {
	currentLayer := c.currentLayer
	transformMatrix := currentLayer.Transform
	clipMatrix := currentLayer.ClipTransform
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
	for _, point := range path.points {
		c.shapeMesh.Append(shapeVertex{
			position: point.coords,
		})
	}

	// draw stencil mask for all sub-shapes
	if fill.Rule != FillRuleSimple {
		c.commandQueue.BindPipeline(c.shapeMaskPipeline)
		c.commandQueue.UniformMatrix4f(c.shapeBlankMaterial.projectionMatrixLocation, c.projectionMatrix.ColumnMajorArray())
		c.commandQueue.UniformMatrix4f(c.shapeBlankMaterial.transformMatrixLocation, transformMatrix.ColumnMajorArray())
		c.commandQueue.UniformMatrix4f(c.shapeBlankMaterial.clipMatrixLocation, clipMatrix.ColumnMajorArray())

		for i, pointOffset := range path.subPathOffsets {
			pointCount := len(path.points) - pointOffset
			if i+1 < len(path.subPathOffsets) {
				pointCount = path.subPathOffsets[i+1] - pointOffset
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

	texture := c.whiteMask
	if fill.Image != nil {
		texture = fill.Image.texture
	}
	color := uiColorToVec(fill.Color)

	c.commandQueue.Uniform4f(c.shapeShadeMaterial.colorLocation, color.Array())
	c.commandQueue.UniformMatrix4f(c.shapeShadeMaterial.projectionMatrixLocation, c.projectionMatrix.ColumnMajorArray())
	c.commandQueue.UniformMatrix4f(c.shapeShadeMaterial.transformMatrixLocation, transformMatrix.ColumnMajorArray())
	c.commandQueue.UniformMatrix4f(c.shapeShadeMaterial.clipMatrixLocation, clipMatrix.ColumnMajorArray())
	c.commandQueue.UniformMatrix4f(c.shapeShadeMaterial.textureTransformMatrixLocation, textureTransformMatrix.ColumnMajorArray())
	c.commandQueue.TextureUnit(0, texture)
	c.commandQueue.Uniform1i(c.shapeShadeMaterial.textureLocation, 0)

	for i, pointOffset := range path.subPathOffsets {
		pointCount := len(path.points) - pointOffset
		if i+1 < len(path.subPathOffsets) {
			pointCount = path.subPathOffsets[i+1] - pointOffset
		}
		c.commandQueue.Draw(vertexOffset+pointOffset, pointCount, 1)
	}
}

func (c *canvasRenderer) strokePath(path *canvasPath) {
	currentLayer := c.currentLayer
	transformMatrix := currentLayer.Transform
	clipMatrix := currentLayer.ClipTransform

	c.commandQueue.BindPipeline(c.contourPipeline)
	c.commandQueue.UniformMatrix4f(c.contourMaterial.projectionMatrixLocation, c.projectionMatrix.ColumnMajorArray())
	c.commandQueue.UniformMatrix4f(c.contourMaterial.transformMatrixLocation, transformMatrix.ColumnMajorArray())
	c.commandQueue.UniformMatrix4f(c.contourMaterial.clipMatrixLocation, clipMatrix.ColumnMajorArray())
	c.commandQueue.Uniform1i(c.contourMaterial.textureLocation, 0)

	for i, pointOffset := range path.subPathOffsets {
		pointCount := len(path.points) - pointOffset
		if i+1 < len(path.subPathOffsets) {
			pointCount = path.subPathOffsets[i+1] - pointOffset
		}

		pointIndex := pointOffset
		current := c.points[pointIndex]
		next := c.points[pointIndex+1]
		currentNormal := endPointNormal(
			current.coords,
			next.coords,
		)
		pointIndex++

		vertexOffset := c.contourMesh.Offset()
		for pointIndex < pointOffset+pointCount {
			prev := current
			prevNormal := currentNormal

			current = c.points[pointIndex]
			if pointIndex != pointOffset+pointCount-1 {
				next := c.points[pointIndex+1]
				currentNormal = midPointNormal(
					prev.coords,
					current.coords,
					next.coords,
				)
			} else {
				currentNormal = endPointNormal(
					prev.coords,
					current.coords,
				)
			}

			prevLeft := contourVertex{
				position: sprec.Vec2Sum(prev.coords, sprec.Vec2Prod(prevNormal, prev.innerSize)),
				color:    prev.color,
			}
			prevRight := contourVertex{
				position: sprec.Vec2Diff(prev.coords, sprec.Vec2Prod(prevNormal, prev.outerSize)),
				color:    prev.color,
			}
			currentLeft := contourVertex{
				position: sprec.Vec2Sum(current.coords, sprec.Vec2Prod(currentNormal, current.innerSize)),
				color:    current.color,
			}
			currentRight := contourVertex{
				position: sprec.Vec2Diff(current.coords, sprec.Vec2Prod(currentNormal, current.outerSize)),
				color:    current.color,
			}

			c.contourMesh.Append(prevLeft)
			c.contourMesh.Append(prevRight)
			c.contourMesh.Append(currentLeft)

			c.contourMesh.Append(prevRight)
			c.contourMesh.Append(currentRight)
			c.contourMesh.Append(currentLeft)

			pointIndex++
		}
		vertexCount := c.contourMesh.Offset() - vertexOffset

		c.commandQueue.Draw(vertexOffset, vertexCount, 1)
	}
}

func (c *canvasRenderer) clipPath(path *canvasPath) {
	// TODO: This can be achieved if depth attachment is used.
	// One pass is drawn with the path where the a stencil mask is written.
	// Then, another pass uses the stencil mask to write a new depth level
	// and erase the stencil mask.
	// All rendering operations will need to be adjusted to perform depth tests
	// with mode EQUAL and the given layer's (or iteration's) depth value.
	// NOTE: How does Pop work with this approach? Do we redraw the clip (meaning
	// we need to keep track of it)?
}

// Typography configures how text is to be drawn.
type Typography struct {

	// Font specifies the font to be used.
	Font *Font

	// Size specifies the font size.
	Size float32

	// Color indicates the color of the text.
	Color Color
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
	Render() (render.Texture, Size)
}

func uiColorToVec(color Color) sprec.Vec4 {
	return sprec.NewVec4(
		float32(color.R)/255.0,
		float32(color.G)/255.0,
		float32(color.B)/255.0,
		float32(color.A)/255.0,
	)
}

func midPointNormal(prev, middle, next sprec.Vec2) sprec.Vec2 {
	normal1 := endPointNormal(prev, middle)
	normal2 := endPointNormal(middle, next)
	normalSum := sprec.Vec2Sum(normal1, normal2)
	dot := sprec.Vec2Dot(normal1, normalSum)
	return sprec.Vec2Quot(normalSum, dot)
}

func endPointNormal(prev, next sprec.Vec2) sprec.Vec2 {
	tangent := sprec.UnitVec2(sprec.Vec2Diff(next, prev))
	return sprec.NewVec2(tangent.Y, -tangent.X)
}
