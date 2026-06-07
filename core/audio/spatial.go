package audio

import "github.com/mokiat/gomath/sprec"

// SpatialListener represents a listener in 3D space for spatial audio.
type SpatialListener interface {

	// Position returns the 3D position of the listener.
	Position() sprec.Vec3

	// SetPosition sets the 3D position of the listener.
	SetPosition(position sprec.Vec3)

	// Rotation returns the orientation of the listener as a quaternion.
	Rotation() sprec.Quat

	// SetRotation sets the orientation of the listener as a quaternion.
	SetRotation(rotation sprec.Quat)
}

// SpatialEmitter represents an emitter in 3D space for spatial audio.
type SpatialEmitter interface {

	// Position returns the 3D position of the emitter.
	Position() sprec.Vec3

	// SetPosition sets the 3D position of the emitter.
	SetPosition(position sprec.Vec3)

	// Rotation returns the orientation of the emitter as a quaternion.
	Rotation() sprec.Quat

	// SetRotation sets the orientation of the emitter as a quaternion.
	SetRotation(rotation sprec.Quat)

	// InnerConeAngle returns the inner cone angle of the emitter.
	//
	// Within this cone the emitter plays at full gain. Between the inner and
	// outer cone the gain is linearly interpolated toward [OuterConeGain].
	InnerConeAngle() sprec.Angle

	// SetInnerConeAngle sets the inner cone angle of the emitter.
	SetInnerConeAngle(angle sprec.Angle)

	// OuterConeAngle returns the outer cone angle of the emitter.
	//
	// Default is 360 degrees, which means no directional attenuation.
	OuterConeAngle() sprec.Angle

	// SetOuterConeAngle sets the outer cone angle of the emitter.
	SetOuterConeAngle(angle sprec.Angle)

	// OuterConeGain returns the gain applied to the emitter when the listener is outside the outer cone.
	//
	// Default is 0.0, which means the emitter is silent when the listener is outside the outer cone.
	OuterConeGain() float32

	// SetOuterConeGain sets the gain applied to the emitter when the listener is outside the outer cone.
	SetOuterConeGain(gain float32)
}
