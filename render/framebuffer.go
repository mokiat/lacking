package render

// FramebufferMarker marks a type as being a Framebuffer.
type FramebufferMarker interface {
	_isFramebufferType()
}

// Framebuffer represents a combination of target textures to be rendered to.
type Framebuffer interface {
	FramebufferMarker
	Resource
}

// FramebufferInfo describes the configuration of a Framebuffer.
type FramebufferInfo struct {

	// Label specifies a human-readable name for the Framebuffer. Intended
	// for debugging and logging purposes only.
	Label string

	// ColorAttachments is the list of color attachments that should be
	// attached to the Framebuffer.
	ColorAttachments [4]Texture

	// DepthAttachment is the depth attachment that should be attached to
	// the Framebuffer.
	DepthAttachment Texture

	// StencilAttachment is the stencil attachment that should be attached
	// to the Framebuffer.
	StencilAttachment Texture

	// DepthStencilAttachment is the depth+stencil attachment that should
	// be attached to the Framebuffer.
	DepthStencilAttachment Texture
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
