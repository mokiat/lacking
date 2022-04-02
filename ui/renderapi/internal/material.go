package internal

import (
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/ui/renderapi/plugin"
)

func newMaterial(shaders plugin.ShaderSet) *Material {
	return &Material{
		vertexSrc:   shaders.VertexShader,
		fragmentSrc: shaders.FragmentShader,
	}
}

type Material struct {
	vertexSrc   func() string
	fragmentSrc func() string

	program                        render.Program
	transformMatrixLocation        render.UniformLocation
	textureTransformMatrixLocation render.UniformLocation
	projectionMatrixLocation       render.UniformLocation
	clipDistancesLocation          render.UniformLocation
	textureLocation                render.UniformLocation
	colorLocation                  render.UniformLocation
}

func (m *Material) Allocate(api render.API) {
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

	m.transformMatrixLocation = m.program.UniformLocation("transformMatrixIn")
	m.textureTransformMatrixLocation = m.program.UniformLocation("textureTransformMatrixIn")
	m.projectionMatrixLocation = m.program.UniformLocation("projectionMatrixIn")
	m.clipDistancesLocation = m.program.UniformLocation("clipDistancesIn")
	m.textureLocation = m.program.UniformLocation("textureIn")
	m.colorLocation = m.program.UniformLocation("colorIn")
}

func (m *Material) Release() {
	m.program.Release()
}
