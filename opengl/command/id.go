package command

type ID struct {
	Type  Type
	Index int
}

type Type int

const (
	TypeClear Type = 1 + iota
	TypeDepthConfig
	TypeChangeFramebuffer
	TypeChangeProgram
	TypeBindUniform
	TypeBindTexture
)
