package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
)

type SubMesh struct {
	material      *Material
	texture       *opengl.TwoDTexture
	vertexOffset  int
	vertexCount   int
	patchVertices int
	primitive     uint32
	clipBounds    sprec.Vec4
}
