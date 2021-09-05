package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
)

type Target struct {
	Framebuffer *opengl.Framebuffer
	Size        sprec.Vec2
}
