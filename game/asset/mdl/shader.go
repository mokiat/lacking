package mdl

import "github.com/mokiat/lacking/game/asset/dto"

type ShaderType = dto.ShaderType

const (
	ShaderTypeGeometry    ShaderType = dto.ShaderTypeGeometry
	ShaderTypeShadow      ShaderType = dto.ShaderTypeShadow
	ShaderTypeForward     ShaderType = dto.ShaderTypeForward
	ShaderTypeSky         ShaderType = dto.ShaderTypeSky
	ShaderTypePostprocess ShaderType = dto.ShaderTypePostprocess
)

func NewShader(shaderType ShaderType) *Shader {
	return &Shader{
		Object:     NewObject(),
		shaderType: shaderType,
	}
}

type Shader struct {
	*Object
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
