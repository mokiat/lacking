package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics/lsl"
	"github.com/mokiat/lacking/render"
)

// GeometryConstraints contains the constraints imposed on the geometry shader
// construction process.
type GeometryConstraints struct {

	// HasArmature specifies whether the mesh has an armature.
	HasArmature bool
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
type GeometryShader interface {
	internal.Shader
	_isGeometryShader()
}

// ShadowShader represents a shader that is used during the shadow pass for
// a particular light source.
type ShadowShader interface {
	internal.Shader
	_isShadowShader()
}

// ForwardShader represents a shader that is used during the forward pass.
type ForwardShader interface {
	internal.Shader
	_isForwardShader()
}

// SkyShader represents a shader that is used during the sky pass.
type SkyShader interface {
	internal.Shader
	_isSkyShader()
}

type customGeometryShader struct {
	builder ShaderBuilder
	ast     *lsl.Shader
}

func (s *customGeometryShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.builder.BuildGeometryCode(GeometryConstraints{
		HasArmature: info.MeshHasArmature,
	}, s.ast)
}

func (*customGeometryShader) _isGeometryShader() {}

type customShadowShader struct {
	builder ShaderBuilder
	ast     *lsl.Shader
}

func (s *customShadowShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.builder.BuildShadowCode(ShadowConstraints{
		HasArmature: info.MeshHasArmature,
	}, s.ast)
}

func (*customShadowShader) _isShadowShader() {}

type customForwardShader struct {
	builder ShaderBuilder
	ast     *lsl.Shader
}

func (s *customForwardShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.builder.BuildForwardCode(ForwardConstraints{
		HasArmature: info.MeshHasArmature,
	}, s.ast)
}

func (*customForwardShader) _isForwardShader() {}

type customSkyShader struct {
	builder ShaderBuilder
	ast     *lsl.Shader
}

func (s *customSkyShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.builder.BuildSkyCode(SkyConstraints{}, s.ast)
}

func (*customSkyShader) _isSkyShader() {}

type defaultGeometryShader struct {
	shaders ShaderCollection

	useAlphaTesting  bool
	useAlbedoTexture bool
}

func (s *defaultGeometryShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.shaders.PBRGeometrySet(PBRGeometryShaderConfig{
		HasArmature:      info.MeshHasArmature,
		HasAlphaTesting:  s.useAlphaTesting,
		HasNormals:       info.MeshHasNormals,
		HasVertexColors:  info.MeshHasVertexColors,
		HasAlbedoTexture: s.useAlbedoTexture,
	})
}

func (*defaultGeometryShader) _isGeometryShader() {}

type defaultShadowShader struct {
	shaders ShaderCollection
}

func (s *defaultShadowShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.shaders.ShadowMappingSet(ShadowMappingShaderConfig{
		HasArmature: info.MeshHasArmature,
	})
}

func (*defaultShadowShader) _isShadowShader() {}

///////////////// OLD CODE FOLLOWS ////////////////////////////

type ShaderCollection struct {
	ShadowMappingSet    func(cfg ShadowMappingShaderConfig) render.ProgramCode
	PBRGeometrySet      func(cfg PBRGeometryShaderConfig) render.ProgramCode
	DirectionalLightSet func() render.ProgramCode
	EmissiveLightSet    func() render.ProgramCode
	AmbientLightSet     func() render.ProgramCode
	PointLightSet       func() render.ProgramCode
	SpotLightSet        func() render.ProgramCode
	SkyboxSet           func() render.ProgramCode
	SkycolorSet         func() render.ProgramCode
	DebugSet            func() render.ProgramCode
	ExposureSet         func() render.ProgramCode
	BloomDownsampleSet  func() render.ProgramCode
	BloomBlurSet        func() render.ProgramCode
	PostprocessingSet   func(cfg PostprocessingShaderConfig) render.ProgramCode
}

type ShadowMappingShaderConfig struct {
	HasArmature bool
}

type PBRGeometryShaderConfig struct {
	HasArmature      bool
	HasAlphaTesting  bool
	HasNormals       bool
	HasVertexColors  bool
	HasAlbedoTexture bool
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
