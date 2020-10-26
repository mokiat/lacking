package render

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/graphics"
)

type AmbientLight struct {
	Color             sprec.Vec3
	IrradianceTexture *graphics.CubeTexture
}

type DirectionalLight struct {
	Color     sprec.Vec3
	Direction sprec.Vec3
}
