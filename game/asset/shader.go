package asset

// Shader represents a shader program that can be used to render a mesh.
type Shader struct {

	// ShaderType specifies the type of the shader.
	ShaderType ShaderType

	// SourceCode is the source code of the shader.
	SourceCode string
}

type ShaderType uint8

const (
	ShaderTypeGeometry ShaderType = iota
	ShaderTypeShadow
	ShaderTypeForward
	ShaderTypeSky
	ShaderTypePostprocess
)
