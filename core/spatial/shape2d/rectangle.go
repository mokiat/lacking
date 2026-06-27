package shape2d

import "github.com/mokiat/gomath/dprec"

// Rectangle represents a two-dimensional rectangle shape.
type Rectangle struct {
	// Center specifies the center point of the rectangle.
	Center dprec.Vec2
	// Rotation specifies the rotation of the rectangle.
	Rotation Rotation
	// Width specifies the length of the rectangle's horizontal sides.
	Width float64
	// Height specifies the length of the rectangle's vertical sides.
	Height float64
}

// TransformedRectangle returns a new Rectangle that is the result of applying
// the specified transform to the given rectangle. The center is moved by the
// transform and the rectangle's orientation is composed with the transform's
// rotation, while the width and height are left unchanged, since a rigid-body
// transform preserves distances.
func TransformedRectangle(rectangle Rectangle, transform Transform) Rectangle {
	return Rectangle{
		Center:   transform.Apply(rectangle.Center),
		Rotation: ChainedRotation(transform.Rotation, rectangle.Rotation),
		Width:    rectangle.Width,
		Height:   rectangle.Height,
	}
}

// ContainsPoint returns whether the specified point lies within the rectangle.
func (r Rectangle) ContainsPoint(point dprec.Vec2) bool {
	offset := dprec.Vec2Diff(point, r.Center)
	localPoint := r.Rotation.Inverse().Apply(offset)
	halfWidth := r.Width * 0.5
	halfHeight := r.Height * 0.5
	return localPoint.X >= -halfWidth &&
		localPoint.X <= halfWidth &&
		localPoint.Y >= -halfHeight &&
		localPoint.Y <= halfHeight
}

// BoundingCircle returns the smallest Circle that fully encompasses the rectangle.
func (r Rectangle) BoundingCircle() Circle {
	halfWidth := r.Width * 0.5
	halfHeight := r.Height * 0.5
	return Circle{
		Center: r.Center,
		Radius: dprec.Sqrt(halfWidth*halfWidth + halfHeight*halfHeight),
	}
}
