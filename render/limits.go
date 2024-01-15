package render

// Limits describes the limits of the implementation.
type Limits interface {

	// UniformBufferOffsetAlignment returns the alignment requirement
	// for uniform buffer offsets.
	UniformBufferOffsetAlignment() int
}
