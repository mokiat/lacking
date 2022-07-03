package render

type ProgramInfo struct {
	VertexShader    Shader
	FragmentShader  Shader
	TextureBindings []TextureBinding
}

func NewTextureBinding(name string, index int) TextureBinding {
	return TextureBinding{
		Name:  name,
		Index: index,
	}
}

type TextureBinding struct {
	Name  string
	Index int
}

type UniformLocation interface{}

type ProgramObject interface {
	_isProgramObject() bool // ensures interface uniqueness
}

type Program interface {
	ProgramObject
	UniformLocation(name string) UniformLocation
	Release()
}
