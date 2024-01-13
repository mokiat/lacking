package ui

import (
	"github.com/mokiat/lacking/render"
)

func newMaterial(programInfo render.ProgramInfo) *material {
	return &material{
		programInfo: programInfo,
	}
}

type material struct {
	programInfo render.ProgramInfo
	program     render.Program
}

func (m *material) Allocate(api render.API) {
	m.program = api.CreateProgram(m.programInfo)
}

func (m *material) Release() {
	m.program.Release()
}
