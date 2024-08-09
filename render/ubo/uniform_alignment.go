package ubo

import "github.com/mokiat/lacking/render"

// DetermineUniformBlockSize returns the size of the uniform block
// in case multiple ones need to be aligned inside a buffer.
func DetermineUniformBlockSize(api render.API, blockSize int) int {
	uniformBufferOffsetAlignemnt := api.Limits().UniformBufferOffsetAlignment()
	if unaligned := blockSize % uniformBufferOffsetAlignemnt; unaligned > 0 {
		blockSize += uniformBufferOffsetAlignemnt - unaligned
	}
	return blockSize
}
