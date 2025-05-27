package mdl

import "github.com/mokiat/lacking/game/asset/dto/shadingdto"

type ShaderType = shadingdto.ShaderType

const (
	ShaderTypeGeometry    ShaderType = shadingdto.ShaderTypeGeometry
	ShaderTypeShadow      ShaderType = shadingdto.ShaderTypeShadow
	ShaderTypeForward     ShaderType = shadingdto.ShaderTypeForward
	ShaderTypeSky         ShaderType = shadingdto.ShaderTypeSky
	ShaderTypePostprocess ShaderType = shadingdto.ShaderTypePostprocess
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
