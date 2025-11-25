package shape3d

import "github.com/mokiat/gomath/dprec"

// NewBox creates a new box with the specified position, rotation and size.
//
// The size is the full size of the box. Internally it will be
// converted to half sizes.
func NewBox(position dprec.Vec3, rotation dprec.Quat, size dprec.Vec3) Box {
	return Box{
		Position:   position,
		Rotation:   rotation,
		HalfWidth:  size.X / 2.0,
		HalfHeight: size.Y / 2.0,
		HalfLength: size.Z / 2.0,
	}
}

// TransformedBox creates a new box from the specified source box by applying
// the specified transformation.
func TransformedBox(source Box, transform Transform) Box {
	boxTransform := ChainedTransform(transform, Transform{
		Translation: source.Position,
		Rotation:    source.Rotation,
	})
	return Box{
		Position:   boxTransform.Translation,
		Rotation:   boxTransform.Rotation,
		HalfWidth:  source.HalfWidth,
		HalfHeight: source.HalfHeight,
		HalfLength: source.HalfLength,
	}
}

// Box represents a cuboid shape.
type Box struct {

	// Position holds the position of the box.
	Position dprec.Vec3

	// Rotation holds the rotation of the box.
	Rotation dprec.Quat

	// HalfWidth holds the half-width of the bx.
	HalfWidth float64

	// HalfHeight holds the half-height of the box.
	HalfHeight float64

	// HalfLength holds the half-length of the box.
	HalfLength float64
}

// BoundingSphere returns the bounding sphere of the box.
func (b *Box) BoundingSphere() Sphere {
	return Sphere{
		Position: b.Position,
		Radius: dprec.Sqrt(
			b.HalfWidth*b.HalfWidth + b.HalfHeight*b.HalfHeight + b.HalfLength*b.HalfLength,
		),
	}
}
