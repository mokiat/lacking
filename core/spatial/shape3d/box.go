package shape3d

import "github.com/mokiat/gomath/dprec"

// Box represents a three-dimensional, arbitrarily oriented box shape.
type Box struct {
	// Center specifies the center point of the box.
	Center dprec.Vec3
	// Rotation specifies the orientation of the box.
	Rotation Rotation
	// Width specifies the size of the box along its local X axis.
	Width float64
	// Height specifies the size of the box along its local Y axis.
	Height float64
	// Length specifies the size of the box along its local Z axis.
	Length float64
}

// TransformedBox returns a new Box that is the result of applying the specified
// transform to the given box. The center is moved by the transform and the box's
// orientation is composed with the transform's rotation, while the width, height
// and length are left unchanged, since a rigid-body transform preserves
// distances.
func TransformedBox(box Box, transform Transform) Box {
	return Box{
		Center:   transform.Apply(box.Center),
		Rotation: ChainedRotation(transform.Rotation, box.Rotation),
		Width:    box.Width,
		Height:   box.Height,
		Length:   box.Length,
	}
}

// ContainsPoint returns whether the specified point lies within the box.
func (b Box) ContainsPoint(point dprec.Vec3) bool {
	offset := dprec.Vec3Diff(point, b.Center)
	localPoint := b.Rotation.Inverse().Apply(offset)
	halfWidth := b.Width * 0.5
	halfHeight := b.Height * 0.5
	halfLength := b.Length * 0.5
	return localPoint.X >= -halfWidth &&
		localPoint.X <= halfWidth &&
		localPoint.Y >= -halfHeight &&
		localPoint.Y <= halfHeight &&
		localPoint.Z >= -halfLength &&
		localPoint.Z <= halfLength
}

// BoundingSphere returns the smallest Sphere that fully encompasses the box.
func (b Box) BoundingSphere() Sphere {
	halfWidth := b.Width * 0.5
	halfHeight := b.Height * 0.5
	halfLength := b.Length * 0.5
	return Sphere{
		Center: b.Center,
		Radius: dprec.Sqrt(halfWidth*halfWidth + halfHeight*halfHeight + halfLength*halfLength),
	}
}
