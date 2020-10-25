package pack

type ShaderProvider interface {
	Shader() *Shader
}

type Shader struct {
	Source string
}
