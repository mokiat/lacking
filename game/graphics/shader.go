package graphics

import "github.com/mokiat/lacking/game/graphics/shading"

type ShaderSet struct {
	VertexShader   string
	FragmentShader string
}

type MeshConfig struct {
	HasArmature bool
}

type ShaderCollection struct {
	BuildGeometry func(meshConfig MeshConfig, fn shading.GeometryFunc) ShaderSet
	BuildForward  func(meshConfig MeshConfig, fn shading.ForwardFunc) ShaderSet

	ShadowMappingSet    func(cfg ShadowMappingShaderConfig) ShaderSet
	PBRGeometrySet      func(cfg PBRGeometryShaderConfig) ShaderSet
	DirectionalLightSet func() ShaderSet
	AmbientLightSet     func() ShaderSet
	PointLightSet       func() ShaderSet
	SpotLightSet        func() ShaderSet
	SkyboxSet           func() ShaderSet
	SkycolorSet         func() ShaderSet
	DebugSet            func() ShaderSet
	ExposureSet         func() ShaderSet
	PostprocessingSet   func(cfg PostprocessingShaderConfig) ShaderSet
}

type ShadowMappingShaderConfig struct {
	HasArmature bool
}

type PBRGeometryShaderConfig struct {
	HasArmature      bool
	HasAlphaTesting  bool
	HasVertexColors  bool
	HasAlbedoTexture bool
}

const (
	ReinhardToneMapping    ToneMapping = "reinhard"
	ExponentialToneMapping ToneMapping = "exponential"
)

type ToneMapping string

type PostprocessingShaderConfig struct {
	ToneMapping ToneMapping
}
