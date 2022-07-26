package shape

import "github.com/mokiat/gomath/dprec"

// NewPlacement creates a new Placement for the specified shape with the
// specified position and rotation.
func NewPlacement(shape Shape, position dprec.Vec3, rotation dprec.Quat) Placement {
	return Placement{
		shape:    shape,
		position: position,
		rotation: rotation,
	}
}

// NewIdentityPlacement returns a new Placement that is centered.
func NewIdentityPlacement(shape Shape) Placement {
	return NewPlacement(shape, dprec.ZeroVec3(), dprec.IdentityQuat())
}

// Placement represents a mechanism through which a shape can have dynamic
// position and orientation assigned.
type Placement struct {
	shape    Shape
	position dprec.Vec3
	rotation dprec.Quat
}

// BoundingSphereRadius returns the radius of a sphere that can encompass
// this shape.
func (p Placement) BoundingSphereRadius() float64 {
	return p.shape.BoundingSphereRadius() + p.position.Length()
}

// Shape returns the Shape held by this Placement.
func (p Placement) Shape() Shape {
	return p.shape
}

// Position returns this Placement's position.
func (p Placement) Position() dprec.Vec3 {
	return p.position
}

// Rotation returns this Placement's rotation.
func (p Placement) Rotation() dprec.Quat {
	return p.rotation
}

// Transformed returns a new Placement that is based on the current one but with
// the specified translation and rotation applied on top of this one.
func (p Placement) Transformed(translation dprec.Vec3, rotation dprec.Quat) Placement {
	p.position = dprec.Vec3Sum(translation, dprec.QuatVec3Rotation(rotation, p.position))
	p.rotation = dprec.QuatProd(rotation, p.rotation)
	return p
}
