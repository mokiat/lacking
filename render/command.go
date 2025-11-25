package render

// BufferMarker marks a type as being a CommandBuffer.
type CommandBufferMarker interface {
	isCommandBufferType()
}

// CommandBuffer is used to record commands that should be executed
// on the GPU.
type CommandBuffer interface {
	CommandBufferMarker

	// Label returns a human-readable name for the command buffer.
	Label() string

	// CopyFramebufferToBuffer copies the contents of the current framebuffer
	// to the specified buffer.
	CopyFramebufferToBuffer(info CopyFramebufferToBufferInfo)

	// CopyFramebufferToTexture copies the contents of the current framebuffer
	// to the specified texture.
	CopyFramebufferToTexture(info CopyFramebufferToTextureInfo)

	// BeginRenderPass starts a new render pass that will render to the
	// specified framebuffer.
	BeginRenderPass(info RenderPassInfo)

	// SetViewport changes the viewport settings of the render pass.
	SetViewport(x, y, width, height uint32)

	// BindPipeline configures the pipeline that should be used for the
	// following draw commands.
	BindPipeline(pipeline Pipeline)

	// TextureUnit configures which texture should be used for the
	// specified texture unit.
	TextureUnit(index uint, texture Texture)

	// SamplerUnit configures which sampler should be used for the
	// specified texture unit.
	SamplerUnit(index uint, sampler Sampler)

	// UniformBufferUnit configures which buffer should be used for the
	// specified buffer unit.
	UniformBufferUnit(index uint, buffer Buffer, offset, size uint32)

	// Draw uses the vertex buffer to draw primitives.
	Draw(vertexOffset, vertexCount, instanceCount uint32)

	// DrawIndexed uses the index buffer to draw primitives.
	DrawIndexed(indexByteOffset, indexCount, instanceCount uint32)

	// EndRenderPass ends the current render pass.
	EndRenderPass()
}

// CommandBufferInfo describes the information needed to create a
// CommandBuffer.
type CommandBufferInfo struct {

	// InitialCapacity specifies the initial capacity of the command buffer
	// in bytes.
	InitialCapacity uint

	// Label is a human-readable name for the command buffer.
	Label string
}
