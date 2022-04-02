package render

type FramebufferInfo struct {
	ColorAttachments       []Texture
	DepthAttachment        Texture
	StencilAttachment      Texture
	DepthStencilAttachment Texture
}

type Framebuffer interface {
	Release()
}
