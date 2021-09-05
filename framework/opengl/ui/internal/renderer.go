package internal

import (
	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
)

const maxVertexCount = 2048 * 10

func NewRenderer() *Renderer {
	return &Renderer{
		shape:           newShape(),
		shapeMaterial:   NewShapeMaterial(),
		contour:         newContour(),
		contourMaterial: nil, // TODO
		mesh:            NewMesh(maxVertexCount),
		whiteMask:       opengl.NewTwoDTexture(),
	}
}

type Renderer struct {
	transformMatrix        sprec.Mat4
	textureTransformMatrix sprec.Mat4
	clipBounds             sprec.Vec4

	shape         *Shape
	shapeMaterial *Material

	contour         *Contour
	contourMaterial *Material

	mesh      *Mesh
	subMeshes []SubMesh
	whiteMask *opengl.TwoDTexture

	target Target
}

func (r *Renderer) Init() {
	r.shapeMaterial.Allocate()
	r.mesh.Allocate()
	r.whiteMask.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             1,
		Height:            1,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.RGBA8,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.UNSIGNED_BYTE,
		Data:              []byte{0xFF, 0xFF, 0xFF, 0xFF},
	})
}

func (r *Renderer) Free() {
	defer r.shapeMaterial.Release()
	defer r.whiteMask.Release()
	defer r.mesh.Release()
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

	vertexOffset := r.mesh.Offset()
	for _, point := range shape.points {
		r.mesh.Append(Vertex{
			position: point.coords,
		})
	}
	vertexCount := r.mesh.Offset() - vertexOffset

	// TODO: HANDLE SubShapes!!!!

	if shape.fill.mode != StencilModeNone {
		// 	// clear stencil
		// 	c.subMeshes = append(c.subMeshes, SubMesh{
		// 		clipBounds: sprec.NewVec4(
		// 			float32(c.currentLayer.ClipBounds.X),
		// 			float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
		// 			float32(c.currentLayer.ClipBounds.Y),
		// 			float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
		// 		),
		// 		material:     c.opaqueMaterial,
		// 		texture:      c.whiteMask,
		// 		vertexOffset: offset,
		// 		vertexCount:  count,
		// 		culling:      false,
		// 		cullFace:     uint32(cullFace),
		// 		primitive:    gl.TRIANGLE_FAN,
		// 		skipColor:    true,
		// 		stencil:      true,
		// 		stencilCfg: stencilConfig{
		// 			stencilFuncFront: stencilFunc{
		// 				fn:   gl.ALWAYS,
		// 				ref:  0,
		// 				mask: 0xFF,
		// 			},
		// 			stencilFuncBack: stencilFunc{
		// 				fn:   gl.ALWAYS,
		// 				ref:  0,
		// 				mask: 0xFF,
		// 			},
		// 			stencilOpFront: stencilOp{
		// 				sfail:  gl.REPLACE,
		// 				dpfail: gl.REPLACE,
		// 				dppass: gl.REPLACE,
		// 			},
		// 			stencilOpBack: stencilOp{
		// 				sfail:  gl.REPLACE,
		// 				dpfail: gl.REPLACE,
		// 				dppass: gl.REPLACE,
		// 			},
		// 		},
		// 	})

		// 	// render stencil mask
		// 	c.subMeshes = append(c.subMeshes, SubMesh{
		// 		clipBounds: sprec.NewVec4(
		// 			float32(c.currentLayer.ClipBounds.X),
		// 			float32(c.currentLayer.ClipBounds.X+c.currentLayer.ClipBounds.Width),
		// 			float32(c.currentLayer.ClipBounds.Y),
		// 			float32(c.currentLayer.ClipBounds.Y+c.currentLayer.ClipBounds.Height),
		// 		),
		// 		material:     c.opaqueMaterial,
		// 		texture:      c.whiteMask,
		// 		vertexOffset: offset,
		// 		vertexCount:  count,
		// 		cullFace:     uint32(cullFace),
		// 		primitive:    gl.TRIANGLE_FAN,
		// 		skipColor:    true, // we don't want to render anything
		// 		stencil:      true,
		// 		stencilCfg: stencilConfig{
		// 			stencilFuncFront: stencilFunc{
		// 				fn:   gl.ALWAYS,
		// 				ref:  0,
		// 				mask: 0xFF,
		// 			},
		// 			stencilFuncBack: stencilFunc{
		// 				fn:   gl.ALWAYS,
		// 				ref:  0,
		// 				mask: 0xFF,
		// 			},
		// 			stencilOpFront: stencilOp{
		// 				sfail:  gl.KEEP,
		// 				dpfail: gl.KEEP,
		// 				dppass: gl.INCR_WRAP, // increase correct winding
		// 			},
		// 			stencilOpBack: stencilOp{
		// 				sfail:  gl.KEEP,
		// 				dpfail: gl.KEEP,
		// 				dppass: gl.DECR_WRAP, // decrease incorrect winding
		// 			},
		// 		},
		// 	})
	}

	texture := r.whiteMask
	if shape.fill.image != nil {
		texture = shape.fill.image.texture
	}

	r.subMeshes = append(r.subMeshes, SubMesh{
		clipBounds:             r.clipBounds,
		material:               r.shapeMaterial,
		transformMatrix:        r.transformMatrix,
		textureTransformMatrix: r.textureTransformMatrix,
		texture:                texture,
		color:                  shape.fill.color,
		vertexOffset:           vertexOffset, // FIXME: Take from subshape
		vertexCount:            vertexCount,  // FIXME: Take from subshape
		primitive:              gl.TRIANGLE_FAN,
		stencil:                shape.fill.mode != StencilModeNone,
		// 		stencilCfg: stencilConfig{
		// 			stencilFuncFront: stencilFunc{
		// 				fn:   gl.LESS,
		// 				ref:  0,
		// 				mask: 0xFF,
		// 			},
		// 			stencilFuncBack: stencilFunc{
		// 				fn:   gl.LESS,
		// 				ref:  0,
		// 				mask: 0xFF,
		// 			},
		// 			stencilOpFront: stencilOp{
		// 				sfail:  gl.KEEP,
		// 				dpfail: gl.KEEP,
		// 				dppass: gl.KEEP,
		// 			},
		// 			stencilOpBack: stencilOp{
		// 				sfail:  gl.KEEP,
		// 				dpfail: gl.KEEP,
		// 				dppass: gl.KEEP,
		// 			},
		// 		},
	})
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
	// TODO: Submit vertices and sub-meshes
}

func (r *Renderer) Begin(target Target) {
	r.target = target
	r.transformMatrix = sprec.IdentityMat4()
	r.textureTransformMatrix = sprec.IdentityMat4()
	r.clipBounds = sprec.NewVec4(
		0.0, target.Size.X,
		0.0, target.Size.Y,
	)
	r.mesh.Reset()
	r.subMeshes = r.subMeshes[:0]
}

func (r *Renderer) End() {
	r.mesh.Update()

	projectionMatrix := sprec.OrthoMat4(
		0.0, r.target.Size.X,
		0.0, r.target.Size.Y,
		0.0, 1.0,
	).ColumnMajorArray()

	gl.Enable(gl.CLIP_DISTANCE0)
	gl.Enable(gl.CLIP_DISTANCE1)
	gl.Enable(gl.CLIP_DISTANCE2)
	gl.Enable(gl.CLIP_DISTANCE3)

	r.target.Framebuffer.Use()
	gl.Viewport(0, 0, int32(r.target.Size.X), int32(r.target.Size.Y))
	gl.Enable(gl.FRAMEBUFFER_SRGB)
	gl.ClearStencil(0)
	gl.Clear(gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// TODO: Maybe optimize by accumulating draw commands
	// if they are similar.
	for _, subMesh := range r.subMeshes {
		material := subMesh.material
		transformMatrix := subMesh.transformMatrix.ColumnMajorArray()
		textureTransformMatrix := subMesh.textureTransformMatrix.ColumnMajorArray()

		if subMesh.skipColor {
			gl.ColorMask(false, false, false, false)
		} else {
			gl.ColorMask(true, true, true, true)
		}
		if subMesh.stencil {
			gl.Enable(gl.STENCIL_TEST)

			cfg := subMesh.stencilCfg
			gl.StencilFuncSeparate(gl.FRONT, cfg.stencilFuncFront.fn, cfg.stencilFuncFront.ref, cfg.stencilFuncFront.mask)
			gl.StencilFuncSeparate(gl.BACK, cfg.stencilFuncBack.fn, cfg.stencilFuncBack.ref, cfg.stencilFuncBack.mask)
			gl.StencilOpSeparate(gl.FRONT, cfg.stencilOpFront.sfail, cfg.stencilOpFront.dpfail, cfg.stencilOpFront.dppass)
			gl.StencilOpSeparate(gl.BACK, cfg.stencilOpBack.sfail, cfg.stencilOpBack.dpfail, cfg.stencilOpBack.dppass)
		} else {
			gl.Disable(gl.STENCIL_TEST)
		}
		if subMesh.culling {
			gl.Enable(gl.CULL_FACE)
			gl.CullFace(subMesh.cullFace)
		} else {
			gl.Disable(gl.CULL_FACE)
		}
		gl.UseProgram(material.program.ID())
		gl.UniformMatrix4fv(material.transformMatrixLocation, 1, false, &transformMatrix[0])
		gl.UniformMatrix4fv(material.textureTransformMatrixLocation, 1, false, &textureTransformMatrix[0])
		gl.UniformMatrix4fv(material.projectionMatrixLocation, 1, false, &projectionMatrix[0])
		gl.Uniform4f(material.clipDistancesLocation, subMesh.clipBounds.X, subMesh.clipBounds.Y, subMesh.clipBounds.Z, subMesh.clipBounds.W)
		gl.Uniform4f(material.colorLocation, subMesh.color.X, subMesh.color.Y, subMesh.color.Z, subMesh.color.W)
		gl.BindTextureUnit(0, subMesh.texture.ID())
		gl.Uniform1i(material.textureLocation, 0)
		gl.BindVertexArray(r.mesh.vertexArray.ID())
		if subMesh.patchVertices > 0 {
			gl.PatchParameteri(gl.PATCH_VERTICES, int32(subMesh.patchVertices))
		}
		gl.DrawArrays(subMesh.primitive, int32(subMesh.vertexOffset), int32(subMesh.vertexCount))
	}

	gl.ColorMask(true, true, true, true)
	gl.Disable(gl.STENCIL_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	// TODO: Remove once the remaining part of the framework
	// can handle resetting its settings.
	gl.Disable(gl.BLEND)

	gl.Disable(gl.CLIP_DISTANCE0)
	gl.Disable(gl.CLIP_DISTANCE1)
	gl.Disable(gl.CLIP_DISTANCE2)
	gl.Disable(gl.CLIP_DISTANCE3)
}

type Fill struct {
	color sprec.Vec4
	image *Image
	mode  StencilMode
}

type StencilMode int

const (
	StencilModeNone StencilMode = iota
	StencilModeNonZero
	StencilModeOdd
)

type Stroke struct {
	size  float32
	color sprec.Vec4
}

func MixStrokes(a, b Stroke, alpha float32) Stroke {
	return Stroke{
		size: (1-alpha)*a.size + alpha*b.size,
		color: sprec.Vec4Sum(
			sprec.Vec4Prod(a.color, (1-alpha)),
			sprec.Vec4Prod(b.color, alpha),
		),
	}
}
