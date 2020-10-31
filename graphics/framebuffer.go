package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Framebuffer struct {
	ID              uint32
	AlbedoTextureID uint32
	NormalTextureID uint32
	DepthTextureID  uint32
	Width           int32
	Height          int32
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

type FramebufferData struct {
	Width               int32
	Height              int32
	HasAlbedoAttachment bool
	UsesHDRAlbedo       bool
	HasNormalAttachment bool
	HasDepthAttachment  bool
}

func (b *Framebuffer) Allocate(data FramebufferData) error {
	b.Width = data.Width
	b.Height = data.Height

	gl.GenFramebuffers(1, &b.ID)
	gl.BindFramebuffer(gl.FRAMEBUFFER, b.ID)

	var drawBufferIDS []uint32
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
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, b.AlbedoTextureID, 0)
		drawBufferIDS = append(drawBufferIDS, gl.COLOR_ATTACHMENT0)
	}
	if data.HasNormalAttachment {
		// NORMAL X, NORMAL Y, NORMAL Z, ROUGHNESS
		gl.GenTextures(1, &b.NormalTextureID)
		gl.BindTexture(gl.TEXTURE_2D, b.NormalTextureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, data.Width, data.Height, 0, gl.RGB, gl.FLOAT, gl.Ptr(nil))
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, b.NormalTextureID, 0)
		drawBufferIDS = append(drawBufferIDS, gl.COLOR_ATTACHMENT1)
	}
	if data.HasDepthAttachment {
		gl.GenTextures(1, &b.DepthTextureID)
		gl.BindTexture(gl.TEXTURE_2D, b.DepthTextureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT, data.Width, data.Height, 0, gl.DEPTH_COMPONENT, gl.FLOAT, gl.Ptr(nil))
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, b.DepthTextureID, 0)
	}
	gl.DrawBuffers(int32(len(drawBufferIDS)), &drawBufferIDS[0])

	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		return fmt.Errorf("framebuffer has incomplete status: %d", status)
	}
	return nil
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
	gl.DeleteFramebuffers(1, &b.ID)
	b.ID = 0
	return nil
}
