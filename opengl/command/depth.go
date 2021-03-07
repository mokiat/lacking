package command

type DepthConfig struct {
	DepthTest  OptionalBool
	DepthWrite OptionalBool
	DepthFunc  OptionalUint32
}
