package render

type CommandQueue interface {
	BindPipeline(pipeline Pipeline)
	Uniform4f(location UniformLocation, values [4]float32)
	Uniform1i(location UniformLocation, value int)
	UniformMatrix4f(location UniformLocation, values [16]float32)
	TextureUnit(index int, texture Texture)
	Draw(vertexOffset, vertexCount, instanceCount int)
	DrawIndexed(indexOffset, indexCount, instanceCount int)
	Release()
}
