package shape2d

import "github.com/mokiat/gomath/dprec"

// NewRectangle returns a rectangle from the specified position, rotation and
// size.
//
// The size parameter specifies the full width and height of the rectangle.
// Internally it will be converted to half sizes.
func NewRectangle(pos dprec.Vec2, rotation dprec.Angle, size dprec.Vec2) Rectangle {
	return Rectangle{
		Position:   pos,
		Rotation:   rotation,
		HalfWidth:  size.X / 2.0,
		HalfHeight: size.Y / 2.0,
	}
}

// TransformedRectangle creates a new rectangle from the specified source
// rectangle by applying the specified transformation.
func TransformedRectangle(source Rectangle, transform Transform) Rectangle {
	rectTransform := ChainedTransform(transform, Transform{
		Translation: source.Position,
		Rotation:    source.Rotation,
	})
	return Rectangle{
		Position:   rectTransform.Translation,
		Rotation:   rectTransform.Rotation,
		HalfWidth:  source.HalfWidth,
		HalfHeight: source.HalfHeight,
	}
}

// Rectnagle represents a 2D rectangle.
type Rectangle struct {

	// Position holds the position of the rectangle.
	Position dprec.Vec2

	// Rotation holds the rotation of the rectangle.
	Rotation dprec.Angle

	// HalfWidth holds the half-width of the rectangle.
	HalfWidth float64

	// HalfHeight holds the half-height of the rectangle.
	HalfHeight float64
}

// BoundingCircle returns the bounding circle of the rectangle.
func (r *Rectangle) BoundingCircle() Circle {
	return Circle{
		Position: r.Position,
		Radius:   dprec.Sqrt(r.HalfWidth*r.HalfWidth + r.HalfHeight*r.HalfHeight),
	}
}
