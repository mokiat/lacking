package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics"
)

var _ graphics.DirectionalLight = (*Light)(nil)
var _ graphics.AmbientLight = (*Light)(nil)

type Light struct {
	internal.Node

	scene *Scene
	prev  *Light
	next  *Light

	mode              LightMode
	intensity         sprec.Vec3
	reflectionTexture *CubeTexture
	refractionTexture *CubeTexture
}

func (l *Light) Intensity() sprec.Vec3 {
	return l.intensity
}

func (l *Light) SetIntensity(intensity sprec.Vec3) {
	l.intensity = intensity
}

func (l *Light) ReflectionTexture() graphics.CubeTexture {
	return l.reflectionTexture
}

func (l *Light) SetReflectionTexture(texture graphics.CubeTexture) {
	l.reflectionTexture = texture.(*CubeTexture)
}

func (l *Light) RefractionTexture() graphics.CubeTexture {
	return l.refractionTexture
}

func (l *Light) SetRefractionTexture(texture graphics.CubeTexture) {
	l.refractionTexture = texture.(*CubeTexture)
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
