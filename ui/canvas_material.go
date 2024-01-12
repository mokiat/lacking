package ui

import (
	"github.com/mokiat/lacking/render"
)

const (
	uniformBufferBindingCamera = 0
)

const (
	textureBindingColorTexture = 0
	textureBindingFontTexture  = 1
)

func newMaterial(programInfo render.ProgramInfo) *material {
	return &material{
		programInfo: programInfo,
	}
}

type material struct {
	programInfo render.ProgramInfo
	program     render.Program

	transformMatrixLocation        render.UniformLocation
	clipMatrixLocation             render.UniformLocation
	textureTransformMatrixLocation render.UniformLocation
	colorLocation                  render.UniformLocation
}

func (m *material) Allocate(api render.API) {
	m.program = api.CreateProgram(m.programInfo)

	m.transformMatrixLocation = m.program.UniformLocation("transformMatrixIn")
	m.clipMatrixLocation = m.program.UniformLocation("clipMatrixIn")
	m.textureTransformMatrixLocation = m.program.UniformLocation("textureTransformMatrixIn")
	m.colorLocation = m.program.UniformLocation("colorIn")
}

func (m *material) Release() {
	m.program.Release()
}
