package render

// API provides access to low-level graphics manipulation and rendering.
type API interface {

	// Limits returns information on the supported limits of the implementation.
	Limits() Limits

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

	// CreateColorTexture2D creates a new 2D Texture that can be
	// used to store color values.
	CreateColorTexture2D(info ColorTexture2DInfo) Texture

	// CreateColorTextureCube creates a new Cube Texture that can be
	// used to store color values.
	CreateColorTextureCube(info ColorTextureCubeInfo) Texture

	// CreateDepthTexture2D creates a new 2D Texture that can be
	// used to store depth values.
	CreateDepthTexture2D(info DepthTexture2DInfo) Texture

	// CreateStencilTexture2D creates a new 2D Texture that can be
	// used to store stencil values.
	CreateStencilTexture2D(info StencilTexture2DInfo) Texture

	// CreateDepthStencilTexture2D creates a new 2D Texture that can be
	// used to store depth and stencil values together.
	CreateDepthStencilTexture2D(info DepthStencilTexture2DInfo) Texture

	// CreateSampler creates a new Sampler object based on the provided
	// SamplerInfo.
	CreateSampler(info SamplerInfo) Sampler

	// CreateVertexBuffer creates a new Buffer object that can be used
	// to store vertex data.
	CreateVertexBuffer(info BufferInfo) Buffer

	// CreateIndexBuffer creates a new Buffer object that can be used
	// to store index data.
	CreateIndexBuffer(info BufferInfo) Buffer

	// CreatePixelTransferBuffer creates a new Buffer object that can be used
	// to store pixel data from a transfer operation.
	CreatePixelTransferBuffer(info BufferInfo) Buffer

	// CreateUniformBuffer creates a new Buffer object that can be used
	// to store uniform data.
	CreateUniformBuffer(info BufferInfo) Buffer

	// CreateVertexArray creates a new VertexArray object that describes
	// the layout of vertex and index data for a mesh.
	CreateVertexArray(info VertexArrayInfo) VertexArray

	// CreatePipeline creates a new Pipeline object that describes the
	// desired rendering state.
	CreatePipeline(info PipelineInfo) Pipeline

	// CreateCommandBuffer creates a new CommandBuffer object with the
	// specified initial capacity. The buffer will automatically grow
	// as needed but it is recommended to provide a reasonable initial
	// capacity to avoid unnecessary allocations.
	CreateCommandBuffer(initialCapacity int) CommandBuffer

	// Queue can be used to schedule commands to be executed on the GPU.
	Queue() Queue
}
