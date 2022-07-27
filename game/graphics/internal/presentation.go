package internal

import "github.com/mokiat/lacking/render"

const (
	UniformBufferBindingCamera   = 0
	UniformBufferBindingModel    = 1
	UniformBufferBindingMaterial = 2
)

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
	}, nil)
	return &PostprocessingPresentation{
		Presentation: Presentation{
			Program: program,
		},
		ExposureLocation: program.UniformLocation("exposureIn"),
	}
}

type SkyboxPresentation struct {
	Presentation

	AlbedoColorLocation render.UniformLocation
}

func NewSkyboxPresentation(api render.API, vertexSrc, fragmentSrc string) *SkyboxPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
		render.NewTextureBinding("albedoCubeTextureIn", TextureBindingSkyboxAlbedoTexture),
	}, []render.UniformBinding{
		render.NewUniformBinding("Camera", UniformBufferBindingCamera),
	})
	return &SkyboxPresentation{
		Presentation: Presentation{
			Program: program,
		},
		AlbedoColorLocation: program.UniformLocation("albedoColorIn"),
	}
}

type ShadowPresentation struct {
	Presentation
}

func NewShadowPresentation(api render.API, vertexSrc, fragmentSrc string) *ShadowPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc, nil, nil)
	return &ShadowPresentation{
		Presentation: Presentation{
			Program: program,
		},
	}
}

type GeometryPresentation struct {
	Presentation
}

func NewGeometryPresentation(api render.API, vertexSrc, fragmentSrc string) *GeometryPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
		render.NewTextureBinding("albedoTwoDTextureIn", TextureBindingGeometryAlbedoTexture),
	}, []render.UniformBinding{
		render.NewUniformBinding("Camera", UniformBufferBindingCamera),
		render.NewUniformBinding("Model", UniformBufferBindingModel),
		render.NewUniformBinding("Material", UniformBufferBindingMaterial),
	})
	return &GeometryPresentation{
		Presentation: Presentation{
			Program: program,
		},
	}
}

type LightingPresentation struct {
	Presentation

	LightDirection render.UniformLocation
	LightIntensity render.UniformLocation
}

func NewLightingPresentation(api render.API, vertexSrc, fragmentSrc string) *LightingPresentation {
	program := buildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
		render.NewTextureBinding("fbColor0TextureIn", TextureBindingLightingFramebufferColor0),
		render.NewTextureBinding("fbColor1TextureIn", TextureBindingLightingFramebufferColor1),
		render.NewTextureBinding("fbDepthTextureIn", TextureBindingLightingFramebufferDepth),
		render.NewTextureBinding("reflectionTextureIn", TextureBindingLightingReflectionTexture),
		render.NewTextureBinding("refractionTextureIn", TextureBindingLightingRefractionTexture),
	}, []render.UniformBinding{
		render.NewUniformBinding("Camera", UniformBufferBindingCamera),
	})
	return &LightingPresentation{
		Presentation: Presentation{
			Program: program,
		},

		LightDirection: program.UniformLocation("lightDirectionIn"),
		LightIntensity: program.UniformLocation("lightIntensityIn"),
	}
}

func buildProgram(api render.API, vertSrc, fragSrc string, textureBindings []render.TextureBinding, uniformBindings []render.UniformBinding) render.Program {
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
		UniformBindings: uniformBindings,
	})
}
