package ui

import (
	"github.com/mokiat/lacking/render"
)

func newMaterial(sourceCode render.ProgramCode) *material {
	return &material{
		sourceCode: sourceCode,
	}
}

type material struct {
	sourceCode render.ProgramCode

	program                        render.Program
	projectionMatrixLocation       render.UniformLocation
	transformMatrixLocation        render.UniformLocation
	clipMatrixLocation             render.UniformLocation
	textureTransformMatrixLocation render.UniformLocation
	textureLocation                render.UniformLocation
	colorLocation                  render.UniformLocation
}

func (m *material) Allocate(api render.API) {
	m.program = api.CreateProgram(render.ProgramInfo{
		SourceCode: m.sourceCode,
	})

	m.projectionMatrixLocation = m.program.UniformLocation("projectionMatrixIn")
	m.transformMatrixLocation = m.program.UniformLocation("transformMatrixIn")
	m.clipMatrixLocation = m.program.UniformLocation("clipMatrixIn")
	m.textureTransformMatrixLocation = m.program.UniformLocation("textureTransformMatrixIn")
	m.textureLocation = m.program.UniformLocation("textureIn")
	m.colorLocation = m.program.UniformLocation("colorIn")
}

func (m *material) Release() {
	m.program.Release()
}
