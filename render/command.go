package render

// BufferMarker marks a type as being a CommandBuffer.
type CommandBufferMarker interface {
	isCommandBufferType()
}

// CommandBuffer is used to record commands that should be executed
// on the GPU.
type CommandBuffer interface {
	CommandBufferMarker

	// CopyFramebufferToBuffer copies the contents of the current framebuffer
	// to the specified buffer.
	CopyFramebufferToBuffer(info CopyFramebufferToBufferInfo)

	// CopyFramebufferToTexture copies the contents of the current framebuffer
	// to the specified texture.
	CopyFramebufferToTexture(info CopyFramebufferToTextureInfo)

	// BeginRenderPass starts a new render pass that will render to the
	// specified framebuffer.
	BeginRenderPass(info RenderPassInfo)

	// BindPipeline configures the pipeline that should be used for the
	// following draw commands.
	BindPipeline(pipeline Pipeline)

	// TextureUnit configures which texture should be used for the
	// specified texture unit.
	TextureUnit(index int, texture Texture)

	// UniformBufferUnit configures which buffer should be used for the
	// specified buffer unit.
	UniformBufferUnit(index int, buffer Buffer, offset, size int)

	// Draw uses the vertex buffer to draw primitives.
	Draw(vertexOffset, vertexCount, instanceCount int)

	// DrawIndexed uses the index buffer to draw primitives.
	DrawIndexed(indexOffset, indexCount, instanceCount int)

	// EndRenderPass ends the current render pass.
	EndRenderPass()
}
