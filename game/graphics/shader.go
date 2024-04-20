package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics/lsl"
	"github.com/mokiat/lacking/render"
)

// ShaderConstraints contains the constraints imposed on the shader construction
// process.
type ShaderConstraints struct {

	// LoadGeometryPreset specifies whether the shader should load the geometry
	// preset.
	LoadGeometryPreset bool

	// LoadSkyPreset specifies whether the shader should load the sky preset.
	LoadSkyPreset bool

	// HasOutput0 specifies whether the shader has an output for the first
	// render target.
	HasOutput0 bool

	// HasOutput1 specifies whether the shader has an output for the second
	// render target.
	HasOutput1 bool

	// HasOutput2 specifies whether the shader has an output for the third
	// render target.
	HasOutput2 bool

	// HasOutput3 specifies whether the shader has an output for the fourth
	// render target.
	HasOutput3 bool

	// HasCoords specifies whether the mesh has coordinates.
	HasCoords bool

	// HasNormals specifies whether the mesh has normals.
	HasNormals bool

	// HasTangents specifies whether the mesh has tangents.
	HasTangents bool

	// HasTexCoords specifies whether the mesh has texture coordinates.
	HasTexCoords bool

	// HasVertexColors specifies whether the mesh has vertex colors.
	HasVertexColors bool

	// HasArmature specifies whether the mesh has an armature.
	HasArmature bool
}

// GeometryConstraints contains the constraints imposed on the geometry shader
// construction process.
type GeometryConstraints struct {

	// HasArmature specifies whether the mesh has an armature.
	HasArmature bool

	// HasNormals specifies whether the mesh has normals.
	HasNormals bool

	// HasTexCoords specifies whether the mesh has texture coordinates.
	HasTexCoords bool

	// HasVertexColors specifies whether the mesh has vertex colors.
	HasVertexColors bool
}

// ShadowConstraints contains the constraints imposed on the shadow shader
// construction process.
type ShadowConstraints struct {

	// HasArmature specifies whether the mesh has an armature.
	HasArmature bool
}

// ForwardConstraints contains the constraints imposed on the forward shader
// construction process.
type ForwardConstraints struct {

	// HasArmature specifies whether the mesh has an armature.
	HasArmature bool
}

// SkyConstraints contains the constraints imposed on the sky shader
// construction process.
type SkyConstraints struct {
}

// ShaderBuilder abstracts the process of building a shader program. The
// implementation of this interface will depend on the rendering backend.
type ShaderBuilder interface {

	// BuildCode creates the program code for a custom shader.
	BuildCode(constraints ShaderConstraints, shader *lsl.Shader) render.ProgramCode

	// BuildGeometryCode creates the program code for a geometry pass.
	BuildGeometryCode(constraints GeometryConstraints, shader *lsl.Shader) render.ProgramCode

	// BuildShadowCode creates the program code for a shadow pass.
	BuildShadowCode(constraints ShadowConstraints, shader *lsl.Shader) render.ProgramCode

	// BuildForwardCode creates the program code for a shadow pass.
	BuildForwardCode(constraints ForwardConstraints, shader *lsl.Shader) render.ProgramCode

	// BuildSkyCode creates the program code for a sky pass.
	BuildSkyCode(constraints SkyConstraints, shader *lsl.Shader) render.ProgramCode
}

// ShaderInfo contains the information needed to create a custom Shader.
type ShaderInfo struct {

	// SourceCode is the source code of the shader.
	SourceCode string
}

// GeometryShader represents a shader that is used during the geometry pass.
type GeometryShader struct {
	builder ShaderBuilder
	ast     *lsl.Shader
}

func (s *GeometryShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.builder.BuildGeometryCode(GeometryConstraints{
		HasArmature:     info.MeshHasArmature,
		HasNormals:      info.MeshHasNormals,
		HasTexCoords:    info.MeshHasTextureUVs,
		HasVertexColors: info.MeshHasVertexColors,
	}, s.ast)
}

// ShadowShader represents a shader that is used during the shadow pass for
// a particular light source.
type ShadowShader struct {
	builder ShaderBuilder
	ast     *lsl.Shader
}

func (s *ShadowShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.builder.BuildShadowCode(ShadowConstraints{
		HasArmature: info.MeshHasArmature,
	}, s.ast)
}

// ForwardShader represents a shader that is used during the forward pass.
type ForwardShader struct {
	builder ShaderBuilder
	ast     *lsl.Shader
}

func (s *ForwardShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.builder.BuildForwardCode(ForwardConstraints{
		HasArmature: info.MeshHasArmature,
	}, s.ast)
}

// SkyShader represents a shader that is used during the sky pass.
type SkyShader struct {
	builder ShaderBuilder
	ast     *lsl.Shader
}

func (s *SkyShader) createProgramCode() render.ProgramCode {
	return s.builder.BuildCode(ShaderConstraints{
		LoadSkyPreset: true,
		HasCoords:     true,
		HasOutput0:    true,
	}, s.ast)
}

///////////////// OLD CODE FOLLOWS ////////////////////////////

type ShaderCollection struct {
	AmbientLightSet     func() render.ProgramCode
	PointLightSet       func() render.ProgramCode
	SpotLightSet        func() render.ProgramCode
	DirectionalLightSet func() render.ProgramCode

	SkyboxSet   func() render.ProgramCode
	SkycolorSet func() render.ProgramCode

	DebugSet func() render.ProgramCode

	ExposureSet func() render.ProgramCode

	BloomDownsampleSet func() render.ProgramCode
	BloomBlurSet       func() render.ProgramCode

	PostprocessingSet func(cfg PostprocessingShaderConfig) render.ProgramCode
}

type PostprocessingShaderConfig struct {
	ToneMapping ToneMapping
	Bloom       bool
}

type ToneMapping string

const (
	ReinhardToneMapping    ToneMapping = "reinhard"
	ExponentialToneMapping ToneMapping = "exponential"
)
