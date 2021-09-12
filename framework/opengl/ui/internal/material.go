package internal

import (
	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/lacking/framework/opengl"
)

func newMaterial(vertexSrc, fragmentSrc func() string) *Material {
	return &Material{
		vertexSrc:   vertexSrc,
		fragmentSrc: fragmentSrc,

		program: opengl.NewProgram(),
	}
}

type Material struct {
	vertexSrc   func() string
	fragmentSrc func() string

	program                        *opengl.Program
	transformMatrixLocation        int32
	textureTransformMatrixLocation int32
	projectionMatrixLocation       int32
	clipDistancesLocation          int32
	textureLocation                int32
	colorLocation                  int32
}

func (m *Material) Allocate() {
	vertexShader := opengl.NewShader()
	vertexShader.Allocate(opengl.ShaderAllocateInfo{
		ShaderType: gl.VERTEX_SHADER,
		SourceCode: m.vertexSrc(),
	})
	defer func() {
		vertexShader.Release()
	}()

	fragmentShader := opengl.NewShader()
	fragmentShader.Allocate(opengl.ShaderAllocateInfo{
		ShaderType: gl.FRAGMENT_SHADER,
		SourceCode: m.fragmentSrc(),
	})
	defer func() {
		fragmentShader.Release()
	}()

	m.program.Allocate(opengl.ProgramAllocateInfo{
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
