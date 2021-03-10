package command

import "github.com/mokiat/lacking/opengl"

type RenderItem struct {
	BackfaceCulling  bool
	Program          *opengl.Program
	Uniforms         UniformRange
	VertexArray      *opengl.VertexArray
	Primitive        uint32
	IndexCount       int32
	IndexOffsetBytes int
}
