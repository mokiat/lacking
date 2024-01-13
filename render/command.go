package render

type CommandQueue interface {
	BindPipeline(pipeline Pipeline)
	Uniform1f(location UniformLocation, value float32)
	Uniform3f(location UniformLocation, values [3]float32)
	Uniform4f(location UniformLocation, values [4]float32)
	UniformBufferUnit(index int, buffer Buffer)
	UniformBufferUnitRange(index int, buffer Buffer, offset, size int)
	TextureUnit(index int, texture Texture)
	Draw(vertexOffset, vertexCount, instanceCount int)
	DrawIndexed(indexOffset, indexCount, instanceCount int)
	CopyContentToBuffer(info CopyContentToBufferInfo)
	UpdateBufferData(buffer Buffer, info BufferUpdateInfo)
	Release()
}
