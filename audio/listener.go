package audio

import "github.com/mokiat/gomath/sprec"

// SpatialListener represents a listener in 3D space for spatial audio.
//
// The listener's position and orientation can be used to create spatial audio
// effects, such as panning and distance attenuation, for [SpatialNode] sources.
//
// Distance attenuation is applied to [SpatialNode] sources based on the
// distance between the source and the listener. The attenuation model is
// "inverse", where the gain is calculated as 1.0 / max(1.0, distance).
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
