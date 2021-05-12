package opengl

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func NewFramebuffer() *Framebuffer {
	return &Framebuffer{}
}

type Framebuffer struct {
	id uint32
}

func (b *Framebuffer) ID() uint32 {
	return b.id
}

func (b *Framebuffer) Allocate(info FramebufferAllocateInfo) {
	if b.id != 0 {
		panic(fmt.Errorf("framebuffer already allocated"))
	}
	gl.CreateFramebuffers(1, &b.id)
	if b.id == 0 {
		panic(fmt.Errorf("failed to allocate framebuffer"))
	}

	var drawBuffers []uint32
	for i, colorAttachment := range info.ColorAttachments {
		if colorAttachment != nil {
			attachmentID := gl.COLOR_ATTACHMENT0 + uint32(i)
			gl.NamedFramebufferTexture(b.id, attachmentID, colorAttachment.ID(), 0)
			drawBuffers = append(drawBuffers, attachmentID)
		}
	}
	if info.DepthStencilAttachment != nil {
		gl.NamedFramebufferTexture(b.id, gl.DEPTH_STENCIL_ATTACHMENT, info.DepthStencilAttachment.ID(), 0)
	} else {
		if info.DepthAttachment != nil {
			gl.NamedFramebufferTexture(b.id, gl.DEPTH_ATTACHMENT, info.DepthAttachment.ID(), 0)
		}
		if info.StencilAttachment != nil {
			gl.NamedFramebufferTexture(b.id, gl.STENCIL_ATTACHMENT, info.StencilAttachment.ID(), 0)
		}
	}
	gl.NamedFramebufferDrawBuffers(b.id, int32(len(drawBuffers)), &drawBuffers[0])
	runtime.KeepAlive(drawBuffers)

	if gl.CheckNamedFramebufferStatus(b.id, gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(fmt.Errorf("framebuffer is incomplete"))
	}
}

func (b *Framebuffer) Release() {
	if b.id == 0 {
		panic(fmt.Errorf("framebuffer already released"))
	}
	gl.DeleteFramebuffers(1, &b.id)
	b.id = 0
}

type FramebufferAllocateInfo struct {
	ColorAttachments       []*Texture
	DepthAttachment        *Texture
	StencilAttachment      *Texture
	DepthStencilAttachment *Texture
}
