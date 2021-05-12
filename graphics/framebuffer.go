package graphics

import (
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/mokiat/lacking/framework/opengl"
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

	Width         int32
	Height        int32
	AlbedoTexture *opengl.TwoDTexture
	NormalTexture *opengl.TwoDTexture
	DepthTexture  *opengl.TwoDTexture
}

func (b *Framebuffer) ID() uint32 {
	if b == nil || b.Framebuffer == nil {
		return 0
	}
	return b.Framebuffer.ID()
}

func (b Framebuffer) HasAlbedoAttachment() bool {
	return b.AlbedoTexture != nil
}

func (b Framebuffer) HasNormalAttachment() bool {
	return b.NormalTexture != nil
}

func (b Framebuffer) HasDepthAttachment() bool {
	return b.DepthTexture != nil
}

func (b *Framebuffer) Allocate(data FramebufferData) error {
	b.Width = data.Width
	b.Height = data.Height

	framebufferInfo := opengl.FramebufferAllocateInfo{}
	if data.HasAlbedoAttachment {
		// R, G, B, METALNESS

		allocateInfo := opengl.TwoDTextureAllocateInfo{
			Width:     data.Width,
			Height:    data.Height,
			MinFilter: gl.NEAREST,
			MagFilter: gl.NEAREST,
			Data:      nil,
		}
		if data.UsesHDRAlbedo {
			allocateInfo.InternalFormat = gl.RGBA32F
			allocateInfo.DataFormat = gl.RGBA
			allocateInfo.DataComponentType = gl.FLOAT
		} else {
			allocateInfo.InternalFormat = gl.RGBA8
			allocateInfo.DataFormat = gl.RGBA
			allocateInfo.DataComponentType = gl.UNSIGNED_BYTE
		}
		b.AlbedoTexture = opengl.NewTwoDTexture()
		b.AlbedoTexture.Allocate(allocateInfo)
		framebufferInfo.ColorAttachments = append(framebufferInfo.ColorAttachments,
			&b.AlbedoTexture.Texture,
		)
	}
	if data.HasNormalAttachment {
		// NORMAL X, NORMAL Y, NORMAL Z, ROUGHNESS

		allocateInfo := opengl.TwoDTextureAllocateInfo{
			Width:             data.Width,
			Height:            data.Height,
			MinFilter:         gl.NEAREST,
			MagFilter:         gl.NEAREST,
			InternalFormat:    gl.RGBA32F,
			DataFormat:        gl.RGBA,
			DataComponentType: gl.FLOAT,
			Data:              nil,
		}
		b.NormalTexture = opengl.NewTwoDTexture()
		b.NormalTexture.Allocate(allocateInfo)
		framebufferInfo.ColorAttachments = append(framebufferInfo.ColorAttachments,
			&b.NormalTexture.Texture,
		)
	}
	if data.HasDepthAttachment {
		allocateInfo := opengl.TwoDTextureAllocateInfo{
			Width:             data.Width,
			Height:            data.Height,
			MinFilter:         gl.NEAREST,
			MagFilter:         gl.NEAREST,
			InternalFormat:    gl.DEPTH_COMPONENT32,
			DataFormat:        gl.DEPTH_COMPONENT,
			DataComponentType: gl.FLOAT,
			Data:              nil,
		}
		b.DepthTexture = opengl.NewTwoDTexture()
		b.DepthTexture.Allocate(allocateInfo)
		framebufferInfo.DepthAttachment = &b.DepthTexture.Texture
	}

	b.Framebuffer = opengl.NewFramebuffer()
	b.Framebuffer.Allocate(framebufferInfo)
	return nil
}

func (b *Framebuffer) Release() error {
	b.Framebuffer.Release()
	b.Framebuffer = nil
	if b.HasAlbedoAttachment() {
		b.AlbedoTexture.Release()
		b.AlbedoTexture = nil
	}
	if b.HasNormalAttachment() {
		b.NormalTexture.Release()
		b.NormalTexture = nil
	}
	if b.HasDepthAttachment() {
		b.DepthTexture.Release()
		b.DepthTexture = nil
	}
	return nil
}
