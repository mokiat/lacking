package render

type CommandQueue interface {
	BindPipeline(pipeline Pipeline)
	TextureUnit(index int, texture Texture)
	UniformBufferUnit(index int, buffer Buffer)
	UniformBufferUnitRange(index int, buffer Buffer, offset, size int)
	Draw(vertexOffset, vertexCount, instanceCount int)
	DrawIndexed(indexOffset, indexCount, instanceCount int)
	CopyContentToBuffer(info CopyContentToBufferInfo)
	// Deprecated: Upload directly through the Queue.
	UpdateBufferData(buffer Buffer, info BufferUpdateInfo)
	Release()
}
