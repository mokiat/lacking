package graphics

import "github.com/mokiat/gomath/sprec"

// Light represents a light emitting object in the scene.
type Light struct {
	Node

	scene *Scene
	prev  *Light
	next  *Light

	mode              LightMode
	intensity         sprec.Vec3
	reflectionTexture *CubeTexture
	refractionTexture *CubeTexture
}

func (l *Light) Mode() LightMode {
	return l.mode
}

func (l *Light) SetMode(mode LightMode) {
	l.mode = mode
}

// Intensity returns the light intensity.
func (l *Light) Intensity() sprec.Vec3 {
	return l.intensity
}

// SetIntensity changes the light intensity.
func (l *Light) SetIntensity(intensity sprec.Vec3) {
	l.intensity = intensity
}

// ReflectionTexture returns the texture that is used to calculate
// the lighting on an object as a result of reflected light rays.
func (l *Light) ReflectionTexture() *CubeTexture {
	return l.reflectionTexture
}

// SetReflectionTexture changes the reflection texture.
func (l *Light) SetReflectionTexture(texture *CubeTexture) {
	l.reflectionTexture = texture
}

// RefractionTexture returns the texture that is used to calculate
// the lighting on an object as a result of refracted light rays.
func (l *Light) RefractionTexture() *CubeTexture {
	return l.refractionTexture
}

// SetRefractionTexture changes the refraction texture.
func (l *Light) SetRefractionTexture(texture *CubeTexture) {
	l.refractionTexture = texture
}

func (l *Light) Delete() {
	l.scene.detachLight(l)
	l.scene.cacheLight(l)
	l.scene = nil
}

const (
	LightModeDirectional LightMode = 1 + iota
	LightModeAmbient
	LightModePoint
)

type LightMode int
