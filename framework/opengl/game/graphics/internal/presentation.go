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

type PostprocessingPresentation struct {
	Presentation

	FramebufferDraw0Location int32

	ExposureLocation int32
}

func NewPostprocessingPresentation(vertexSrc, fragmentSrc string) *PostprocessingPresentation {
	program := buildProgram(vertexSrc, fragmentSrc)
	return &PostprocessingPresentation{
		Presentation: Presentation{
			Program: program,
		},
		FramebufferDraw0Location: program.UniformLocation("fbColor0TextureIn"),
		ExposureLocation:         program.UniformLocation("exposureIn"),
	}
}

type SkyboxPresentation struct {
	Presentation

	ProjectionMatrixLocation int32
	ViewMatrixLocation       int32

	AlbedoCubeTextureLocation int32
}

func NewSkyboxPresentation(vertexSrc, fragmentSrc string) *SkyboxPresentation {
	program := buildProgram(vertexSrc, fragmentSrc)
	return &SkyboxPresentation{
		Presentation: Presentation{
			Program: program,
		},
		ProjectionMatrixLocation:  program.UniformLocation("projectionMatrixIn"),
		ViewMatrixLocation:        program.UniformLocation("viewMatrixIn"),
		AlbedoCubeTextureLocation: program.UniformLocation("albedoCubeTextureIn"),
	}
}

type ShadowPresentation struct {
	Presentation
}

func NewShadowPresentation(vertexSrc, fragmentSrc string) *ShadowPresentation {
	program := buildProgram(vertexSrc, fragmentSrc)
	return &ShadowPresentation{
		Presentation: Presentation{
			Program: program,
		},
	}
}

type GeometryPresentation struct {
	Presentation

	ProjectionMatrixLocation int32
	ModelMatrixLocation      int32
	ViewMatrixLocation       int32

	MetalnessLocation int32
	RoughnessLocation int32

	AlbedoColorLocation   int32
	AlbedoTextureLocation int32
}

func NewGeometryPresentation(vertexSrc, fragmentSrc string) *GeometryPresentation {
	program := buildProgram(vertexSrc, fragmentSrc)
	return &GeometryPresentation{
		Presentation: Presentation{
			Program: program,
		},
		ProjectionMatrixLocation: program.UniformLocation("projectionMatrixIn"),
		ModelMatrixLocation:      program.UniformLocation("modelMatrixIn"),
		ViewMatrixLocation:       program.UniformLocation("viewMatrixIn"),
		MetalnessLocation:        program.UniformLocation("metalnessIn"),
		RoughnessLocation:        program.UniformLocation("roughnessIn"),
		AlbedoColorLocation:      program.UniformLocation("albedoColorIn"),
		AlbedoTextureLocation:    program.UniformLocation("albedoTwoDTextureIn"),
	}
}

type LightingPresentation struct {
	Presentation

	FramebufferDraw0Location int32
	FramebufferDraw1Location int32
	FramebufferDepthLocation int32

	ProjectionMatrixLocation int32
	CameraMatrixLocation     int32
	ViewMatrixLocation       int32

	ReflectionTextureLocation int32
	RefractionTextureLocation int32

	LightDirection int32
	LightIntensity int32
}

func NewLightingPresentation(vertexSrc, fragmentSrc string) *LightingPresentation {
	program := buildProgram(vertexSrc, fragmentSrc)
	return &LightingPresentation{
		Presentation: Presentation{
			Program: program,
		},

		FramebufferDraw0Location: program.UniformLocation("fbColor0TextureIn"),
		FramebufferDraw1Location: program.UniformLocation("fbColor1TextureIn"),
		FramebufferDepthLocation: program.UniformLocation("fbDepthTextureIn"),

		ProjectionMatrixLocation: program.UniformLocation("projectionMatrixIn"),
		CameraMatrixLocation:     program.UniformLocation("cameraMatrixIn"),
		ViewMatrixLocation:       program.UniformLocation("viewMatrixIn"),

		ReflectionTextureLocation: program.UniformLocation("reflectionTextureIn"),
		RefractionTextureLocation: program.UniformLocation("refractionTextureIn"),

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
