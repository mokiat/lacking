package internal

import "github.com/mokiat/lacking/render"

type CascadeShadowMap struct {
	Texture     render.Texture
	Framebuffer render.Framebuffer
}

type AtlasShadowMap struct {
	Texture     render.Texture
	Framebuffer render.Framebuffer
}
