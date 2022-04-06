package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/ui/renderapi/plugin"
)

const maxVertexCount = 524288

func NewRenderer(api render.API, shaders plugin.ShaderCollection) *Renderer {
	return &Renderer{
		api:                api,
		shape:              newShape(),
		shapeMesh:          newShapeMesh(maxVertexCount),
		shapeMaterial:      newMaterial(shaders.ShapeMaterial),
		shapeBlankMaterial: newMaterial(shaders.ShapeBlankMaterial),
		contour:            newContour(),
		contourMesh:        newContourMesh(maxVertexCount),
		contourMaterial:    newMaterial(shaders.ContourMaterial),
		text:               newText(),
		textMesh:           newTextMesh(maxVertexCount),
		textMaterial:       newMaterial(shaders.TextMaterial),
	}
}

type Renderer struct {
	api render.API

	commandQueue render.CommandQueue

	projectionMatrix       sprec.Mat4
	transformMatrix        sprec.Mat4
	textureTransformMatrix sprec.Mat4
	clipBounds             sprec.Vec4

	shape              *Shape
	shapeMesh          *ShapeMesh
	shapeMaterial      *Material
	shapeBlankMaterial *Material

	shapeMaskPipeline             render.Pipeline
	shapeShadeModeNonePipeline    render.Pipeline
	shapeShadeModeNonZeroPipeline render.Pipeline
	shapeShadeModeOddPipeline     render.Pipeline

	contour         *Contour
	contourMesh     *ContourMesh
	contourMaterial *Material

	text         *Text
	textMesh     *TextMesh
	textMaterial *Material
	textPipeline render.Pipeline

	whiteMask render.Texture

	subMeshes []SubMesh

	target Target
}

func (r *Renderer) Init() {
	r.commandQueue = r.api.CreateCommandQueue()

	r.shapeMesh.Allocate(r.api)
	r.shapeMaterial.Allocate(r.api)
	r.shapeBlankMaterial.Allocate(r.api)
	r.contourMesh.Allocate(r.api)
	r.contourMaterial.Allocate(r.api)
	r.textMesh.Allocate(r.api)
	r.textMaterial.Allocate(r.api)

	r.shapeMaskPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.shapeBlankMaterial.program,
		VertexArray:     r.shapeMesh.vertexArray,
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
		ColorWrite:                  [4]bool{false, false, false, false},
		BlendEnabled:                false,
		BlendColor:                  sprec.ZeroVec4(),
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
	r.shapeShadeModeNonePipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:                     r.shapeMaterial.program,
		VertexArray:                 r.shapeMesh.vertexArray,
		Topology:                    render.TopologyTriangleFan,
		Culling:                     render.CullModeNone,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		StencilTest:                 false,
		ColorWrite:                  [4]bool{true, true, true, true},
		BlendEnabled:                true,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
	r.shapeShadeModeNonZeroPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.shapeMaterial.program,
		VertexArray:     r.shapeMesh.vertexArray,
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
		ColorWrite:                  [4]bool{true, true, true, true},
		BlendEnabled:                true,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
	r.shapeShadeModeOddPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:     r.shapeMaterial.program,
		VertexArray: r.shapeMesh.vertexArray,
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
		ColorWrite:                  [4]bool{true, true, true, true},
		BlendEnabled:                true,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})

	r.textPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:                     r.textMaterial.program,
		VertexArray:                 r.textMesh.vertexArray,
		Topology:                    render.TopologyTriangles,
		Culling:                     render.CullModeNone,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		StencilTest:                 false,
		ColorWrite:                  [4]bool{true, true, true, true},
		BlendEnabled:                true,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorSourceAlpha,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})

	r.whiteMask = r.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           1,
		Height:          1,
		Filtering:       render.FilterModeNearest,
		Wrapping:        render.WrapModeClamp,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
		Data:            []byte{0xFF, 0xFF, 0xFF, 0xFF},
	})
}

func (r *Renderer) Free() {
	defer r.commandQueue.Release()
	defer r.shapeMesh.Release()
	defer r.shapeMaterial.Release()
	defer r.shapeBlankMaterial.Release()
	defer r.contourMesh.Release()
	defer r.contourMaterial.Release()
	defer r.textMesh.Release()
	defer r.textMaterial.Release()
	defer r.textPipeline.Release()
	defer r.whiteMask.Release()
}

func (r *Renderer) Transform() sprec.Mat4 {
	return r.transformMatrix
}

func (r *Renderer) SetTransform(transform sprec.Mat4) {
	r.transformMatrix = transform
}

func (r *Renderer) TextureTransform() sprec.Mat4 {
	return r.textureTransformMatrix
}

func (r *Renderer) SetTextureTransform(textureTransform sprec.Mat4) {
	r.textureTransformMatrix = textureTransform
}

func (r *Renderer) ClipBounds() (left, right, top, bottom float32) {
	return r.clipBounds.X, r.clipBounds.Y, r.clipBounds.Z, r.clipBounds.W
}

func (r *Renderer) SetClipBounds(left, right, top, bottom float32) {
	r.clipBounds = sprec.NewVec4(left, right, top, bottom)
}

func (r *Renderer) BeginShape(fill Fill) *Shape {
	if r.shape == nil {
		panic("shape already started")
	}
	result := r.shape
	result.Init(fill)
	r.shape = nil
	return result
}

func (r *Renderer) EndShape(shape *Shape) {
	if r.shape != nil {
		panic("shape already ended")
	}
	r.shape = shape

	vertexOffset := r.shapeMesh.Offset()
	for _, point := range shape.points {
		r.shapeMesh.Append(ShapeVertex{
			position: point.coords,
		})
	}

	// draw stencil mask for all sub-shapes
	if shape.fill.mode != StencilModeNone {
		r.commandQueue.BindPipeline(r.shapeMaskPipeline)
		r.commandQueue.Uniform4f(r.shapeBlankMaterial.clipDistancesLocation, [4]float32{
			r.clipBounds.X, r.clipBounds.Y, r.clipBounds.Z, r.clipBounds.W, // TODO: Add Array method to Vec4
		})
		r.commandQueue.UniformMatrix4f(r.shapeBlankMaterial.projectionMatrixLocation, r.projectionMatrix.ColumnMajorArray())
		r.commandQueue.UniformMatrix4f(r.shapeBlankMaterial.transformMatrixLocation, r.transformMatrix.ColumnMajorArray())

		for _, subShape := range shape.subShapes {
			r.commandQueue.Draw(vertexOffset+subShape.pointOffset, subShape.pointCount, 1)
		}
	}

	// render color shader shape and clear stencil buffer
	switch shape.fill.mode {
	case StencilModeNone:
		r.commandQueue.BindPipeline(r.shapeShadeModeNonePipeline)
	case StencilModeNonZero:
		r.commandQueue.BindPipeline(r.shapeShadeModeNonZeroPipeline)
	case StencilModeOdd:
		r.commandQueue.BindPipeline(r.shapeShadeModeOddPipeline)
	default:
		r.commandQueue.BindPipeline(r.shapeShadeModeNonePipeline)
	}

	texture := r.whiteMask
	if shape.fill.image != nil {
		texture = shape.fill.image.texture
	}

	r.commandQueue.Uniform4f(r.shapeMaterial.clipDistancesLocation, [4]float32{
		r.clipBounds.X, r.clipBounds.Y, r.clipBounds.Z, r.clipBounds.W, // TODO: Add Array method to Vec4
	})
	r.commandQueue.Uniform4f(r.shapeMaterial.colorLocation, [4]float32{
		shape.fill.color.X, shape.fill.color.Y, shape.fill.color.Z, shape.fill.color.W, // TODO: Add Array method to Vec4
	})
	r.commandQueue.UniformMatrix4f(r.shapeMaterial.projectionMatrixLocation, r.projectionMatrix.ColumnMajorArray())
	r.commandQueue.UniformMatrix4f(r.shapeMaterial.transformMatrixLocation, r.transformMatrix.ColumnMajorArray())
	r.commandQueue.UniformMatrix4f(r.shapeMaterial.textureTransformMatrixLocation, r.textureTransformMatrix.ColumnMajorArray())
	r.commandQueue.TextureUnit(0, texture)
	r.commandQueue.Uniform1i(r.shapeMaterial.textureLocation, 0)

	for _, subShape := range shape.subShapes {
		r.commandQueue.Draw(vertexOffset+subShape.pointOffset, subShape.pointCount, 1)
	}
}

func (r *Renderer) BeginContour() *Contour {
	if r.contour == nil {
		panic("contour already started")
	}
	result := r.contour
	result.Init()
	r.contour = nil
	return result
}

func (r *Renderer) EndContour(contour *Contour) {
	if r.contour != nil {
		panic("contour already ended")
	}
	r.contour = contour

	for _, subContour := range contour.subContours {
		pointIndex := subContour.pointOffset
		current := contour.points[pointIndex]
		next := contour.points[pointIndex+1]
		currentNormal := endPointNormal(
			current.coords,
			next.coords,
		)
		pointIndex++

		vertexOffset := r.contourMesh.Offset()
		for pointIndex < subContour.pointOffset+subContour.pointCount {
			prev := current
			prevNormal := currentNormal

			current = contour.points[pointIndex]
			if pointIndex != subContour.pointOffset+subContour.pointCount-1 {
				next := contour.points[pointIndex+1]
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

			prevLeft := ContourVertex{
				position: sprec.Vec2Sum(prev.coords, sprec.Vec2Prod(prevNormal, prev.stroke.innerSize)),
				color:    prev.stroke.color,
			}
			prevRight := ContourVertex{
				position: sprec.Vec2Diff(prev.coords, sprec.Vec2Prod(prevNormal, prev.stroke.outerSize)),
				color:    prev.stroke.color,
			}
			currentLeft := ContourVertex{
				position: sprec.Vec2Sum(current.coords, sprec.Vec2Prod(currentNormal, current.stroke.innerSize)),
				color:    prev.stroke.color,
			}
			currentRight := ContourVertex{
				position: sprec.Vec2Diff(current.coords, sprec.Vec2Prod(currentNormal, current.stroke.outerSize)),
				color:    prev.stroke.color,
			}

			r.contourMesh.Append(prevLeft)
			r.contourMesh.Append(prevRight)
			r.contourMesh.Append(currentLeft)

			r.contourMesh.Append(prevRight)
			r.contourMesh.Append(currentRight)
			r.contourMesh.Append(currentLeft)

			pointIndex++
		}
		vertexCount := r.contourMesh.Offset() - vertexOffset

		r.subMeshes = append(r.subMeshes, SubMesh{
			clipBounds:      r.clipBounds,
			material:        r.contourMaterial,
			vertexArray:     r.contourMesh.vertexArray,
			transformMatrix: r.transformMatrix,
			vertexOffset:    vertexOffset,
			vertexCount:     vertexCount,
			// primitive:       gl.TRIANGLES,
		})
	}
}

func (r *Renderer) BeginText(typography Typography) *Text {
	if r.text == nil {
		panic("text already started")
	}
	result := r.text
	result.Init(typography)
	r.text = nil
	return result
}

func (r *Renderer) EndText(text *Text) {
	if r.text != nil {
		panic("text already ended")
	}
	r.text = text

	vertexOffset := r.textMesh.Offset()
	for _, paragraph := range text.paragraphs {
		offset := paragraph.position
		lastGlyph := (*fontGlyph)(nil)

		paragraphChars := text.characters[paragraph.charOffset : paragraph.charOffset+paragraph.charCount]
		for _, ch := range paragraphChars {
			lineHeight := text.font.lineHeight * text.fontSize
			lineAscent := text.font.lineAscent * text.fontSize
			if ch == '\r' {
				offset.X = paragraph.position.X
				lastGlyph = nil
				continue
			}
			if ch == '\n' {
				offset.X = paragraph.position.X
				offset.Y += lineHeight
				lastGlyph = nil
				continue
			}

			if glyph, ok := text.font.glyphs[ch]; ok {
				advance := glyph.advance * text.fontSize
				leftBearing := glyph.leftBearing * text.fontSize
				rightBearing := glyph.rightBearing * text.fontSize
				ascent := glyph.ascent * text.fontSize
				descent := glyph.descent * text.fontSize

				vertTopLeft := TextVertex{
					position: sprec.Vec2Sum(
						sprec.NewVec2(
							leftBearing,
							lineAscent-ascent,
						),
						offset,
					),
					texCoord: sprec.NewVec2(glyph.leftU, glyph.topV),
				}
				vertTopRight := TextVertex{
					position: sprec.Vec2Sum(
						sprec.NewVec2(
							advance-rightBearing,
							lineAscent-ascent,
						),
						offset,
					),
					texCoord: sprec.NewVec2(glyph.rightU, glyph.topV),
				}
				vertBottomLeft := TextVertex{
					position: sprec.Vec2Sum(
						sprec.NewVec2(
							leftBearing,
							lineAscent+descent,
						),
						offset,
					),
					texCoord: sprec.NewVec2(glyph.leftU, glyph.bottomV),
				}
				vertBottomRight := TextVertex{
					position: sprec.Vec2Sum(
						sprec.NewVec2(
							advance-rightBearing,
							lineAscent+descent,
						),
						offset,
					),
					texCoord: sprec.NewVec2(glyph.rightU, glyph.bottomV),
				}

				r.textMesh.Append(vertTopLeft)
				r.textMesh.Append(vertBottomLeft)
				r.textMesh.Append(vertBottomRight)

				r.textMesh.Append(vertTopLeft)
				r.textMesh.Append(vertBottomRight)
				r.textMesh.Append(vertTopRight)

				offset.X += advance
				if lastGlyph != nil {
					offset.X += lastGlyph.kerns[ch] * text.fontSize
				}
				lastGlyph = glyph
			}
		}
	}
	vertexCount := r.textMesh.Offset() - vertexOffset

	r.commandQueue.BindPipeline(r.textPipeline)
	r.commandQueue.Uniform4f(r.textMaterial.clipDistancesLocation, [4]float32{
		r.clipBounds.X, r.clipBounds.Y, r.clipBounds.Z, r.clipBounds.W, // TODO: Add Array method to Vec4
	})
	r.commandQueue.Uniform4f(r.textMaterial.colorLocation, [4]float32{
		text.color.X, text.color.Y, text.color.Z, text.color.W, // TODO: Add Array method to Vec4
	})
	r.commandQueue.UniformMatrix4f(r.textMaterial.projectionMatrixLocation, r.projectionMatrix.ColumnMajorArray())
	r.commandQueue.UniformMatrix4f(r.textMaterial.transformMatrixLocation, r.transformMatrix.ColumnMajorArray())
	r.commandQueue.TextureUnit(0, text.font.texture)
	r.commandQueue.Uniform1i(r.textMaterial.textureLocation, 0)
	r.commandQueue.Draw(vertexOffset, vertexCount, 1)
}

func (r *Renderer) DrawSurface(surface Surface) {
	r.subMeshes = append(r.subMeshes, SubMesh{
		surface:    surface,
		clipBounds: r.clipBounds,
	})
}

func (r *Renderer) Begin(target Target) {
	r.target = target
	r.projectionMatrix = sprec.OrthoMat4(
		0.0, float32(target.Width),
		0.0, float32(target.Height),
		0.0, 1.0,
	)
	r.transformMatrix = sprec.IdentityMat4()
	r.textureTransformMatrix = sprec.IdentityMat4()
	r.clipBounds = sprec.NewVec4(
		0.0, float32(target.Width),
		0.0, float32(target.Height),
	)
	r.shapeMesh.Reset()
	r.contourMesh.Reset()
	r.textMesh.Reset()
	r.subMeshes = r.subMeshes[:0]

	r.api.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: target.Framebuffer,
		Viewport: render.Area{
			X:      0,
			Y:      0,
			Width:  target.Width,
			Height: target.Height,
		},
		DepthLoadOp:       render.LoadOperationDontCare,
		DepthStoreOp:      render.StoreOperationDontCare,
		StencilLoadOp:     render.LoadOperationClear,
		StencilStoreOp:    render.StoreOperationDontCare,
		StencilClearValue: 0x00,
	})
}

func (r *Renderer) End() {
	r.shapeMesh.Update()
	r.contourMesh.Update()
	r.textMesh.Update()

	r.api.SubmitQueue(r.commandQueue)
	r.api.EndRenderPass()

	// r.enableOptions()

	// // TODO: Maybe optimize by accumulating draw commands
	// // if they are similar.
	// for _, subMesh := range r.subMeshes {
	// 	if subMesh.surface != nil {
	// 		r.disableOptions()

	// 		x := int(subMesh.clipBounds.X)
	// 		y := int(subMesh.clipBounds.Z)
	// 		width := int(subMesh.clipBounds.Y - subMesh.clipBounds.X)
	// 		height := int(subMesh.clipBounds.W - subMesh.clipBounds.Z)
	// 		subMesh.surface.Render(
	// 			x,
	// 			r.target.Height-height-y,
	// 			width,
	// 			height,
	// 		)

	// 		r.enableOptions()
	// 		continue
	// 	}

	// 	material := subMesh.material
	// 	transformMatrix := subMesh.transformMatrix.ColumnMajorArray()
	// 	textureTransformMatrix := subMesh.textureTransformMatrix.ColumnMajorArray()

	// 	if subMesh.skipColor {
	// 		gl.ColorMask(false, false, false, false)
	// 	} else {
	// 		gl.ColorMask(true, true, true, true)
	// 	}
	// 	if subMesh.stencil {
	// 		gl.Enable(gl.STENCIL_TEST)

	// 		cfg := subMesh.stencilCfg
	// 		gl.StencilFuncSeparate(gl.FRONT, cfg.stencilFuncFront.fn, cfg.stencilFuncFront.ref, cfg.stencilFuncFront.mask)
	// 		gl.StencilFuncSeparate(gl.BACK, cfg.stencilFuncBack.fn, cfg.stencilFuncBack.ref, cfg.stencilFuncBack.mask)
	// 		gl.StencilOpSeparate(gl.FRONT, cfg.stencilOpFront.sfail, cfg.stencilOpFront.dpfail, cfg.stencilOpFront.dppass)
	// 		gl.StencilOpSeparate(gl.BACK, cfg.stencilOpBack.sfail, cfg.stencilOpBack.dpfail, cfg.stencilOpBack.dppass)
	// 	} else {
	// 		gl.Disable(gl.STENCIL_TEST)
	// 	}
	// 	if subMesh.culling {
	// 		gl.Enable(gl.CULL_FACE)
	// 		gl.CullFace(subMesh.cullFace)
	// 	} else {
	// 		gl.Disable(gl.CULL_FACE)
	// 	}
	// 	gl.UseProgram(material.program.ID())
	// 	gl.Uniform4f(material.clipDistancesLocation, subMesh.clipBounds.X, subMesh.clipBounds.Y, subMesh.clipBounds.Z, subMesh.clipBounds.W)
	// 	gl.UniformMatrix4fv(material.transformMatrixLocation, 1, false, &transformMatrix[0])
	// 	if material.textureTransformMatrixLocation != -1 {
	// 		gl.UniformMatrix4fv(material.textureTransformMatrixLocation, 1, false, &textureTransformMatrix[0])
	// 	}
	// 	gl.UniformMatrix4fv(material.projectionMatrixLocation, 1, false, &projectionMatrix[0])
	// 	if material.colorLocation != -1 {
	// 		gl.Uniform4f(material.colorLocation, subMesh.color.X, subMesh.color.Y, subMesh.color.Z, subMesh.color.W)
	// 	}
	// 	if material.textureLocation != -1 {
	// 		gl.BindTextureUnit(0, subMesh.texture.ID())
	// 		gl.Uniform1i(material.textureLocation, 0)
	// 	}
	// 	gl.BindVertexArray(subMesh.vertexArray.ID())
	// 	gl.DrawArrays(subMesh.primitive, int32(subMesh.vertexOffset), int32(subMesh.vertexCount))
	// }

	// r.disableOptions()
}

// func (r *Renderer) enableOptions() {
// 	// r.target.Framebuffer.Use()

// 	gl.Viewport(0, 0, int32(r.target.Width), int32(r.target.Height))
// 	gl.ClearStencil(0)
// 	gl.Clear(gl.STENCIL_BUFFER_BIT)
// 	// gl.Disable(gl.DEPTH_TEST)
// 	// gl.DepthMask(false)
// 	// gl.Enable(gl.BLEND)
// 	// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
// }

// func (r *Renderer) disableOptions() {
// 	gl.ColorMask(true, true, true, true)
// 	gl.Disable(gl.STENCIL_TEST)
// 	gl.Enable(gl.CULL_FACE)
// 	gl.CullFace(gl.BACK)
// 	gl.Disable(gl.BLEND)
// }

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

type Surface interface {
	Render(x, y, width, height int)
}
