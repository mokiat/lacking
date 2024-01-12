package render

// FramebufferObject marks a type as being a Framebuffer object.
type FramebufferObject interface {
	_isFramebufferObject() bool // ensures interface uniqueness
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

// Framebuffer represents a combination of target textures to be rendered to.
type Framebuffer interface {
	FramebufferObject

	// Release releases the resources associated with the Framebuffer.
	Release()
}

// ColorTexture2DInfo describes the configuration of a copy operation from
// the current framebuffer to a texture.
type CopyContentToTextureInfo struct {

	// Texture is the texture that should be updated with the contents of
	// the current framebuffer.
	Texture Texture

	// TextureLevel is the mipmap level of the texture that should be updated.
	TextureLevel int

	// TextureX is the X offset of the texture that should be updated.
	TextureX int

	// TextureY is the Y offset of the texture that should be updated.
	TextureY int

	// FramebufferX is the X offset of the framebuffer that should be copied.
	FramebufferX int

	// FramebufferY is the Y offset of the framebuffer that should be copied.
	FramebufferY int

	// Width is the width amount of the framebuffer that should be copied.
	Width int

	// Height is the height amount of the framebuffer that should be copied.
	Height int

	// GenerateMipmaps indicates whether or not mipmaps should be generated.
	GenerateMipmaps bool
}

// CopyContentToBufferInfo describes the configuration of a copy operation
// from the current framebuffer to a pixel transfer buffer.
type CopyContentToBufferInfo struct {

	// Buffer is the pixel transfer buffer that should be updated with the
	// contents of the current framebuffer.
	Buffer Buffer

	// X is the X offset of the framebuffer that should be copied.
	X int

	// Y is the Y offset of the framebuffer that should be copied.
	Y int

	// Width is the width amount of the framebuffer that should be copied.
	Width int

	// Height is the height amount of the framebuffer that should be copied.
	Height int

	// Format is the format of the pixel transfer buffer.
	Format DataFormat

	// Offset is the offset into the pixel transfer buffer.
	Offset int
}
