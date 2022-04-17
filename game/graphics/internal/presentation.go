package internal

import "github.com/mokiat/lacking/render"

type Presentation struct {
	Program render.Program
}

func (p *Presentation) Delete() {
	p.Program.Release()
}

type PostprocessingPresentation struct {
	Presentation

	FramebufferDraw0Location render.UniformLocation
	ExposureLocation         render.UniformLocation
}

func NewPostprocessingPresentation(api render.API, vertexSrc, fragmentSrc string) *PostprocessingPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc)
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

	ProjectionMatrixLocation  render.UniformLocation
	ViewMatrixLocation        render.UniformLocation
	AlbedoCubeTextureLocation render.UniformLocation
	AlbedoColorLocation       render.UniformLocation
}

func NewSkyboxPresentation(api render.API, vertexSrc, fragmentSrc string) *SkyboxPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc)
	return &SkyboxPresentation{
		Presentation: Presentation{
			Program: program,
		},
		ProjectionMatrixLocation:  program.UniformLocation("projectionMatrixIn"),
		ViewMatrixLocation:        program.UniformLocation("viewMatrixIn"),
		AlbedoCubeTextureLocation: program.UniformLocation("albedoCubeTextureIn"),
		AlbedoColorLocation:       program.UniformLocation("albedoColorIn"),
	}
}

type ShadowPresentation struct {
	Presentation
}

func NewShadowPresentation(api render.API, vertexSrc, fragmentSrc string) *ShadowPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc)
	return &ShadowPresentation{
		Presentation: Presentation{
			Program: program,
		},
	}
}

type GeometryPresentation struct {
	Presentation

	ProjectionMatrixLocation render.UniformLocation
	ModelMatrixLocation      render.UniformLocation
	ViewMatrixLocation       render.UniformLocation
	MetalnessLocation        render.UniformLocation
	RoughnessLocation        render.UniformLocation
	AlbedoColorLocation      render.UniformLocation
	AlbedoTextureLocation    render.UniformLocation
}

func NewGeometryPresentation(api render.API, vertexSrc, fragmentSrc string) *GeometryPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc)
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

	FramebufferDraw0Location  render.UniformLocation
	FramebufferDraw1Location  render.UniformLocation
	FramebufferDepthLocation  render.UniformLocation
	ProjectionMatrixLocation  render.UniformLocation
	CameraMatrixLocation      render.UniformLocation
	ViewMatrixLocation        render.UniformLocation
	ReflectionTextureLocation render.UniformLocation
	RefractionTextureLocation render.UniformLocation
	LightDirection            render.UniformLocation
	LightIntensity            render.UniformLocation
}

func NewLightingPresentation(api render.API, vertexSrc, fragmentSrc string) *LightingPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc)
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

func buildProgram(api render.API, vertSrc, fragSrc string) render.Program {
	vertexShader := api.CreateVertexShader(render.ShaderInfo{
		SourceCode: vertSrc,
	})
	defer vertexShader.Release()

	fragmentShader := api.CreateFragmentShader(render.ShaderInfo{
		SourceCode: fragSrc,
	})
	defer fragmentShader.Release()

	return api.CreateProgram(render.ProgramInfo{
		VertexShader:   vertexShader,
		FragmentShader: fragmentShader,
	})
}
