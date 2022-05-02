package ui

import (
	"github.com/mokiat/lacking/render"
)

func newMaterial(shaders ShaderSet) *material {
	return &material{
		vertexSrc:   shaders.VertexShader,
		fragmentSrc: shaders.FragmentShader,
	}
}

type material struct {
	vertexSrc   func() string
	fragmentSrc func() string

	program                        render.Program
	projectionMatrixLocation       render.UniformLocation
	transformMatrixLocation        render.UniformLocation
	clipMatrixLocation             render.UniformLocation
	textureTransformMatrixLocation render.UniformLocation
	textureLocation                render.UniformLocation
	colorLocation                  render.UniformLocation
}

func (m *material) Allocate(api render.API) {
	vertexShader := api.CreateVertexShader(render.ShaderInfo{
		SourceCode: m.vertexSrc(),
	})
	defer func() {
		vertexShader.Release()
	}()

	fragmentShader := api.CreateFragmentShader(render.ShaderInfo{
		SourceCode: m.fragmentSrc(),
	})
	defer func() {
		fragmentShader.Release()
	}()

	m.program = api.CreateProgram(render.ProgramInfo{
		VertexShader:   vertexShader,
		FragmentShader: fragmentShader,
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
