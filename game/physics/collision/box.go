package collision

import "github.com/mokiat/gomath/dprec"

// NewBox creates a new Box shape.
func NewBox(position dprec.Vec3, rotation dprec.Quat, size dprec.Vec3) Box {
	return Box{
		transform: TRTransform(position, rotation),
		size:      size,
	}
}

// Box represents a 3D box shape.
type Box struct {
	transform Transform
	size      dprec.Vec3
}

// Replace replaces this shape with the template one after the specified
// transformation has been applied to it.
func (b *Box) Replace(template Box, transform Transform) {
	b.transform = ChainedTransform(transform, template.transform)
	b.size = template.size
}

// Position returns the position of this box.
func (b *Box) Position() dprec.Vec3 {
	return b.transform.Translation
}

// Rotation returns the rotation of this box.
func (b *Box) Rotation() dprec.Quat {
	return b.transform.Rotation
}

// Size returns the size of this box.
func (b *Box) Size() dprec.Vec3 {
	return b.size
}

// Width returns the width of this box.
func (b *Box) Width() float64 {
	return b.size.X
}

// Height returns the height of this box.
func (b *Box) Height() float64 {
	return b.size.Y
}

// Length returns the length of this box.
func (b *Box) Length() float64 {
	return b.size.Z
}

// HalfWidth returns half of the width of this box.
func (b *Box) HalfWidth() float64 {
	return b.size.X / 2.0
}

// HalfHeight returns half of the height of this box.
func (b *Box) HalfHeight() float64 {
	return b.size.Y / 2.0
}

// HalfLength returns half of the length of this box.
func (b *Box) HalfLength() float64 {
	return b.size.Z / 2.0
}

// BoundingSphere returns a sphere that encompases this box.
func (b *Box) BoundingSphere() Sphere {
	return NewSphere(b.transform.Translation, b.size.Length()/2.0)
}
