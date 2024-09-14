package render

import "github.com/mokiat/gog/opt"

// FramebufferMarker marks a type as being a Framebuffer.
type FramebufferMarker interface {
	_isFramebufferType()
}

// Framebuffer represents a combination of target textures to be rendered to.
type Framebuffer interface {
	FramebufferMarker
	Resource

	// Label returns a human-readable name for the Framebuffer.
	Label() string
}

// FramebufferInfo describes the configuration of a Framebuffer.
type FramebufferInfo struct {

	// Label specifies a human-readable name for the Framebuffer. Intended
	// for debugging and logging purposes only.
	Label string

	// ColorAttachments is the list of color attachments that should be
	// attached to the Framebuffer.
	ColorAttachments [4]opt.T[TextureAttachment]

	// DepthAttachment is the depth attachment that should be attached to
	// the Framebuffer.
	DepthAttachment opt.T[TextureAttachment]

	// StencilAttachment is the stencil attachment that should be attached
	// to the Framebuffer.
	StencilAttachment opt.T[TextureAttachment]

	// DepthStencilAttachment is the depth+stencil attachment that should
	// be attached to the Framebuffer.
	DepthStencilAttachment opt.T[TextureAttachment]
}

// PlainTextureAttachment creates a TextureAttachment that for the specified
// texture at the root mipmap layer and depth.
func PlainTextureAttachment(texture Texture) TextureAttachment {
	return TextureAttachment{
		Texture: texture,
	}
}

// TextureAttachment represents a framebuffer attachment.
type TextureAttachment struct {

	// Texture is the texture that should be attached.
	Texture Texture

	// Depth is the depth of the texture that should be attached, in case of a
	// texture array or 3D texture.
	Depth uint32

	// MipmapLayer is the mipmap level of the texture that should be attached.
	MipmapLayer uint32
}

// CopyFramebufferToTextureInfo describes the configuration of a copy operation
// from the current framebuffer to a texture.
type CopyFramebufferToTextureInfo struct {

	// Texture is the texture that should be updated with the contents of
	// the current framebuffer.
	Texture Texture

	// TextureLevel is the mipmap level of the texture that should be updated.
	TextureLevel uint32

	// TextureX is the X offset of the texture that should be updated.
	TextureX uint32

	// TextureY is the Y offset of the texture that should be updated.
	TextureY uint32

	// FramebufferX is the X offset of the framebuffer that should be copied.
	FramebufferX uint32

	// FramebufferY is the Y offset of the framebuffer that should be copied.
	FramebufferY uint32

	// Width is the width amount of the framebuffer that should be copied.
	Width uint32

	// Height is the height amount of the framebuffer that should be copied.
	Height uint32

	// GenerateMipmaps indicates whether or not mipmaps should be generated.
	GenerateMipmaps bool
}

// CopyFramebufferToBufferInfo describes the configuration of a copy operation
// from the current framebuffer to a pixel transfer buffer.
type CopyFramebufferToBufferInfo struct {

	// Buffer is the pixel transfer buffer that should be updated with the
	// contents of the current framebuffer.
	Buffer Buffer

	// Offset is the offset into the pixel transfer buffer.
	Offset uint32

	// X is the X offset of the framebuffer that should be copied.
	X uint32

	// Y is the Y offset of the framebuffer that should be copied.
	Y uint32

	// Width is the width amount of the framebuffer that should be copied.
	Width uint32

	// Height is the height amount of the framebuffer that should be copied.
	Height uint32

	// Format is the format of the data.
	Format DataFormat
}
