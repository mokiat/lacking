package ui

import "github.com/mokiat/lacking/render"

// ShaderCollection holds the set of shaders to be used for rendering.
type ShaderCollection struct {
	ShapeShadedSet func() render.ProgramCode
	ShapeBlankSet  func() render.ProgramCode
	ContourSet     func() render.ProgramCode
	TextSet        func() render.ProgramCode
}
