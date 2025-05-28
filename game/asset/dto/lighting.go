package dto

import "github.com/mokiat/gomath/dprec"

const LightingChunkID = "lacking:lighting"

type LightingChunkHolder struct {
	LightingChunk *LightingChunk `chunk:"lacking:lighting"`
}

type LightingChunk struct {
	// AmbientLights is the collection of ambient lights that are part of the
	// scene.
	AmbientLights []AmbientLight

	// PointLights is the collection of point lights that are part of the scene.
	PointLights []PointLight

	// SpotLights is the collection of spot lights that are part of the scene.
	SpotLights []SpotLight

	// DirectionalLights is the collection of directional lights that are part
	// of the scene.
	DirectionalLights []DirectionalLight
}

// AmbientLight represents a light source that emits light in all directions
// from all points in space.
type AmbientLight struct {

	// ID is the unique identifier of the light within the file.
	ID uint32

	// NodeID is the ID of the node that is associated with the light.
	NodeID uint32

	// ReflectionTextureID is the ID of the cube texture that is used
	// for reflection mapping.
	ReflectionTextureID uint32

	// RefractionTextureID is the ID of the cube texture that is used
	// for refraction mapping.
	RefractionTextureID uint32

	// CastShadow specifies whether a SSAO-type technique should be applied.
	CastShadow bool
}

// PointLight represents a light source that emits light in all directions
// from a single point in space.
type PointLight struct {

	// ID is the unique identifier of the light within the file.
	ID uint32

	// NodeID is the ID of the node that is associated with the light.
	NodeID uint32

	// EmitColor is the linear color of the light that is emitted.
	EmitColor dprec.Vec3

	// EmitDistance is the distance at which the light intensity reaches zero.
	EmitDistance float64

	// CastShadow specifies whether the light should cast shadows.
	CastShadow bool
}

// SpotLight represents a light source that emits light in a single, conical
// direction.
type SpotLight struct {

	// ID is the unique identifier of the light within the file.
	ID uint32

	// NodeID is the ID of the node that is associated with the light.
	NodeID uint32

	// EmitColor is the linear color of the light that is emitted.
	EmitColor dprec.Vec3

	// EmitDistance is the distance at which the light intensity reaches zero.
	EmitDistance float64

	// EmitAngleOuter is the angle at which the light intensity reaches zero.
	EmitAngleOuter dprec.Angle

	// EmitAngleInner is the angle at which the light intensity starts to
	// decline until it reaches the outer angle.
	EmitAngleInner dprec.Angle

	// CastShadow specifies whether the light should cast shadows.
	CastShadow bool
}

// DirectionalLight represents a light source that emits light in a single
// direction from infinitely away in space.
type DirectionalLight struct {

	// ID is the unique identifier of the light within the file.
	ID uint32

	// NodeID is the ID of the node that is associated with the light.
	NodeID uint32

	// EmitColor is the linear color of the light that is emitted.
	EmitColor dprec.Vec3

	// CastShadow specifies whether the light should cast shadows.
	CastShadow bool
}
