package internal

import "github.com/mokiat/lacking/framework/opengl"

type SubMesh struct {
	material     *Material
	texture      *opengl.TwoDTexture
	vertexOffset int
	vertexCount  int
	primitive    uint32
}