package render

type ProgramInfo struct {
	VertexShader   Shader
	FragmentShader Shader
}

type UniformLocation interface{}

type Program interface {
	UniformLocation(name string) UniformLocation
	Release()
}
