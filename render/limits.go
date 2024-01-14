package render

// Limits describes the limits of the implementation.
type Limits interface {

	// UniformBufferOffsetAlignment returns the alignment requirement
	// for uniform buffer offsets.
	UniformBufferOffsetAlignment() int
}

// DetermineUniformBlockSize returns the size of the uniform block
// in case multiple ones need to be aligned inside a buffer.
func DetermineUniformBlockSize(api API, blockSize int) int {
	uniformBufferOffsetAlignemnt := api.Limits().UniformBufferOffsetAlignment()
	if excess := blockSize % uniformBufferOffsetAlignemnt; excess > 0 {
		blockSize += uniformBufferOffsetAlignemnt - excess
	}
	return blockSize
}
