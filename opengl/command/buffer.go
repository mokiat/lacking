package command

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/opengl"
)

// TODO: You can compact most commands by using concrete
// types that are of smaller size (i.e. uint16 instead of int)

const (
	maxCommands = 16384

	maxChangeFramebufferCommands = 64
	maxClearFramebufferCommands  = 64
	maxChangeDepthConfigCommands = 64
	maxRenderItemCommands        = 2048

	maxUniforms    = 2048
	maxFloats      = 16384
	maxTextures    = 1024
	maxClearColors = 256
)

func NewBuffer() *Buffer {
	return &Buffer{
		commands: make([]ID, 0, maxCommands),

		changeFramebufferCommands: make([]ChangeFramebuffer, 0, maxChangeFramebufferCommands),
		clearFramebufferCommands:  make([]ClearFramebuffer, 0, maxClearFramebufferCommands),
		changeDepthConfigCommands: make([]ChangeDepthConfig, 0, maxChangeDepthConfigCommands),
		renderItemCommands:        make([]RenderItem, 0, maxRenderItemCommands),

		uniformBuffer: make([]Uniform, 0, maxUniforms),
		floatBuffer:   make([]float32, 0, maxFloats),
		textureBuffer: make([]*opengl.Texture, 0, maxTextures),
		clearColors:   make([]ClearColor, 0, maxClearColors),
	}
}

type Buffer struct {
	commands []ID

	changeFramebufferCommands []ChangeFramebuffer
	clearFramebufferCommands  []ClearFramebuffer
	changeDepthConfigCommands []ChangeDepthConfig
	renderItemCommands        []RenderItem

	uniformBuffer []Uniform
	floatBuffer   []float32
	textureBuffer []*opengl.Texture
	clearColors   []ClearColor
}

func (b *Buffer) Reset() {
	b.commands = b.commands[:0]

	b.changeFramebufferCommands = b.changeFramebufferCommands[:0]
	b.clearFramebufferCommands = b.clearFramebufferCommands[:0]
	b.changeDepthConfigCommands = b.changeDepthConfigCommands[:0]
	b.renderItemCommands = b.renderItemCommands[:0]

	b.uniformBuffer = b.uniformBuffer[:0]
	b.floatBuffer = b.floatBuffer[:0]
	b.textureBuffer = b.textureBuffer[:0]
	b.clearColors = b.clearColors[:0]
}

func (b *Buffer) Overwrite(other *Buffer) {
	other.commands = other.commands[:len(b.commands)]
	copy(other.commands, b.commands)

	other.changeFramebufferCommands = other.changeFramebufferCommands[:len(b.changeFramebufferCommands)]
	copy(other.changeFramebufferCommands, b.changeFramebufferCommands)

	other.clearFramebufferCommands = other.clearFramebufferCommands[:len(b.clearFramebufferCommands)]
	copy(other.clearFramebufferCommands, b.clearFramebufferCommands)

	other.changeDepthConfigCommands = other.changeDepthConfigCommands[:len(b.changeDepthConfigCommands)]
	copy(other.changeDepthConfigCommands, b.changeDepthConfigCommands)

	other.renderItemCommands = other.renderItemCommands[:len(b.renderItemCommands)]
	copy(other.renderItemCommands, b.renderItemCommands)

	other.uniformBuffer = other.uniformBuffer[:len(b.uniformBuffer)]
	copy(other.uniformBuffer, b.uniformBuffer)

	other.floatBuffer = other.floatBuffer[:len(b.floatBuffer)]
	copy(other.floatBuffer, b.floatBuffer)

	other.textureBuffer = other.textureBuffer[:len(b.textureBuffer)]
	copy(other.textureBuffer, b.textureBuffer)

	other.clearColors = other.clearColors[:len(b.clearColors)]
	copy(other.clearColors, b.clearColors)
}

func (b *Buffer) Append(cmd ID) {
	b.commands = append(b.commands, cmd)
}

func (b *Buffer) Each(fn func(id ID)) {
	for _, id := range b.commands {
		fn(id)
	}
}

func (b *Buffer) ChangeFramebuffer(
	framebuffer *opengl.Framebuffer,
	viewport opengl.Area,
	scissor opengl.Area,
) ID {
	b.changeFramebufferCommands = append(b.changeFramebufferCommands, ChangeFramebuffer{
		Framebuffer: SpecifiedFramebuffer(framebuffer),
		Viewport:    SpecifiedArea(viewport),
		Scissor:     SpecifiedArea(scissor),
	})
	return ID{
		Type:  TypeChangeFramebuffer,
		Index: len(b.changeFramebufferCommands) - 1,
	}
}

func (b *Buffer) ClearFramebuffer(
	colors ClearColorRange,
	depth OptionalFloat32,
	stencil OptionalUint32,
) ID {
	b.clearFramebufferCommands = append(b.clearFramebufferCommands, ClearFramebuffer{
		Colors:  colors,
		Depth:   depth,
		Stencil: stencil,
	})
	return ID{
		Type:  TypeClearFramebuffer,
		Index: len(b.clearFramebufferCommands) - 1,
	}
}

func (b *Buffer) ClearColorRange(colors ...ClearColor) ClearColorRange {
	b.clearColors = append(b.clearColors, colors...)
	return ClearColorRange{
		Offset: len(b.clearColors) - len(colors),
		Count:  len(colors),
	}
}

func (b *Buffer) ChangeDepthConfig(
	depthTest bool,
	depthWrite bool,
	depthFunc uint32,
) ID {
	b.changeDepthConfigCommands = append(b.changeDepthConfigCommands, ChangeDepthConfig{
		DepthTest:  depthTest,
		DepthWrite: depthWrite,
		DepthFunc:  depthFunc,
	})
	return ID{
		Type:  TypeChangeDepthConfig,
		Index: len(b.changeDepthConfigCommands) - 1,
	}
}

func (b *Buffer) ReorderConfig(enabled bool) ID {
	return ID{
		Type:  TypeChangeReorderConfig,
		Index: -1, // TODO
	}
}

func (b *Buffer) RenderItem(
	backfaceCulling bool,
	program *opengl.Program,
	uniforms UniformRange,
	vertexArray *opengl.VertexArray,
	primitive uint32,
	indexCount int32,
	indexOffsetBytes int,
) ID {
	b.renderItemCommands = append(b.renderItemCommands, RenderItem{
		BackfaceCulling:  backfaceCulling,
		Program:          program,
		Uniforms:         uniforms,
		VertexArray:      vertexArray,
		Primitive:        primitive,
		IndexCount:       indexCount,
		IndexOffsetBytes: indexOffsetBytes,
	})
	return ID{
		Type:  TypeRenderItem,
		Index: len(b.renderItemCommands) - 1,
	}
}

func (b *Buffer) UniformRange(uniforms ...Uniform) UniformRange {
	b.uniformBuffer = append(b.uniformBuffer, uniforms...)
	return UniformRange{
		Offset: len(b.uniformBuffer) - len(uniforms),
		Count:  len(uniforms),
	}
}

func (b *Buffer) UniformTexture(name string, value *opengl.Texture) Uniform {
	b.textureBuffer = append(b.textureBuffer, value)
	return Uniform{
		Name:   name,
		Kind:   UniformKindTexture,
		Offset: len(b.textureBuffer) - 1,
	}
}

func (b *Buffer) UniformMatrix4f(name string, value sprec.Mat4) Uniform {
	b.floatBuffer = append(b.floatBuffer,
		value.M11, value.M21, value.M31, value.M41,
		value.M12, value.M22, value.M32, value.M42,
		value.M13, value.M23, value.M33, value.M43,
		value.M14, value.M24, value.M34, value.M44,
	)
	return Uniform{
		Name:   name,
		Kind:   UniformKindMatrix4f,
		Offset: len(b.floatBuffer) - 16,
	}
}
