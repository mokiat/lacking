package command

import (
	"sync"

	"github.com/go-gl/gl/v4.6-core/gl"
)

// TODO: Optimize by not changing state if there is no difference

func NewRenderer() *Renderer {
	return &Renderer{
		buffer: NewBuffer(),
	}
}

type Renderer struct {
	renderMU sync.Mutex
	buffer   *Buffer
}

func (r *Renderer) Schedule(buffer *Buffer) {
	r.renderMU.Lock()
	defer r.renderMU.Unlock()

	// TODO: Pour commands into renderer buffer while properly sorting
	buffer.Overwrite(r.buffer)
}

func (r *Renderer) Render() {
	r.renderMU.Lock()
	defer r.renderMU.Unlock()

	r.RenderBuffer(r.buffer)
}

func (r *Renderer) RenderBuffer(buffer *Buffer) {
	// TODO: Reset cache state

	gl.Enable(gl.FRAMEBUFFER_SRGB)

	for _, cmd := range buffer.commands {
		switch cmd.Type {
		case TypeChangeFramebuffer:
			r.changeFramebuffer(buffer, buffer.changeFramebufferCommands[cmd.Index])
		case TypeClearFramebuffer:
			r.clearFramebuffer(buffer, buffer.clearFramebufferCommands[cmd.Index])
		case TypeChangeDepthConfig:
			r.changeDepthConfig(buffer, buffer.changeDepthConfigCommands[cmd.Index])
		case TypeRenderItem:
			r.renderItem(buffer, buffer.renderItemCommands[cmd.Index])
		}
	}
}

func (r *Renderer) changeFramebuffer(buffer *Buffer, cmd ChangeFramebuffer) {
	if cmd.Framebuffer.IsSet() {
		framebuffer := cmd.Framebuffer.Value()
		// TODO: Use DSA glNamedFramebufferDrawBuffers
		if framebuffer == nil {
			gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		} else {
			gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer.ID())
		}
	}
	if cmd.Viewport.IsSet() {
		area := cmd.Viewport.Value()
		gl.Viewport(int32(area.X), int32(area.Y), int32(area.Width), int32(area.Height))
	}
	if cmd.Scissor.IsSet() {
		area := cmd.Viewport.Value()
		gl.Scissor(int32(area.X), int32(area.Y), int32(area.Width), int32(area.Height))
	}
}

func (r *Renderer) clearFramebuffer(buffer *Buffer, cmd ClearFramebuffer) {
	// TODO: Optimize: both can be done in a single pass
	// TODO: Be smarter and clear active (not default) framebuffer
	if cmd.Depth.IsSet() {
		value := cmd.Depth.Value()
		gl.ClearNamedFramebufferfv(0, gl.DEPTH, 0, &value)
	}
	if cmd.Stencil.IsSet() {
		value := cmd.Stencil.Value()
		gl.ClearNamedFramebufferuiv(0, gl.STENCIL, 0, &value)
	}
	var rgba [4]float32
	for extra := 0; extra < cmd.Colors.Count; extra++ {
		value := buffer.clearColors[cmd.Colors.Offset+extra]
		rgba[0] = value.Color.X
		rgba[1] = value.Color.Y
		rgba[2] = value.Color.Z
		rgba[3] = value.Color.W
		gl.ClearNamedFramebufferfv(0, gl.COLOR, int32(value.Attachment), &rgba[0])
	}
}

func (r *Renderer) changeDepthConfig(buffer *Buffer, cmd ChangeDepthConfig) {
	if cmd.DepthTest {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
	gl.DepthMask(cmd.DepthWrite)
	gl.DepthFunc(cmd.DepthFunc)
}

func (r *Renderer) renderItem(buffer *Buffer, cmd RenderItem) {
	if cmd.BackfaceCulling {
		gl.Enable(gl.CULL_FACE)
	} else {
		gl.Disable(gl.CULL_FACE)
	}

	gl.UseProgram(cmd.Program.ID())

	textureUnit := uint32(0)
	for extra := 0; extra < cmd.Uniforms.Count; extra++ {
		uniform := buffer.uniformBuffer[cmd.Uniforms.Offset+extra]
		location := cmd.Program.UniformLocation(uniform.Name) // TODO: Cache
		switch uniform.Kind {
		case UniformKindTexture:
			gl.BindTextureUnit(textureUnit, buffer.textureBuffer[uniform.Offset].ID())
			gl.Uniform1i(location, int32(textureUnit))
			textureUnit++
		case UniformKindMatrix4f:
			floatData := &buffer.floatBuffer[uniform.Offset]
			gl.UniformMatrix4fv(location, 1, false, floatData)
		case UniformKind4f:
			// TODO Check if performance is better with Uniform4f
			floatData := &buffer.floatBuffer[uniform.Offset]
			gl.Uniform4fv(location, 1, floatData)
		case UniformKind1f:
			// TODO Check if performance is better with Uniform1f
			floatData := &buffer.floatBuffer[uniform.Offset]
			gl.Uniform1fv(location, 1, floatData)
		default:
			panic("unknown uniform kind")
		}
	}

	gl.BindVertexArray(cmd.VertexArray.ID())
	// gl.LineWidth(2) // TODO: Only in case of line
	gl.DrawElements(cmd.Primitive, cmd.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(cmd.IndexOffsetBytes))
}
