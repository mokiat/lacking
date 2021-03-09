package command

import "github.com/mokiat/lacking/opengl"

type RenderItem struct {
	Program  *opengl.Program
	Uniforms UniformRange
}
