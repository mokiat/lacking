package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

type CascadeShadowMap struct {
	Texture     render.Texture
	Framebuffer render.Framebuffer
}

type CascadeShadowMapRef struct {
	CascadeShadowMap
	ProjectionMatrix sprec.Mat4
}

type AtlasShadowMap struct {
	Texture     render.Texture
	Framebuffer render.Framebuffer
	// TODO: Viewport?
}

type AtlasShadowMapRef struct {
	AtlasShadowMap
	// TODO: Projection matrix?
	// TODO: View matrix?
}
