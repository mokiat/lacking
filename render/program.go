package render

type ProgramInfo struct {
	VertexShader   Shader
	FragmentShader Shader
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
