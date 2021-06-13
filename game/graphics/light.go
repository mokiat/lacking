package graphics

import "github.com/mokiat/gomath/sprec"

// Light represents a light emitting object in the scene.
type Light interface {
	Node

	// Intensity returns the light intensity.
	Intensity() sprec.Vec3

	// SetIntensity changes the light intensity.
	SetIntensity(intensity sprec.Vec3)

	// Delete removes this light source.
	Delete()
}

// AmbientLight is a light source that emits light from the
// scene surroundings.
type AmbientLight interface {
	Light

	// ReflectionTexture returns the texture that is used to calculate
	// the lighting on an object as a result of reflected light rays.
	ReflectionTexture() CubeTexture

	// SetReflectionTexture changes the reflection texture.
	SetReflectionTexture(texture CubeTexture)

	// RefractionTexture returns the texture that is used to calculate
	// the lighting on an object as a result of refracted light rays.
	RefractionTexture() CubeTexture

	// SetRefractionTexture changes the refraction texture.
	SetRefractionTexture(texture CubeTexture)
}

// DirectionalLight is a light object that emits parallel light
// rays into a single direction (going into the Z direction).
type DirectionalLight interface {
	Light
}
