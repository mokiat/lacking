package asset

import "github.com/mokiat/gomath/dprec"

// AmbientLight represents a light source that emits light in all directions
// from all points in space.
type AmbientLight struct {

	// NodeIndex is the index of the node that is associated with the light.
	NodeIndex uint32

	// ReflectionTextureIndex is the index of the cube texture that is used
	// for reflection mapping.
	ReflectionTextureIndex uint32

	// RefractionTextureIndex is the index of the cube texture that is used
	// for refraction mapping.
	RefractionTextureIndex uint32

	// CastShadow specifies whether a SSAO-type technique should be applied.
	CastShadow bool
}

// PointLight represents a light source that emits light in all directions
// from a single point in space.
type PointLight struct {

	// NodeIndex is the index of the node that is associated with the light.
	NodeIndex uint32

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

	// NodeIndex is the index of the node that is associated with the light.
	NodeIndex uint32

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

	// NodeIndex is the index of the node that is associated with the light.
	NodeIndex uint32

	// EmitColor is the linear color of the light that is emitted.
	EmitColor dprec.Vec3

	// EmitDistance is the distance at which the light intensity reaches zero.
	EmitDistance float64

	// CastShadow specifies whether the light should cast shadows.
	CastShadow bool
}
