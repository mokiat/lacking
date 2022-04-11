package plugin

import "github.com/mokiat/lacking/game/graphics"

type ShaderCollection struct {
	ExposureSet         func() ShaderSet
	PostprocessingSet   func(mapping ToneMapping) ShaderSet
	DirectionalLightSet func() ShaderSet
	AmbientLightSet     func() ShaderSet
	SkyboxSet           func() ShaderSet
	SkycolorSet         func() ShaderSet
	PBRShaderSet        func(definition graphics.PBRMaterialDefinition) ShaderSet
}

type ShaderSet struct {
	VertexShader   func() string
	FragmentShader func() string
}

type ToneMapping string

const (
	ReinhardToneMapping    ToneMapping = "reinhard"
	ExponentialToneMapping ToneMapping = "exponential"
)
