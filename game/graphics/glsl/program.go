package glsl

import (
	"strings"

	"github.com/mokiat/lacking/render"
)

// ProgramCode is a GLSL-specific implementation of render.ProgramCode.
type ProgramCode struct {
	render.ProgramCodeMarker

	// VertexCode specifies the vertex shader code.
	VertexCode string

	// FragmentCode specifies the fragment shader code.
	FragmentCode string
}

func (p ProgramCode) String() string {
	// TODO: Check why this is required.
	vertex := strings.ReplaceAll(p.VertexCode, "\n\n", "\n")
	fragment := strings.ReplaceAll(p.FragmentCode, "\n\n", "\n")
	return vertex + "\n\n---\n\n" + fragment
}
