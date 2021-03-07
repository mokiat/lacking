package command

import (
	"sync"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/mokiat/lacking/opengl"
)

func NewRenderer() *Renderer {
	return &Renderer{
		buffer: NewBuffer(),
	}
}

type Renderer struct {
	renderMU sync.Mutex
	buffer   *Buffer

	activeProgram *opengl.Program
}

func (r *Renderer) Schedule(buffer *Buffer) {
	r.renderMU.Lock()
	defer r.renderMU.Unlock()

	// TODO: Pour commands into renderer buffer
	// while properly sorting and collapsing

	buffer.Overwrite(r.buffer)
}

func (r *Renderer) Render() {
	r.renderMU.Lock()
	defer r.renderMU.Unlock()

	// TODO: Reset cache state

	gl.Enable(gl.FRAMEBUFFER_SRGB)
	r.buffer.Each(func(id ID) {
		switch id.Type {
		case TypeClear:
			r.processClearCommand(r.buffer.ClearCommand(id.Index))
		case TypeDepthConfig:
			r.processDepthConfigCommand(r.buffer.DepthConfigCommand(id.Index))
		case TypeChangeFramebuffer:
			r.processChangeFramebufferCommand(r.buffer.ChangeFramebufferCommand(id.Index))
		case TypeChangeProgram:
			r.processChangeProgramCommand(r.buffer.ChangeProgramCommand(id.Index))
		case TypeBindUniform:
			r.processBindUniformCommand(r.buffer.BindUniformCommand(id.Index))
		case TypeBindTexture:
			r.processBindTextureCommand(r.buffer.BindTextureCommand(id.Index))
		}
	})
}

func (r *Renderer) processClearCommand(cmd Clear) {
	// TODO: Optimize: both can be done in a single pass
	// TODO: Be smarter and clear active (not default) framebuffer
	if cmd.ClearDepth.IsSet() {
		value := cmd.ClearDepth.Value()
		gl.ClearNamedFramebufferfv(0, gl.DEPTH, 0, &value)
	}
	if cmd.ClearStencil.IsSet() {
		value := cmd.ClearStencil.Value()
		gl.ClearNamedFramebufferuiv(0, gl.STENCIL, 0, &value)
	}
	var rgba [4]float32
	for _, clearColor := range cmd.ClearColors {
		if clearColor.IsSet() {
			value := clearColor.Value()
			rgba[0] = value.Color.X
			rgba[1] = value.Color.Y
			rgba[2] = value.Color.Z
			rgba[3] = value.Color.W
			gl.ClearNamedFramebufferfv(0, gl.COLOR, int32(value.Attachment), &rgba[0])
		}
	}
}

func (r *Renderer) processDepthConfigCommand(cmd DepthConfig) {
	if cmd.DepthTest.IsSet() {
		if cmd.DepthTest.Value() {
			gl.Enable(gl.DEPTH_TEST)
		} else {
			gl.Disable(gl.DEPTH_TEST)
		}
	}
	if cmd.DepthWrite.IsSet() {
		if cmd.DepthWrite.Value() {
			gl.DepthMask(true)
		} else {
			gl.DepthMask(false)
		}
	}
	if cmd.DepthFunc.IsSet() {
		gl.DepthFunc(cmd.DepthFunc.Value())
	}
}

func (r *Renderer) processChangeFramebufferCommand(cmd ChangeFramebuffer) {
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

func (r *Renderer) processChangeProgramCommand(cmd ChangeProgram) {
	gl.UseProgram(cmd.Program.ID())
	r.activeProgram = cmd.Program
}

func (r *Renderer) processBindUniformCommand(cmd BindUniform) {
	location := r.activeProgram.UniformLocation(cmd.Name)
	switch cmd.Uniform.Kind {
	case UniformKindMatrix4f:
		gl.UniformMatrix4fv(location, 1, false, &cmd.Uniform.FloatData[0])
	case UniformKind1f:
		gl.Uniform1fv(location, 1, &cmd.Uniform.FloatData[0])
	default:
		panic("unknown uniform kind")
	}
}

func (r *Renderer) processBindTextureCommand(cmd BindTexture) {
	// TODO:
}
