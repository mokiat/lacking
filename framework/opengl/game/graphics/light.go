package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics"
)

var _ graphics.DirectionalLight = (*Light)(nil)

type Light struct {
	internal.Node

	scene *Scene
	prev  *Light
	next  *Light

	mode      LightMode
	intensity sprec.Vec3
}

func (l *Light) Intensity() sprec.Vec3 {
	return l.intensity
}

func (l *Light) SetIntensity(intensity sprec.Vec3) {
	l.intensity = intensity
}

func (l *Light) Delete() {
	l.scene.detachLight(l)
	l.scene.cacheLight(l)
	l.scene = nil
}

const (
	LightModeDirectional LightMode = 1 + iota
	LightModeAmbient
)

type LightMode int
