package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

type DirectionalShadowMap struct {
	ArrayTexture render.Texture
	Cascades     []DirectionalShadowMapCascade
}

type DirectionalShadowMapCascade struct {
	Framebuffer      render.Framebuffer
	ProjectionMatrix sprec.Mat4
	Near             float32
	Far              float32
}

type DirectionalShadowMapRef struct {
	DirectionalShadowMap
}

type SpotShadowMap struct {
	Texture     render.Texture
	Framebuffer render.Framebuffer
}

type PointShadowMap struct {
	ArrayTexture render.Texture
	Framebuffers [6]render.Framebuffer
}
