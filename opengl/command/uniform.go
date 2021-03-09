package command

type UniformKind int

const (
	UniformKindTexture UniformKind = 1 + iota
	UniformKindMatrix4f
	UniformKind4f
	UniformKind1f
)

type Uniform struct {
	Name   string
	Kind   UniformKind
	Offset int
}

type UniformRange struct {
	Offset int
	Count  int
}
