package render

type ShaderInfo struct {
	SourceCode string
}

type Shader interface {
	Release()
}
