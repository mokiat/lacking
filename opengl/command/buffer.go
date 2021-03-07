package command

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/opengl"
)

// TODO: You can compact most commands by using concrete
// types that are of smaller size (i.e. uint16 instead of int)

const (
	maxCommands                  = 16384
	maxBufferFloats              = 16384
	maxClearCommands             = 64
	maxDepthConfigCommands       = 64
	maxChangeFramebufferCommands = 64
	maxChangeProgramCommands     = 2048
	maxBindUniformCommands       = 2048
	maxBindTextureCommands       = 2048
)

func NewBuffer() *Buffer {
	return &Buffer{
		commands:                  make([]ID, 0, maxCommands),
		floatBuffer:               make([]float32, 0, maxBufferFloats),
		clearCommands:             make([]Clear, 0, maxClearCommands),
		depthConfigCommands:       make([]DepthConfig, 0, maxDepthConfigCommands),
		changeFramebufferCommands: make([]ChangeFramebuffer, 0, maxChangeFramebufferCommands),
		changeProgramCommands:     make([]ChangeProgram, 0, maxChangeProgramCommands),
		bindUniformCommands:       make([]BindUniform, 0, maxBindUniformCommands),
		bindTextureCommands:       make([]BindTexture, 0, maxBindTextureCommands),
	}
}

type Buffer struct {
	commands                  []ID
	floatBuffer               []float32
	clearCommands             []Clear
	depthConfigCommands       []DepthConfig
	changeFramebufferCommands []ChangeFramebuffer
	changeProgramCommands     []ChangeProgram
	bindUniformCommands       []BindUniform
	bindTextureCommands       []BindTexture
}

func (b *Buffer) Reset() {
	b.commands = b.commands[:0]
	b.floatBuffer = b.floatBuffer[:0]
	b.clearCommands = b.clearCommands[:0]
	b.depthConfigCommands = b.depthConfigCommands[:0]
	b.changeFramebufferCommands = b.changeFramebufferCommands[:0]
	b.changeProgramCommands = b.changeProgramCommands[:0]
	b.bindUniformCommands = b.bindUniformCommands[:0]
	b.bindTextureCommands = b.bindTextureCommands[:0]
}

func (b *Buffer) Overwrite(other *Buffer) {
	other.commands = other.commands[:len(b.commands)]
	copy(other.commands, b.commands)

	other.floatBuffer = other.floatBuffer[:len(b.floatBuffer)]
	copy(other.floatBuffer, b.floatBuffer)

	other.clearCommands = other.clearCommands[:len(b.clearCommands)]
	copy(other.clearCommands, b.clearCommands)

	other.depthConfigCommands = other.depthConfigCommands[:len(b.depthConfigCommands)]
	copy(other.depthConfigCommands, b.depthConfigCommands)

	other.changeFramebufferCommands = other.changeFramebufferCommands[:len(b.changeFramebufferCommands)]
	copy(other.changeFramebufferCommands, b.changeFramebufferCommands)

	other.changeProgramCommands = other.changeProgramCommands[:len(b.changeProgramCommands)]
	copy(other.changeProgramCommands, b.changeProgramCommands)

	other.bindUniformCommands = other.bindUniformCommands[:len(b.bindUniformCommands)]
	copy(other.bindUniformCommands, b.bindUniformCommands)

	other.bindTextureCommands = other.bindTextureCommands[:len(b.bindTextureCommands)]
	copy(other.bindTextureCommands, b.bindTextureCommands)
}

func (b *Buffer) Each(fn func(id ID)) {
	for _, id := range b.commands {
		fn(id)
	}
}

func (b *Buffer) AppendClearCommand(cmd Clear) ID {
	id := ID{
		Type:  TypeClear,
		Index: len(b.clearCommands),
	}
	b.commands = append(b.commands, id)
	b.clearCommands = append(b.clearCommands, cmd)
	return id
}

func (b *Buffer) ClearCommand(index int) Clear {
	return b.clearCommands[index]
}

func (b *Buffer) AppendDepthConfigCommand(cmd DepthConfig) ID {
	id := ID{
		Type:  TypeDepthConfig,
		Index: len(b.depthConfigCommands),
	}
	b.commands = append(b.commands, id)
	b.depthConfigCommands = append(b.depthConfigCommands, cmd)
	return id
}

func (b *Buffer) DepthConfigCommand(index int) DepthConfig {
	return b.depthConfigCommands[index]
}

func (b *Buffer) AppendChangeFramebufferCommand(cmd ChangeFramebuffer) ID {
	id := ID{
		Type:  TypeChangeFramebuffer,
		Index: len(b.changeFramebufferCommands),
	}
	b.commands = append(b.commands, id)
	b.changeFramebufferCommands = append(b.changeFramebufferCommands, cmd)
	return id
}

func (b *Buffer) ChangeFramebufferCommand(index int) ChangeFramebuffer {
	return b.changeFramebufferCommands[index]
}

func (b *Buffer) AppendChangeProgramCommand(cmd ChangeProgram) ID {
	id := ID{
		Type:  TypeChangeProgram,
		Index: len(b.changeProgramCommands),
	}
	b.commands = append(b.commands, id)
	b.changeProgramCommands = append(b.changeProgramCommands, cmd)
	return id
}

func (b *Buffer) ChangeProgramCommand(index int) ChangeProgram {
	return b.changeProgramCommands[index]
}

func (b *Buffer) AppendBindUniformCommand(cmd BindUniform) ID {
	id := ID{
		Type:  TypeBindUniform,
		Index: len(b.bindUniformCommands),
	}
	b.commands = append(b.commands, id)
	b.bindUniformCommands = append(b.bindUniformCommands, cmd)
	return id
}

func (b *Buffer) BindUniformCommand(index int) BindUniform {
	return b.bindUniformCommands[index]
}

func (b *Buffer) AppendBindTextureCommand(cmd BindTexture) ID {
	id := ID{
		Type:  TypeBindTexture,
		Index: len(b.bindTextureCommands),
	}
	b.commands = append(b.commands, id)
	b.bindTextureCommands = append(b.bindTextureCommands, cmd)
	return id
}

func (b *Buffer) BindTextureCommand(index int) BindTexture {
	return b.bindTextureCommands[index]
}

func (b *Buffer) ClearColor(attachment int, color sprec.Vec4) {
	b.AppendClearCommand(Clear{
		ClearColors: [8]OptionalClearColor{
			SpecifiedClearColor(ClearColor{
				Attachment: attachment,
				Color:      color,
			}),
		},
	})
}

func (b *Buffer) ClearDepth(value float32) {
	b.AppendClearCommand(Clear{
		ClearDepth: SpecifiedFloat32(value),
	})
}

func (b *Buffer) ClearStencil(value uint32) {
	b.AppendClearCommand(Clear{
		ClearStencil: SpecifiedUint32(value),
	})
}

func (b *Buffer) EnableDepthTest() {
	b.AppendDepthConfigCommand(DepthConfig{
		DepthTest: SpecifiedBool(true),
	})
}

func (b *Buffer) DisableDepthTest() {
	b.AppendDepthConfigCommand(DepthConfig{
		DepthTest: SpecifiedBool(false),
	})
}

func (b *Buffer) EnableDepthWrite() {
	b.AppendDepthConfigCommand(DepthConfig{
		DepthWrite: SpecifiedBool(true),
	})
}

func (b *Buffer) DisableDepthWrite() {
	b.AppendDepthConfigCommand(DepthConfig{
		DepthWrite: SpecifiedBool(false),
	})
}

func (b *Buffer) UseDepthFunc(value uint32) {
	b.AppendDepthConfigCommand(DepthConfig{
		DepthFunc: SpecifiedUint32(value),
	})
}

func (b *Buffer) UseFramebuffer(framebuffer *opengl.Framebuffer) {
	b.AppendChangeFramebufferCommand(ChangeFramebuffer{
		Framebuffer: SpecifiedFramebuffer(framebuffer),
	})
}

func (b *Buffer) UseViewport(area opengl.Area) {
	b.AppendChangeFramebufferCommand(ChangeFramebuffer{
		Viewport: SpecifiedArea(area),
	})
}

func (b *Buffer) UseScissor(area opengl.Area) {
	b.AppendChangeFramebufferCommand(ChangeFramebuffer{
		Scissor: SpecifiedArea(area),
	})
}

func (b *Buffer) UseProgram(program *opengl.Program) {
	b.AppendChangeProgramCommand(ChangeProgram{
		Program: program,
	})
}

func (b *Buffer) UniformMatrix4f(value sprec.Mat4) Uniform {
	offset := len(b.floatBuffer)
	b.floatBuffer = append(b.floatBuffer,
		value.M11, value.M21, value.M31, value.M41,
		value.M12, value.M22, value.M32, value.M42,
		value.M13, value.M23, value.M33, value.M43,
		value.M14, value.M24, value.M34, value.M44,
	)
	return Uniform{
		Kind:      UniformKindMatrix4f,
		FloatData: b.floatBuffer[offset : offset+16],
	}
}

func (b *Buffer) Uniform1f(value float32) Uniform {
	offset := len(b.floatBuffer)
	b.floatBuffer = append(b.floatBuffer, value)
	return Uniform{
		Kind:      UniformKind1f,
		FloatData: b.floatBuffer[offset : offset+1],
	}
}

func (b *Buffer) BindUniform(name string, uniform Uniform) {
	b.AppendBindUniformCommand(BindUniform{
		Name:    name,
		Uniform: uniform,
	})
}

func (b *Buffer) BindTexture(name string, texture *opengl.Texture) {
	b.AppendBindTextureCommand(BindTexture{
		Name:    name,
		Texture: texture,
	})
}
