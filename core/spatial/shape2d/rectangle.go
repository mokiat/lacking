package shape2d

import "github.com/mokiat/gomath/dprec"

// Rectangle represents a two-dimensional rectangle shape.
type Rectangle struct {
	// Center specifies the center point of the rectangle.
	Center dprec.Vec2
	// Rotation specifies the rotation of the rectangle.
	Rotation Rotation
	// HalfWidth specifies half the length of the rectangle's horizontal sides.
	HalfWidth float64
	// HalfHeight specifies half the length of the rectangle's vertical sides.
	HalfHeight float64
}

// TransformedRectangle returns a new [Rectangle] that is the result of applying
// the specified transform to the given rectangle. The center is moved by the
// transform and the rectangle's orientation is composed with the transform's
// rotation, while the half-width and half-height are left unchanged, since a
// rigid-body transform preserves distances.
func TransformedRectangle(rectangle Rectangle, transform Transform) Rectangle {
	return Rectangle{
		Center:     transform.Apply(rectangle.Center),
		Rotation:   ChainedRotation(transform.Rotation, rectangle.Rotation),
		HalfWidth:  rectangle.HalfWidth,
		HalfHeight: rectangle.HalfHeight,
	}
}

// ContainsPoint returns whether the specified point lies within the rectangle.
func (r Rectangle) ContainsPoint(point dprec.Vec2) bool {
	offset := dprec.Vec2Diff(point, r.Center)
	localPoint := r.Rotation.Inverse().Apply(offset)
	return localPoint.X >= -r.HalfWidth &&
		localPoint.X <= r.HalfWidth &&
		localPoint.Y >= -r.HalfHeight &&
		localPoint.Y <= r.HalfHeight
}

// BoundingCircle returns the smallest [Circle] that fully encompasses the rectangle.
func (r Rectangle) BoundingCircle() Circle {
	return Circle{
		Center: r.Center,
		Radius: dprec.Sqrt(r.HalfWidth*r.HalfWidth + r.HalfHeight*r.HalfHeight),
	}
}
