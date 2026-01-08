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
