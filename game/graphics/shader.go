package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics/lsl"
	"github.com/mokiat/lacking/render"
)

// ShaderMeshConstraints contains the constraints imposed by the mesh.
type ShaderMeshConstraints struct {

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

// ShaderOutputConstraints contains the constraints imposed by the designated
// render target.
type ShaderOutputConstraints struct {

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
}

// ShaderTypeConstraints contains the constraints imposed by the shader type.
type ShaderTypeConstraints struct {

	// Type specifies the type of shader.
	Type ShaderType
}

// ShaderConstraints contains the constraints imposed on the shader construction
// process.
type ShaderConstraints struct {
	ShaderMeshConstraints
	ShaderOutputConstraints
	ShaderTypeConstraints
}

// GeometryConstraints contains the constraints imposed on the geometry shader
// construction process.
type GeometryConstraints struct {

	// HasArmature specifies whether the mesh has an armature.
	HasArmature bool

	// HasNormals specifies whether the mesh has normals.
	HasNormals bool

	// HasTangents specifies whether the mesh has tangents.
	HasTangents bool

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

// ShaderBuilder abstracts the process of building a shader program. The
// implementation of this interface will depend on the rendering backend.
type ShaderBuilder interface {

	// BuildCode creates the program code for a custom shader.
	BuildCode(constraints ShaderConstraints, shader *lsl.Shader) render.ProgramCode

	// BuildGeometryCode creates the program code for a geometry pass.
	BuildGeometryCode(constraints GeometryConstraints, shader *lsl.Shader) render.ProgramCode
}

// ShaderInfo contains the information needed to create a custom Shader.
type ShaderInfo struct {

	// ShaderType specifies the type of the shader.
	ShaderType ShaderType

	// SourceCode is the source code of the shader.
	SourceCode string
}

const (
	ShaderTypeGeometry ShaderType = iota
	ShaderTypeShadow
	ShaderTypeForward
	ShaderTypeSky
	ShaderTypePostprocess
)

// ShaderType specifies the type of a shader.
type ShaderType uint8

// String returns the string representation of the shader type.
func (t ShaderType) String() string {
	switch t {
	case ShaderTypeGeometry:
		return "geometry"
	case ShaderTypeShadow:
		return "shadow"
	case ShaderTypeForward:
		return "forward"
	case ShaderTypeSky:
		return "sky"
	case ShaderTypePostprocess:
		return "postprocess"
	default:
		return "unknown"
	}
}

// Shader represents a custom shader program.
type Shader struct {
	ast *lsl.Shader
}

// Deprecated: Rework
func (e *Engine) createGeometryProgramCode(shader *lsl.Shader, info internal.ShaderProgramCodeInfo) render.ProgramCode {
	return e.shaderBuilder.BuildGeometryCode(GeometryConstraints{
		HasArmature:     info.MeshHasArmature,
		HasNormals:      info.MeshHasNormals,
		HasTangents:     info.MeshHasTangents,
		HasTexCoords:    info.MeshHasTextureUVs,
		HasVertexColors: info.MeshHasVertexColors,
	}, shader)
}

///////////////// OLD CODE FOLLOWS ////////////////////////////

type ShaderCollection struct {
	AmbientLightSet     func() render.ProgramCode
	PointLightSet       func() render.ProgramCode
	SpotLightSet        func() render.ProgramCode
	DirectionalLightSet func() render.ProgramCode

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
