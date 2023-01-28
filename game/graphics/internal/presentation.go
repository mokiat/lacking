package internal

import "github.com/mokiat/lacking/render"

const (
	UniformBufferBindingCamera   = 0
	UniformBufferBindingModel    = 1
	UniformBufferBindingMaterial = 2
	UniformBufferBindingLight    = 3
)

const (
	TextureBindingGeometryAlbedoTexture = 0

	TextureBindingLightingFramebufferColor0 = 0
	TextureBindingLightingFramebufferColor1 = 1
	TextureBindingLightingFramebufferColor2 = 2
	TextureBindingLightingFramebufferDepth  = 3
	TextureBindingShadowFramebufferDepth    = 4
	TextureBindingLightingReflectionTexture = 4
	TextureBindingLightingRefractionTexture = 5

	TextureBindingPostprocessFramebufferColor0 = 0

	TextureBindingSkyboxAlbedoTexture = 0
)

func NewShadowProgram(api render.API, vertexSrc, fragmentSrc string) render.Program {
	return BuildProgram(api, vertexSrc, fragmentSrc, nil, []render.UniformBinding{
		render.NewUniformBinding("Light", UniformBufferBindingLight),
		render.NewUniformBinding("Model", UniformBufferBindingModel),
	})
}

func NewGeometryProgram(api render.API, vertexSrc, fragmentSrc string) render.Program {
	return BuildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
		render.NewTextureBinding("albedoTwoDTextureIn", TextureBindingGeometryAlbedoTexture),
	}, []render.UniformBinding{
		render.NewUniformBinding("Camera", UniformBufferBindingCamera),
		render.NewUniformBinding("Model", UniformBufferBindingModel),
		render.NewUniformBinding("Material", UniformBufferBindingMaterial),
	})
}

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
	program := BuildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
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
	program := BuildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
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

type LightingPresentation struct {
	Presentation

	// TODO: Move to lighting uniform buffer
	LightDirection render.UniformLocation
	LightIntensity render.UniformLocation
	LightRange     render.UniformLocation
	LightOuterCos  render.UniformLocation
	LightInnerCos  render.UniformLocation
}

func NewLightingPresentation(api render.API, vertexSrc, fragmentSrc string) *LightingPresentation {
	program := BuildProgram(api, vertexSrc, fragmentSrc, []render.TextureBinding{
		render.NewTextureBinding("fbColor0TextureIn", TextureBindingLightingFramebufferColor0),
		render.NewTextureBinding("fbColor1TextureIn", TextureBindingLightingFramebufferColor1),
		render.NewTextureBinding("fbDepthTextureIn", TextureBindingLightingFramebufferDepth),
		render.NewTextureBinding("fbShadowTextureIn", TextureBindingShadowFramebufferDepth),
		render.NewTextureBinding("reflectionTextureIn", TextureBindingLightingReflectionTexture),
		render.NewTextureBinding("refractionTextureIn", TextureBindingLightingRefractionTexture),
	}, []render.UniformBinding{
		render.NewUniformBinding("Light", UniformBufferBindingLight),
		render.NewUniformBinding("Camera", UniformBufferBindingCamera),
	})
	return &LightingPresentation{
		Presentation: Presentation{
			Program: program,
		},

		LightDirection: program.UniformLocation("lightDirectionIn"),
		LightIntensity: program.UniformLocation("lightIntensityIn"),
		LightRange:     program.UniformLocation("lightRangeIn"),
		LightOuterCos:  program.UniformLocation("lightOuterCosIn"),
		LightInnerCos:  program.UniformLocation("lightInnerCosIn"),
	}
}

func BuildProgram(api render.API, vertSrc, fragSrc string, textureBindings []render.TextureBinding, uniformBindings []render.UniformBinding) render.Program {
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
