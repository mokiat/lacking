package render

type FramebufferInfo struct {
	ColorAttachments       []Texture
	DepthAttachment        Texture
	StencilAttachment      Texture
	DepthStencilAttachment Texture
}

type CopyContentToTextureInfo struct {
	Texture         Texture
	TextureLevel    int
	TextureX        int
	TextureY        int
	FramebufferX    int
	FramebufferY    int
	Width           int
	Height          int
	GenerateMipmaps bool
}

type Framebuffer interface {
	Release()
}
