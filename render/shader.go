package render

type ShaderInfo struct {
	SourceCode string
}

type ShaderObject interface {
	_isShaderObject() bool // ensures interface uniqueness
}

type Shader interface {
	ShaderObject
	Release()
}
