package shape

import "github.com/mokiat/gomath/dprec"

// Point represents a single point in 3D space.
type Point dprec.Vec3

// Transformed returns a new Point that is based on this one but with the
// specified Transform applied to it.
func (p Point) Transformed(parent Transform) Point {
	return Point(dprec.Vec3Sum(
		parent.position,
		dprec.QuatVec3Rotation(parent.rotation, dprec.Vec3(p)),
	))
}
