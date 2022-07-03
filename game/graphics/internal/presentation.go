package internal

import "github.com/mokiat/lacking/render"

const (
	TextureBindingGeometryAlbedoTexture = 0

	TextureBindingLightingFramebufferColor0 = 0
	TextureBindingLightingFramebufferColor1 = 1
	TextureBindingLightingFramebufferColor2 = 2
	TextureBindingLightingFramebufferDepth  = 3
	TextureBindingLightingReflectionTexture = 4
	TextureBindingLightingRefractionTexture = 5

	TextureBindingPostprocessFramebufferColor0 = 0

	TextureBindingSkyboxAlbedoTexture = 0
)

type Presentation struct {
	Program render.Program
}

func (p *Presentation) Delete() {
	p.Program.Release()
}

type PostprocessingPresentation struct {
	Presentation

	ExposureLocation render.UniformLocation
}

func NewPostprocessingPresentation(api render.API, vertexSrc, fragmentSrc string) *PostprocessingPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
		render.NewTextureBinding("fbColor0TextureIn", TextureBindingPostprocessFramebufferColor0),
	})
	return &PostprocessingPresentation{
		Presentation: Presentation{
			Program: program,
		},
		ExposureLocation: program.UniformLocation("exposureIn"),
	}
}

type SkyboxPresentation struct {
	Presentation

	ProjectionMatrixLocation render.UniformLocation
	ViewMatrixLocation       render.UniformLocation
	AlbedoColorLocation      render.UniformLocation
}

func NewSkyboxPresentation(api render.API, vertexSrc, fragmentSrc string) *SkyboxPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
		render.NewTextureBinding("albedoCubeTextureIn", TextureBindingSkyboxAlbedoTexture),
	})
	return &SkyboxPresentation{
		Presentation: Presentation{
			Program: program,
		},
		ProjectionMatrixLocation: program.UniformLocation("projectionMatrixIn"),
		ViewMatrixLocation:       program.UniformLocation("viewMatrixIn"),
		AlbedoColorLocation:      program.UniformLocation("albedoColorIn"),
	}
}

type ShadowPresentation struct {
	Presentation
}

func NewShadowPresentation(api render.API, vertexSrc, fragmentSrc string) *ShadowPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc, nil)
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
}

func NewGeometryPresentation(api render.API, vertexSrc, fragmentSrc string) *GeometryPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
		render.NewTextureBinding("albedoTwoDTextureIn", TextureBindingGeometryAlbedoTexture),
	})
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
	}
}

type LightingPresentation struct {
	Presentation

	ProjectionMatrixLocation render.UniformLocation
	CameraMatrixLocation     render.UniformLocation
	ViewMatrixLocation       render.UniformLocation
	LightDirection           render.UniformLocation
	LightIntensity           render.UniformLocation
}

func NewLightingPresentation(api render.API, vertexSrc, fragmentSrc string) *LightingPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
		render.NewTextureBinding("fbColor0TextureIn", TextureBindingLightingFramebufferColor0),
		render.NewTextureBinding("fbColor1TextureIn", TextureBindingLightingFramebufferColor1),
		render.NewTextureBinding("fbDepthTextureIn", TextureBindingLightingFramebufferDepth),
		render.NewTextureBinding("reflectionTextureIn", TextureBindingLightingReflectionTexture),
		render.NewTextureBinding("refractionTextureIn", TextureBindingLightingRefractionTexture),
	})
	return &LightingPresentation{
		Presentation: Presentation{
			Program: program,
		},

		ProjectionMatrixLocation: program.UniformLocation("projectionMatrixIn"),
		CameraMatrixLocation:     program.UniformLocation("cameraMatrixIn"),
		ViewMatrixLocation:       program.UniformLocation("viewMatrixIn"),

		LightDirection: program.UniformLocation("lightDirectionIn"),
		LightIntensity: program.UniformLocation("lightIntensityIn"),
	}
}

func buildProgram(api render.API, vertSrc, fragSrc string, textureBindings []render.TextureBinding) render.Program {
	vertexShader := api.CreateVertexShader(render.ShaderInfo{
		SourceCode: vertSrc,
	})
	defer vertexShader.Release()

	fragmentShader := api.CreateFragmentShader(render.ShaderInfo{
		SourceCode: fragSrc,
	})
	defer fragmentShader.Release()

	return api.CreateProgram(render.ProgramInfo{
		VertexShader:    vertexShader,
		FragmentShader:  fragmentShader,
		TextureBindings: textureBindings,
	})
}
