package ui

import (
	"github.com/mokiat/gblob"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

const (
	maxLayerDepth = 256

	maxVertexCount = 524288
)

const (
	uniformBufferBindingCamera = iota
	uniformBufferBindingModel
	uniformBufferBindingMaterial
)

const (
	textureBindingColorTexture = iota
	textureBindingFontTexture
)

func newCanvasRenderer(api render.API, shaders ShaderCollection) *canvasRenderer {
	return &canvasRenderer{
		canvasPath: newCanvasPath(),

		api: api,

		shapeMesh: newShapeMesh(maxVertexCount),
		shapeShadeMaterial: newMaterial(render.ProgramInfo{
			Label:      "Shaded Shape",
			SourceCode: shaders.ShapeShadedSet(),
			TextureBindings: []render.TextureBinding{
				render.NewTextureBinding("colorTextureIn", textureBindingColorTexture),
			},
			UniformBindings: []render.UniformBinding{
				render.NewUniformBinding("Camera", uniformBufferBindingCamera),
				render.NewUniformBinding("Model", uniformBufferBindingModel),
				render.NewUniformBinding("Material", uniformBufferBindingMaterial),
			},
		}),
		shapeBlankMaterial: newMaterial(render.ProgramInfo{
			Label:           "Blank Shape",
			SourceCode:      shaders.ShapeBlankSet(),
			TextureBindings: []render.TextureBinding{},
			UniformBindings: []render.UniformBinding{
				render.NewUniformBinding("Camera", uniformBufferBindingCamera),
				render.NewUniformBinding("Model", uniformBufferBindingModel),
			},
		}),

		contourMesh: newContourMesh(maxVertexCount),
		contourMaterial: newMaterial(render.ProgramInfo{
			Label:           "Contour Material",
			SourceCode:      shaders.ContourSet(),
			TextureBindings: []render.TextureBinding{},
			UniformBindings: []render.UniformBinding{
				render.NewUniformBinding("Camera", uniformBufferBindingCamera),
				render.NewUniformBinding("Model", uniformBufferBindingModel),
			},
		}),

		textMesh: newTextMesh(maxVertexCount),
		textMaterial: newMaterial(render.ProgramInfo{
			Label:      "Text Material",
			SourceCode: shaders.TextSet(),
			TextureBindings: []render.TextureBinding{
				render.NewTextureBinding("fontTextureIn", textureBindingFontTexture),
			},
			UniformBindings: []render.UniformBinding{
				render.NewUniformBinding("Camera", uniformBufferBindingCamera),
				render.NewUniformBinding("Model", uniformBufferBindingModel),
				render.NewUniformBinding("Material", uniformBufferBindingMaterial),
			},
		}),

		topLayer: &canvasLayer{},
	}
}

type canvasRenderer struct {
	*canvasPath

	api          render.API
	commandQueue render.CommandQueue

	cameraUniformBufferData gblob.LittleEndianBlock
	cameraUniformBuffer     render.Buffer

	modelUniformBufferData gblob.LittleEndianBlock
	modelUniformBuffer     render.Buffer
	modelIsDirty           bool

	materialUniformBufferData gblob.LittleEndianBlock
	materialUniformBuffer     render.Buffer

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

	topLayer     *canvasLayer
	currentLayer *canvasLayer
}

func (c *canvasRenderer) onCreate() {
	c.commandQueue = c.api.CreateCommandQueue()

	c.cameraUniformBufferData = make([]byte, 64) // 1 x mat4
	c.cameraUniformBuffer = c.api.CreateUniformBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(c.cameraUniformBufferData),
	})

	c.modelUniformBufferData = make([]byte, 2*64) // 2 x mat4
	c.modelUniformBuffer = c.api.CreateUniformBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(c.modelUniformBufferData),
	})

	c.materialUniformBufferData = make([]byte, 64+16) // 1 x mat4 + 1 x vec4
	c.materialUniformBuffer = c.api.CreateUniformBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(c.materialUniformBufferData),
	})

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
	c.shapeBlankMaterial.Allocate(c.api)
	c.shapeMaskPipeline = c.api.CreatePipeline(render.PipelineInfo{
		Program:     c.shapeBlankMaterial.program,
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
	c.shapeShadeMaterial.Allocate(c.api)
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

	defer c.cameraUniformBuffer.Release()
	defer c.modelUniformBuffer.Release()
	defer c.materialUniformBuffer.Release()

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
	c.commandQueue.UniformBufferUnit(uniformBufferBindingCamera, c.cameraUniformBuffer)
	c.commandQueue.UniformBufferUnit(uniformBufferBindingModel, c.modelUniformBuffer)
	c.commandQueue.UniformBufferUnit(uniformBufferBindingMaterial, c.materialUniformBuffer)

	c.updateCameraUniformBuffer(size)

	c.modelIsDirty = true
	c.currentLayer = c.topLayer
	c.SetTransform(sprec.IdentityMat4())
	c.SetClipRect(0, float32(size.Width), 0, float32(size.Height))

	c.shapeMesh.Reset()
	c.contourMesh.Reset()
	c.textMesh.Reset()
}

func (c *canvasRenderer) onEnd() {
	c.api.Invalidate()
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
	c.modelIsDirty = true
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
	c.modelIsDirty = true
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
	c.modelIsDirty = true
}

// Translate moves the drawing position by the specified delta amount.
func (c *canvasRenderer) Translate(delta sprec.Vec2) {
	c.currentLayer.Transform = sprec.Mat4Prod(
		c.currentLayer.Transform,
		sprec.TranslationMat4(delta.X, delta.Y, 0.0),
	)
	c.modelIsDirty = true
}

// Rotate rotates the drawing coordinate system by the specified angle.
func (c *canvasRenderer) Rotate(angle sprec.Angle) {
	c.currentLayer.Transform = sprec.Mat4Prod(
		c.currentLayer.Transform,
		sprec.RotationMat4(angle, 0.0, 0.0, 1.0),
	)
	c.modelIsDirty = true
}

// Scale scales the drawing coordinate system by the specified amount in
// both directions.
func (c *canvasRenderer) Scale(amount sprec.Vec2) {
	c.currentLayer.Transform = sprec.Mat4Prod(
		c.currentLayer.Transform,
		sprec.ScaleMat4(amount.X, amount.Y, 1.0),
	)
	c.modelIsDirty = true
}

// ClipRect creates a clipping rectangle region. This clipping mechanism
// is slighly faster than using Clip with a Path and is used by the UI framework
// for clipping Element contents.
//
// Note: This clipping model does not nest, hence you can escape the boundaries
// of your Element depending on the provided values. In most cases, the Clip
// method should be used instead.
func (c *canvasRenderer) ClipRect(position, size sprec.Vec2) {
	// TODO: Make this apply on top of parent clip rects, so that an element
	// cannot escape its bounds.

	c.currentLayer.ClipTransform = sprec.Mat4Prod(
		sprec.NewMat4(
			1.0, 0.0, 0.0, -position.X,
			-1.0, 0.0, 0.0, position.X+size.X,
			0.0, 1.0, 0.0, -position.Y,
			0.0, -1.0, 0.0, position.Y+size.Y,
		),
		sprec.InverseMat4(c.currentLayer.Transform),
	)
	c.modelIsDirty = true
}

// SetClipRect creates a clipping rectangle region. This clipping mechanism
// is slighly faster than using Clip with a Path and is used by the UI framework
// for clipping Element contents.
//
// Note: This clipping model does not nest, hence you can escape the boundaries
// of your Element depending on the provided values. In most cases, the Clip
// method should be used instead.
//
// Deprecated: Use ClipRect instead
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
	c.modelIsDirty = true
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

// FillTextLine draws a solid text line at the specified position using the
// provided typography settings.
func (c *canvasRenderer) FillTextLine(text []rune, position sprec.Vec2, typography Typography) {
	font := typography.Font
	fontSize := typography.Size

	vertexOffset := c.textMesh.Offset()
	offset := position
	lastGlyph := (*fontGlyph)(nil)

	for _, ch := range text {
		lineAscent := font.lineAscent * fontSize

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
	if vertexCount > 0 {
		c.updateModelUniformBuffer(c.currentLayer)
		c.updateMaterialUniformBufferFromTypography(typography)
		c.commandQueue.BindPipeline(c.textPipeline)
		c.commandQueue.TextureUnit(textureBindingFontTexture, font.texture)
		c.commandQueue.Draw(vertexOffset, vertexCount, 1)
	}
}

// FillText draws a solid text at the specified position using the provided
// typography settings.
//
// Deprecated: Use FillTextLine
func (c *canvasRenderer) FillText(text string, position sprec.Vec2, typography Typography) {
	font := typography.Font
	fontSize := typography.Size

	vertexOffset := c.textMesh.Offset()
	offset := position
	lastGlyph := (*fontGlyph)(nil)

	for _, ch := range text {
		lineHeight := font.lineHeight * fontSize
		lineAscent := font.lineAscent * fontSize
		if ch == '\r' {
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
	if vertexCount > 0 {
		c.updateModelUniformBuffer(c.currentLayer)
		c.updateMaterialUniformBufferFromTypography(typography)

		c.commandQueue.BindPipeline(c.textPipeline)
		c.commandQueue.TextureUnit(textureBindingFontTexture, font.texture)
		c.commandQueue.Draw(vertexOffset, vertexCount, 1)
	}
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
	vertexOffset := c.shapeMesh.Offset()
	for _, point := range path.points {
		c.shapeMesh.Append(shapeVertex{
			position: point.coords,
		})
	}

	c.updateModelUniformBuffer(c.currentLayer)

	// draw stencil mask for all sub-shapes
	if fill.Rule != FillRuleSimple {
		c.commandQueue.BindPipeline(c.shapeMaskPipeline)
		for i, pointOffset := range path.subPathOffsets {
			pointCount := len(path.points) - pointOffset
			if i+1 < len(path.subPathOffsets) {
				pointCount = path.subPathOffsets[i+1] - pointOffset
			}
			c.commandQueue.Draw(vertexOffset+pointOffset, pointCount, 1)
		}
	}

	c.updateMaterialUniformBufferFromFill(fill)

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

	c.commandQueue.TextureUnit(textureBindingColorTexture, texture)
	for i, pointOffset := range path.subPathOffsets {
		pointCount := len(path.points) - pointOffset
		if i+1 < len(path.subPathOffsets) {
			pointCount = path.subPathOffsets[i+1] - pointOffset
		}
		c.commandQueue.Draw(vertexOffset+pointOffset, pointCount, 1)
	}
}

func (c *canvasRenderer) strokePath(path *canvasPath) {
	if len(path.points) == 0 {
		return
	}

	c.updateModelUniformBuffer(c.currentLayer)
	c.commandQueue.BindPipeline(c.contourPipeline)

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

func (c *canvasRenderer) updateCameraUniformBuffer(size Size) {
	cameraPlotter := blob.NewPlotter(c.cameraUniformBufferData)
	cameraPlotter.PlotSPMat4(sprec.OrthoMat4(
		0.0, float32(size.Width),
		0.0, float32(size.Height),
		0.0, 1.0,
	))
	c.commandQueue.UpdateBufferData(c.cameraUniformBuffer, render.BufferUpdateInfo{
		Data: c.cameraUniformBufferData,
	})
}

func (c *canvasRenderer) updateModelUniformBuffer(layer *canvasLayer) {
	if c.modelIsDirty {
		modelPlotter := blob.NewPlotter(c.modelUniformBufferData)
		modelPlotter.PlotSPMat4(layer.Transform)
		modelPlotter.PlotSPMat4(layer.ClipTransform)
		c.commandQueue.UpdateBufferData(c.modelUniformBuffer, render.BufferUpdateInfo{
			Data: c.modelUniformBufferData,
		})
		c.modelIsDirty = false
	}
}

func (c *canvasRenderer) updateMaterialUniformBufferFromFill(fill Fill) {
	materialPlotter := blob.NewPlotter(c.materialUniformBufferData)
	materialPlotter.PlotSPMat4(sprec.Mat4MultiProd(
		sprec.ScaleMat4(1.0/fill.ImageSize.X, 1.0/fill.ImageSize.Y, 1.0),
		sprec.TranslationMat4(-fill.ImageOffset.X, -fill.ImageOffset.Y, 0.0),
	))
	materialPlotter.PlotSPVec4(uiColorToVec(fill.Color))

	c.commandQueue.UpdateBufferData(c.materialUniformBuffer, render.BufferUpdateInfo{
		Data: c.materialUniformBufferData,
	})
}

func (c *canvasRenderer) updateMaterialUniformBufferFromTypography(typography Typography) {
	materialPlotter := blob.NewPlotter(c.materialUniformBufferData)
	materialPlotter.Skip(64)
	materialPlotter.PlotSPVec4(uiColorToVec(typography.Color))

	c.commandQueue.UpdateBufferData(c.materialUniformBuffer, render.BufferUpdateInfo{
		Data: c.materialUniformBufferData,
	})
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
