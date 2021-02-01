package pack

import "hash"

type ShaderProvider interface {
	Shader(ctx *Context) (*Shader, error)
	Digest(hasher hash.Hash) error
}

type Shader struct {
	Source string
}
