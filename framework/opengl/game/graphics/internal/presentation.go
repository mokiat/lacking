package internal

import (
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/mokiat/lacking/framework/opengl"
)

type Presentation struct {
	Program *opengl.Program
}

func (p *Presentation) Delete() {
	p.Program.Release()
}

type LightingPresentation struct {
	Presentation

	FramebufferDraw0 int32
	FramebufferDraw1 int32
	FramebufferDepth int32

	ProjectionMatrixLocation int32
	ModelviewMatrixLocation  int32
	CameraMatrixLocation     int32
	ViewMatrixLocation       int32

	LightDirection int32
	LightIntensity int32
}

func NewLightingPresentation(vertexSrc, fragmentSrc string) *LightingPresentation {
	program := buildProgram(vertexSrc, fragmentSrc)
	return &LightingPresentation{
		Presentation: Presentation{
			Program: program,
		},

		FramebufferDraw0: program.UniformLocation("fbColor0TextureIn"),
		FramebufferDraw1: program.UniformLocation("fbColor1TextureIn"),
		FramebufferDepth: program.UniformLocation("fbDepthTextureIn"),

		ProjectionMatrixLocation: program.UniformLocation("projectionMatrixIn"),
		ModelviewMatrixLocation:  program.UniformLocation("modelviewMatrixIn"),
		CameraMatrixLocation:     program.UniformLocation("cameraMatrixIn"),
		ViewMatrixLocation:       program.UniformLocation("viewMatrixIn"),

		LightDirection: program.UniformLocation("lightDirectionIn"),
		LightIntensity: program.UniformLocation("lightIntensityIn"),
	}
}

func buildProgram(vertSrc, fragSrc string) *opengl.Program {
	vertexShader := opengl.NewShader()
	vertexShader.Allocate(opengl.ShaderAllocateInfo{
		ShaderType: gl.VERTEX_SHADER,
		SourceCode: vertSrc,
	})
	defer func() {
		vertexShader.Release()
	}()

	fragmentShader := opengl.NewShader()
	fragmentShader.Allocate(opengl.ShaderAllocateInfo{
		ShaderType: gl.FRAGMENT_SHADER,
		SourceCode: fragSrc,
	})
	defer func() {
		fragmentShader.Release()
	}()

	program := opengl.NewProgram()
	program.Allocate(opengl.ProgramAllocateInfo{
		VertexShader:   vertexShader,
		FragmentShader: fragmentShader,
	})
	return program
}
