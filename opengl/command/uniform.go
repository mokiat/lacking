package command

type Uniform struct {
	Kind      UniformKind
	FloatData []float32
}

type UniformKind int

const (
	UniformKind1f UniformKind = 1 + iota
	UniformKindMatrix4f
)

type BindUniform struct {
	Name    string
	Uniform Uniform
}
