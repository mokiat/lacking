package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
)

type SubMesh struct {
	material               *Material
	transformMatrix        sprec.Mat4
	textureTransformMatrix sprec.Mat4
	texture                *opengl.TwoDTexture
	color                  sprec.Vec4
	vertexOffset           int
	vertexCount            int
	primitive              uint32
	culling                bool
	cullFace               uint32
	clipBounds             sprec.Vec4
	skipColor              bool
	stencil                bool
	stencilCfg             stencilConfig
}

type stencilConfig struct {
	stencilOpFront   stencilOp
	stencilOpBack    stencilOp
	stencilFuncFront stencilFunc
	stencilFuncBack  stencilFunc
}

type stencilOp struct {
	sfail  uint32
	dpfail uint32
	dppass uint32
}

type stencilFunc struct {
	fn   uint32
	ref  int32
	mask uint32
}
