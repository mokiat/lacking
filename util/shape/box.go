package shape

import "github.com/mokiat/gomath/dprec"

// NewStaticBox creates a new StaticBox shape.
func NewStaticBox(width, height, length float64) StaticBox {
	size := dprec.NewVec3(width, height, length)
	return StaticBox{
		Transform: IdentityTransform(),
		size:      size,
		bsRadius:  size.Length() / 2.0,
	}
}

// StaticBox represents an immutable box shape.
type StaticBox struct {
	Transform
	size     dprec.Vec3
	bsRadius float64
}

// BoundingSphereRadius returns the radius of a sphere that can encompass
// this shape.
func (b StaticBox) BoundingSphereRadius() float64 {
	return b.bsRadius
}

// Width returns the width of this StaticBox.
func (b StaticBox) Width() float64 {
	return b.size.X
}

// HalfWidth returns half of the width of this StaticBox.
func (b StaticBox) HalfWidth() float64 {
	return b.size.X / 2.0
}

// WithWidth returns a new StaticBox that is based on this one but has the
// specified width.
func (b StaticBox) WithWidth(width float64) StaticBox {
	b.size.X = width
	b.bsRadius = b.size.Length() / 2.0
	return b
}

// Height returns the height of this StaticBox.
func (b StaticBox) Height() float64 {
	return b.size.Y
}

// HalfHeight returns half of the height of this StaticBox.
func (b StaticBox) HalfHeight() float64 {
	return b.size.Y / 2.0
}

// WithHeight returns a new StaticBox that is based on this one but has the
// specified height.
func (b StaticBox) WithHeight(height float64) StaticBox {
	b.size.Y = height
	b.bsRadius = b.size.Length() / 2.0
	return b
}

// Length returns the length of this StaticBox.
func (b StaticBox) Length() float64 {
	return b.size.Z
}

// HalfLength returns half of the length of this StaticBox.
func (b StaticBox) HalfLength() float64 {
	return b.size.Z / 2.0
}

// WithLength returns a new StaticBox that is based on this one but has the
// specified length.
func (b StaticBox) WithLength(length float64) StaticBox {
	b.size.Z = length
	b.bsRadius = b.size.Length() / 2.0
	return b
}

// WithTransform returns a new StaticBox that is based on this one but has
// the specified transform.
func (b StaticBox) WithTransform(transform Transform) StaticBox {
	b.Transform = transform
	return b
}

// Transformed returns a new StaticBox that is based on this one but has
// the specified transform applied to it.
func (b StaticBox) Transformed(parent Transform) StaticBox {
	b.Transform = b.Transform.Transformed(parent)
	return b
}
