package render

// API provides access to low-level graphics manipulation and rendering.
type API interface {

	// Limits returns information on the supported limits of the implementation.
	Limits() Limits

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

	// CreateFramebuffer creates a new Framebuffer object based on the
	// provided FramebufferInfo.
	CreateFramebuffer(info FramebufferInfo) Framebuffer

	// CreateProgram creates a new Program object based on the provided
	// ProgramInfo.
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

	// CreateCommandBuffer creates a new CommandBuffer object with the
	// specified initial capacity. The buffer will automatically grow
	// as needed but it is recommended to provide a reasonable initial
	// capacity to avoid unnecessary allocations.
	CreateCommandBuffer(initialCapacity int) CommandBuffer

	// Queue can be used to schedule commands to be executed on the GPU.
	Queue() Queue
}
