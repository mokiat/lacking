package graphics

import (
	"github.com/mokiat/lacking/game/graphics/shading"
	"github.com/mokiat/lacking/render"
)

type MeshConfig struct {
	HasArmature bool
}

type ShaderCollection struct {
	BuildGeometry func(meshConfig MeshConfig, fn shading.GeometryFunc) render.ProgramCode
	BuildForward  func(meshConfig MeshConfig, fn shading.ForwardFunc) render.ProgramCode

	ShadowMappingSet    func(cfg ShadowMappingShaderConfig) render.ProgramCode
	PBRGeometrySet      func(cfg PBRGeometryShaderConfig) render.ProgramCode
	DirectionalLightSet func() render.ProgramCode
	AmbientLightSet     func() render.ProgramCode
	PointLightSet       func() render.ProgramCode
	SpotLightSet        func() render.ProgramCode
	SkyboxSet           func() render.ProgramCode
	SkycolorSet         func() render.ProgramCode
	DebugSet            func() render.ProgramCode
	ExposureSet         func() render.ProgramCode
	PostprocessingSet   func(cfg PostprocessingShaderConfig) render.ProgramCode
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
