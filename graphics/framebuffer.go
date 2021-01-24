package graphics

import (
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/mokiat/lacking/opengl"
)

type FramebufferData struct {
	Width               int32
	Height              int32
	HasAlbedoAttachment bool
	UsesHDRAlbedo       bool
	HasNormalAttachment bool
	HasDepthAttachment  bool
}

type Framebuffer struct {
	Framebuffer *opengl.Framebuffer

	Width           int32
	Height          int32
	AlbedoTextureID uint32
	NormalTextureID uint32
	DepthTextureID  uint32
}

func (b *Framebuffer) ID() uint32 {
	if b == nil || b.Framebuffer == nil {
		return 0
	}
	return b.Framebuffer.ID()
}

func (b Framebuffer) HasAlbedoAttachment() bool {
	return b.AlbedoTextureID != 0
}

func (b Framebuffer) HasNormalAttachment() bool {
	return b.NormalTextureID != 0
}

func (b Framebuffer) HasDepthAttachment() bool {
	return b.DepthTextureID != 0
}

func (b *Framebuffer) Allocate(data FramebufferData) error {
	b.Width = data.Width
	b.Height = data.Height

	framebufferInfo := opengl.FramebufferAllocateInfo{}

	if data.HasAlbedoAttachment {
		// R, G, B, METALNESS
		gl.GenTextures(1, &b.AlbedoTextureID)
		gl.BindTexture(gl.TEXTURE_2D, b.AlbedoTextureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		if data.UsesHDRAlbedo {
			gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, data.Width, data.Height, 0, gl.RGBA, gl.FLOAT, gl.Ptr(nil))
		} else {
			gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, data.Width, data.Height, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(nil))
		}
		framebufferInfo.ColorAttachments = append(framebufferInfo.ColorAttachments,
			opengl.NewTexture(b.AlbedoTextureID),
		)
	}
	if data.HasNormalAttachment {
		// NORMAL X, NORMAL Y, NORMAL Z, ROUGHNESS
		gl.GenTextures(1, &b.NormalTextureID)
		gl.BindTexture(gl.TEXTURE_2D, b.NormalTextureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, data.Width, data.Height, 0, gl.RGB, gl.FLOAT, gl.Ptr(nil))
		framebufferInfo.ColorAttachments = append(framebufferInfo.ColorAttachments,
			opengl.NewTexture(b.NormalTextureID),
		)
	}
	if data.HasDepthAttachment {
		gl.GenTextures(1, &b.DepthTextureID)
		gl.BindTexture(gl.TEXTURE_2D, b.DepthTextureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT, data.Width, data.Height, 0, gl.DEPTH_COMPONENT, gl.FLOAT, gl.Ptr(nil))
		framebufferInfo.DepthAttachment = opengl.NewTexture(b.DepthTextureID)
	}

	b.Framebuffer = opengl.NewFramebuffer()
	return b.Framebuffer.Allocate(framebufferInfo)
}

func (b *Framebuffer) Release() error {
	if b.HasAlbedoAttachment() {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, 0, 0)
		gl.DeleteTextures(1, &b.AlbedoTextureID)
		b.AlbedoTextureID = 0
	}
	if b.HasNormalAttachment() {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, 0, 0)
		gl.DeleteTextures(1, &b.NormalTextureID)
		b.NormalTextureID = 0
	}
	if b.HasDepthAttachment() {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, 0, 0)
		gl.DeleteTextures(1, &b.DepthTextureID)
		b.DepthTextureID = 0
	}
	return b.Framebuffer.Release()
}
