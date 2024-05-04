package mdl

import "github.com/mokiat/lacking/game/asset"

type ShaderType = asset.ShaderType

const (
	ShaderTypeGeometry    ShaderType = asset.ShaderTypeGeometry
	ShaderTypeShadow      ShaderType = asset.ShaderTypeShadow
	ShaderTypeForward     ShaderType = asset.ShaderTypeForward
	ShaderTypeSky         ShaderType = asset.ShaderTypeSky
	ShaderTypePostprocess ShaderType = asset.ShaderTypePostprocess
)

func NewShader(shaderType ShaderType) *Shader {
	return &Shader{
		shaderType: shaderType,
	}
}

type Shader struct {
	shaderType ShaderType
	sourceCode string
}

func (s *Shader) SourceCode() string {
	return s.sourceCode
}

func (s *Shader) SetSourceCode(sourceCode string) {
	s.sourceCode = sourceCode
}

func (s *Shader) ShaderType() ShaderType {
	return s.shaderType
}

func (s *Shader) SetShaderType(shaderType ShaderType) {
	s.shaderType = shaderType
}
