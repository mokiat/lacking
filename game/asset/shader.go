package asset

// Shader represents a shader program that can be used to render a mesh.
type Shader struct {

	// ShaderType specifies the type of the shader.
	ShaderType ShaderType

	// SourceCode is the source code of the shader.
	SourceCode string
}

// ShaderType specifies the type of a shader.
type ShaderType uint8

const (
	// ShaderTypeGeometry is a shader that is used during a geometry pass.
	ShaderTypeGeometry ShaderType = iota

	// ShaderTypeShadow is a shader that is used during a shadow pass.
	ShaderTypeShadow

	// ShaderTypeForward is a shader that is used during a forward pass.
	ShaderTypeForward

	// ShaderTypeSky is a shader that is used to render the sky.
	ShaderTypeSky

	// ShaderTypePostprocess is a shader that is used during a post-processing
	// pass.
	ShaderTypePostprocess
)
