package shape

import "github.com/mokiat/gomath/dprec"

// Note: Having to cast Point to Vec3 and vice versa does not affect
// performance in any way.

// Point represents a single point in 3D space.
type Point dprec.Vec3

// Transformed returns a new Point that is based on this one but with the
// specified Transform applied to it.
func (p Point) Transformed(parent Transform) Point {
	// Note: Doing an identity check on the transform,
	// as a form of quick return, actually worsens the performance.
	return Point(dprec.Vec3Sum(
		parent.position,
		dprec.QuatVec3Rotation(parent.rotation, dprec.Vec3(p)),
	))
}
