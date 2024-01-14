package internal

import "github.com/mokiat/lacking/render"

const (
	UniformBufferBindingCamera          = 0
	UniformBufferBindingModel           = 1
	UniformBufferBindingMaterial        = 2
	UniformBufferBindingLight           = 3
	UniformBufferBindingLightProperties = 4

	UniformBufferBindingPostprocess = 0
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

func NewShadowProgram(api render.API, sourceCode render.ProgramCode) render.Program {
	return BuildProgram(api, sourceCode,
		[]render.TextureBinding{},
		[]render.UniformBinding{
			render.NewUniformBinding("Light", UniformBufferBindingLight),
			render.NewUniformBinding("Model", UniformBufferBindingModel),
		},
	)
}

func NewGeometryProgram(api render.API, sourceCode render.ProgramCode) render.Program {
	return BuildProgram(api, sourceCode,
		[]render.TextureBinding{
			render.NewTextureBinding("albedoTwoDTextureIn", TextureBindingGeometryAlbedoTexture),
		},
		[]render.UniformBinding{
			render.NewUniformBinding("Camera", UniformBufferBindingCamera),
			render.NewUniformBinding("Model", UniformBufferBindingModel),
			render.NewUniformBinding("Material", UniformBufferBindingMaterial),
		},
	)
}

type Presentation struct {
	Program render.Program
}

func (p *Presentation) Delete() {
	p.Program.Release()
}

type PostprocessingPresentation struct {
	Presentation
}

func NewPostprocessingPresentation(api render.API, sourceCode render.ProgramCode) *PostprocessingPresentation {
	program := BuildProgram(api, sourceCode,
		[]render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", TextureBindingPostprocessFramebufferColor0),
		},
		[]render.UniformBinding{
			render.NewUniformBinding("Postprocess", UniformBufferBindingPostprocess),
		},
	)
	return &PostprocessingPresentation{
		Presentation: Presentation{
			Program: program,
		},
	}
}

type SkyboxPresentation struct {
	Presentation

	AlbedoColorLocation render.UniformLocation
}

func NewSkyboxPresentation(api render.API, sourceCode render.ProgramCode) *SkyboxPresentation {
	program := BuildProgram(api, sourceCode,
		[]render.TextureBinding{
			render.NewTextureBinding("albedoCubeTextureIn", TextureBindingSkyboxAlbedoTexture),
		},
		[]render.UniformBinding{
			render.NewUniformBinding("Camera", UniformBufferBindingCamera),
		},
	)
	return &SkyboxPresentation{
		Presentation: Presentation{
			Program: program,
		},
		AlbedoColorLocation: program.UniformLocation("albedoColorIn"),
	}
}

type LightingPresentation struct {
	Presentation
}

func NewLightingPresentation(api render.API, sourceCode render.ProgramCode) *LightingPresentation {
	program := BuildProgram(api, sourceCode,
		[]render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", TextureBindingLightingFramebufferColor0),
			render.NewTextureBinding("fbColor1TextureIn", TextureBindingLightingFramebufferColor1),
			render.NewTextureBinding("fbDepthTextureIn", TextureBindingLightingFramebufferDepth),
			render.NewTextureBinding("fbShadowTextureIn", TextureBindingShadowFramebufferDepth),
			render.NewTextureBinding("reflectionTextureIn", TextureBindingLightingReflectionTexture),
			render.NewTextureBinding("refractionTextureIn", TextureBindingLightingRefractionTexture),
		},
		[]render.UniformBinding{
			render.NewUniformBinding("Light", UniformBufferBindingLight),
			render.NewUniformBinding("Camera", UniformBufferBindingCamera),
			render.NewUniformBinding("LightProperties", UniformBufferBindingLightProperties),
		},
	)
	return &LightingPresentation{
		Presentation: Presentation{
			Program: program,
		},
	}
}

func BuildProgram(api render.API, sourceCode render.ProgramCode, textureBindings []render.TextureBinding, uniformBindings []render.UniformBinding) render.Program {
	return api.CreateProgram(render.ProgramInfo{
		SourceCode:      sourceCode,
		TextureBindings: textureBindings,
		UniformBindings: uniformBindings,
	})
}
