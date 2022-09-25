package shape

import "github.com/mokiat/gomath/dprec"

// NewStaticBox creates a new StaticBox shape.
func NewStaticBox(width, height, length float64) StaticBox {
	size := dprec.NewVec3(width, height, length)
	return StaticBox{
		size:     size,
		bsRadius: size.Length() / 2.0,
	}
}

// StaticBox represents a box shape that cannot be resized.
type StaticBox struct {
	size     dprec.Vec3
	bsRadius float64
}

// BoundingSphereRadius returns the radius of a sphere that can encompass
// this shape.
func (b StaticBox) BoundingSphereRadius() float64 {
	return b.bsRadius
}

// Width returns the width of this box.
func (b StaticBox) Width() float64 {
	return b.size.X
}

// Height returns the height of this box.
func (b StaticBox) Height() float64 {
	return b.size.Y
}

// Length returns the length of this box.
func (b StaticBox) Length() float64 {
	return b.size.Z
}
