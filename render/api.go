package render

type API interface {
	Capabilities() Capabilities

	DefaultFramebuffer() Framebuffer

	CreateFramebuffer(info FramebufferInfo) Framebuffer
	CreateColorTexture2D(info ColorTexture2DInfo) Texture
	CreateColorTextureCube(info ColorTextureCubeInfo) Texture
	CreateDepthTexture2D(info DepthTexture2DInfo) Texture
	CreateStencilTexture2D(info StencilTexture2DInfo) Texture
	CreateDepthStencilTexture2D(info DepthStencilTexture2DInfo) Texture
	CreateVertexShader(info ShaderInfo) Shader
	CreateFragmentShader(info ShaderInfo) Shader
	CreateProgram(info ProgramInfo) Program
	CreateVertexBuffer(info BufferInfo) Buffer
	CreateIndexBuffer(info BufferInfo) Buffer
	CreateVertexArray(info VertexArrayInfo) VertexArray
	CreatePipeline(info PipelineInfo) Pipeline
	CreateCommandQueue() CommandQueue

	BeginRenderPass(info RenderPassInfo)
	EndRenderPass()

	BindPipeline(pipeline Pipeline)
	Uniform4f(location UniformLocation, values [4]float32)
	Uniform1i(location UniformLocation, value int)
	UniformMatrix4f(location UniformLocation, values [16]float32)
	TextureUnit(index int, texture Texture)
	Draw(vertexOffset, vertexCount, instanceCount int)
	DrawIndexed(indexOffset, indexCount, instanceCount int)
	SubmitQueue(queue CommandQueue)
}
