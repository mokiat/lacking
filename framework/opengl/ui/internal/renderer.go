package internal

import (
	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/ui"
)

const maxVertexCount = 2048

func NewRenderer() *Renderer {
	return &Renderer{
		shape:          newShape(),
		contour:        newContour(),
		mesh:           NewMesh(maxVertexCount),
		whiteMask:      opengl.NewTwoDTexture(),
		opaqueMaterial: NewDrawMaterial(),
	}
}

type Renderer struct {
	transformMatrix        sprec.Mat3
	textureTransformMatrix sprec.Mat3
	clipBounds             sprec.Vec4
	shape                  *Shape
	contour                *Contour

	mesh           *Mesh
	subMeshes      []SubMesh
	whiteMask      *opengl.TwoDTexture
	opaqueMaterial *Material
}

func (r *Renderer) Create() {
	r.mesh.Allocate()
	r.opaqueMaterial.Allocate()
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

func (r *Renderer) Destroy() {
	defer r.whiteMask.Release()
	defer r.opaqueMaterial.Release()
	defer r.mesh.Release()
}

func (r *Renderer) Transform() sprec.Mat3 {
	return r.transformMatrix
}

func (r *Renderer) SetTransform(transform sprec.Mat3) {
	r.transformMatrix = transform
}

func (r *Renderer) TextureTransform() sprec.Mat3 {
	return r.textureTransformMatrix
}

func (r *Renderer) SetTextureTransform(textureTransform sprec.Mat3) {
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
			color:    ui.Red(), // FIXME: Remove, instead through uniform
		})
	}
	vertexCount := r.mesh.Offset() - vertexOffset

	// translation := sprec.NewVec2(
	// 	float32(c.currentLayer.Translation.X),
	// 	float32(c.currentLayer.Translation.Y),
	// )

	// cullFace := gl.BACK
	// if c.activeShape.Winding == ui.WindingCW {
	// 	cullFace = gl.FRONT
	// }

	// if c.activeShape.Rule == ui.FillRuleSimple {
	r.subMeshes = append(r.subMeshes, SubMesh{
		clipBounds:   r.clipBounds,
		material:     r.opaqueMaterial,
		texture:      r.whiteMask,
		vertexOffset: vertexOffset,
		vertexCount:  vertexCount,
		cullFace:     uint32(gl.BACK), // uint32(cullFace),
		primitive:    gl.TRIANGLE_FAN,
	})
	// }

	// if c.activeShape.Rule != ui.FillRuleSimple {
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

	// 	// render final polygon
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
	// 		skipColor:    false, // we want to render now
	// 		stencil:      true,
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
	// 	})
	// }
	// TODO: Submit vertices and sub-meshes
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

type Fill struct {
	// Rule            FillRule
	// Winding         Winding
	color sprec.Vec4
	image *Image
}

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
