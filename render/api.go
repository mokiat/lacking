package render

// API provides access to low-level graphics manipulation and rendering.
type API interface {

	// Capabilities returns the information on the supported features and
	// performance characteristics of the implementation.
	Capabilities() Capabilities

	// DefaultFramebuffer returns the default framebuffer that is provided
	// by the target window surface.
	DefaultFramebuffer() Framebuffer

	// DetermineContentFormat returns the format that should be used
	// with CopyContentToTexture and similar methods when dealing with
	// the specified Framebuffer.
	DetermineContentFormat(framebuffer Framebuffer) DataFormat

	CreateFramebuffer(info FramebufferInfo) Framebuffer

	CreateProgram(info ProgramInfo) Program
	CreateColorTexture2D(info ColorTexture2DInfo) Texture
	CreateColorTextureCube(info ColorTextureCubeInfo) Texture
	CreateDepthTexture2D(info DepthTexture2DInfo) Texture
	CreateStencilTexture2D(info StencilTexture2DInfo) Texture
	CreateDepthStencilTexture2D(info DepthStencilTexture2DInfo) Texture
	CreateVertexBuffer(info BufferInfo) Buffer
	CreateIndexBuffer(info BufferInfo) Buffer
	CreatePixelTransferBuffer(info BufferInfo) Buffer
	CreateUniformBuffer(info BufferInfo) Buffer
	CreateVertexArray(info VertexArrayInfo) VertexArray
	CreatePipeline(info PipelineInfo) Pipeline

	// Deprecated: use Queue instead.
	CreateCommandQueue() CommandQueue

	// TODO
	// Queue() CommandQueue // this should be an immediate queue

	BeginRenderPass(info RenderPassInfo)
	EndRenderPass()

	// Invalidate indicates that the graphics state might have changed
	// from outside this API and any cached state by the API should
	// be discarded.
	//
	// Using this command will force a subsequent draw command to initialize
	// all graphics state (e.g. blend state, depth state, stencil state, etc.)
	Invalidate()

	BindPipeline(pipeline Pipeline)
	Uniform1f(location UniformLocation, value float32)
	Uniform1i(location UniformLocation, value int)
	Uniform3f(location UniformLocation, values [3]float32)
	Uniform4f(location UniformLocation, values [4]float32)
	UniformMatrix4f(location UniformLocation, values [16]float32)
	UniformBufferUnit(index int, buffer Buffer)
	UniformBufferUnitRange(index int, buffer Buffer, offset, size int)
	TextureUnit(index int, texture Texture)
	Draw(vertexOffset, vertexCount, instanceCount int)
	DrawIndexed(indexOffset, indexCount, instanceCount int)
	CopyContentToTexture(info CopyContentToTextureInfo)
	SubmitQueue(queue CommandQueue)
	CreateFence() Fence
}
