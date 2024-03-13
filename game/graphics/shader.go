package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics/shading"
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

// ShaderBuilder abstracts the process of building a shader program. The
// implementation of this interface will depend on the rendering backend.
type ShaderBuilder interface {

	// BuildGeometryCode creates the program code for a geometry pass.
	BuildGeometryCode(constraints GeometryConstraints, vertex shading.GenericBuilderFunc, fragment shading.GenericBuilderFunc) render.ProgramCode

	// BuildShadowCode creates the program code for a shadow pass.
	BuildShadowCode(constraints ShadowConstraints, vertex shading.GenericBuilderFunc, fragment shading.GenericBuilderFunc) render.ProgramCode

	// BuildForwardCode creates the program code for a shadow pass.
	BuildForwardCode(constraints ForwardConstraints, vertex shading.GenericBuilderFunc, fragment shading.GenericBuilderFunc) render.ProgramCode
}

// GeometryShaderInfo contains the information needed to create a
// custom GeometryShader.
type GeometryShaderInfo struct {

	// VertexBuilder is the function that will be used to build the
	// program code for the vertex shader.
	VertexBuilder shading.GeometryVertexBuilderFunc

	// FragmentBuilder is the function that will be used to build the
	// program code for the fragment shader.
	FragmentBuilder shading.GeometryFragmentBuilderFunc
}

// GeometryShader represents a shader that is used during the geometry pass.
type GeometryShader interface {
	internal.Shader
	_isGeometryShader()
}

// ShadowShaderInfo contains the information needed to create a
// custom ShadowShader.
type ShadowShaderInfo struct {

	// VertexBuilder is the function that will be used to build the
	// program code for the vertex shader.
	VertexBuilder shading.ShadowVertexBuilderFunc

	// FragmentBuilder is the function that will be used to build the
	// program code for the fragment shader.
	FragmentBuilder shading.ShadowFragmentBuilderFunc
}

// ShadowShader represents a shader that is used during the shadow pass for
// a particular light source.
type ShadowShader interface {
	internal.Shader
	_isShadowShader()
}

// ForwardShaderInfo contains the information needed to create a
// custom ForwardShader.
type ForwardShaderInfo struct {

	// VertexBuilder is the function that will be used to build the
	// program code for the vertex shader.
	VertexBuilder shading.ForwardVertexBuilderFunc

	// FragmentBuilder is the function that will be used to build the
	// program code for the fragment shader.
	FragmentBuilder shading.ForwardFragmentBuilderFunc
}

// ForwardShader represents a shader that is used during the forward pass.
type ForwardShader interface {
	internal.Shader
	_isForwardShader()
}

type customGeometryShader struct {
	builder  ShaderBuilder
	vertex   shading.GenericBuilderFunc
	fragment shading.GenericBuilderFunc
}

func (s *customGeometryShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.builder.BuildGeometryCode(GeometryConstraints{
		HasArmature: info.MeshHasArmature,
	}, s.vertex, s.fragment)
}

func (*customGeometryShader) _isGeometryShader() {}

type customShadowShader struct {
	builder  ShaderBuilder
	vertex   shading.GenericBuilderFunc
	fragment shading.GenericBuilderFunc
}

func (s *customShadowShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.builder.BuildShadowCode(ShadowConstraints{
		HasArmature: info.MeshHasArmature,
	}, s.vertex, s.fragment)
}

func (*customShadowShader) _isShadowShader() {}

type customForwardShader struct {
	builder  ShaderBuilder
	vertex   shading.GenericBuilderFunc
	fragment shading.GenericBuilderFunc
}

func (s *customForwardShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.builder.BuildForwardCode(ForwardConstraints{
		HasArmature: info.MeshHasArmature,
	}, s.vertex, s.fragment)
}

func (*customForwardShader) _isForwardShader() {}

type defaultGeometryShader struct {
	shaders ShaderCollection

	useAlphaTesting  bool
	useAlbedoTexture bool
}

func (s *defaultGeometryShader) CreateProgramCode(info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return s.shaders.PBRGeometrySet(PBRGeometryShaderConfig{
		HasArmature:      info.MeshHasArmature,
		HasAlphaTesting:  s.useAlphaTesting,
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

type PostprocessingShaderConfig struct {
	ToneMapping ToneMapping
}

type ToneMapping string

const (
	ReinhardToneMapping    ToneMapping = "reinhard"
	ExponentialToneMapping ToneMapping = "exponential"
)
