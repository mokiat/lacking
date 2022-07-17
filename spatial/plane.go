package spatial

import "github.com/mokiat/gomath/sprec"

// Plane represents a plane that can be used to split a 3D space
// into two sub-spaces. It is represented through the mathematical formula
//
// 		a*x + b*y + c*z + d = 0
//
// A point (x, y, z) that produces zero in the left part of the equation is
// considered to lay on the plane. A positive value indicates that the point
// is inside the sub-region defined by the plane. A negative value indicates
// that the point is outside the sub-region.
type Plane sprec.Vec4

// ContainsPoint checks whether a point with the specified position
// is inside the sub-region of this Plane or at least partailly inside.
//
// The Plane need not be normalized.
func (p Plane) ContainsPoint(position sprec.Vec3) bool {
	return p.X*position.X+p.Y*position.Y+p.Z*position.Z+p.W >= 0.0
}

// ContainsSphere checks whether a sphere with the specified position and radius
// is inside the sub-region of this Plane or at least partailly inside.
//
// For this method to work, the Plane must be normalized.
func (p Plane) ContainsSphere(position sprec.Vec3, radius float32) bool {
	return p.X*position.X+p.Y*position.Y+p.Z*position.Z+p.W >= -radius
}

// Normalized returns a normalized Plane, which is a plane that has
// a normal component (a, b, c) that is of unit length.
//
// Normalized clipping planes produce correct distances to the plane when
// multiplied (dot product) by a 4D positional vector.
func (p Plane) Normalized() Plane {
	normal := sprec.NewVec3(p.X, p.Y, p.Z)
	distance := p.W
	correction := 1.0 / normal.Length()
	return Plane(sprec.NewVec4(
		normal.X*correction,
		normal.Y*correction,
		normal.Z*correction,
		distance*correction,
	))
}
