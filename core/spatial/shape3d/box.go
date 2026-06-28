package shape3d

import "github.com/mokiat/gomath/dprec"

// Box represents a three-dimensional, arbitrarily oriented box shape.
type Box struct {
	// Center specifies the center point of the box.
	Center dprec.Vec3
	// Rotation specifies the orientation of the box.
	Rotation Rotation
	// HalfWidth specifies half the size of the box along its local X axis.
	HalfWidth float64
	// HalfHeight specifies half the size of the box along its local Y axis.
	HalfHeight float64
	// HalfLength specifies half the size of the box along its local Z axis.
	HalfLength float64
}

// TransformedBox returns a new [Box] that is the result of applying the specified
// transform to the given box. The center is moved by the transform and the box's
// orientation is composed with the transform's rotation, while the half-width,
// half-height and half-length are left unchanged, since a rigid-body transform
// preserves distances.
func TransformedBox(box Box, transform Transform) Box {
	return Box{
		Center:     transform.Apply(box.Center),
		Rotation:   ChainedRotation(transform.Rotation, box.Rotation),
		HalfWidth:  box.HalfWidth,
		HalfHeight: box.HalfHeight,
		HalfLength: box.HalfLength,
	}
}

// ContainsPoint returns whether the specified point lies within the box.
func (b Box) ContainsPoint(point dprec.Vec3) bool {
	offset := dprec.Vec3Diff(point, b.Center)
	localPoint := b.Rotation.Inverse().Apply(offset)
	return localPoint.X >= -b.HalfWidth &&
		localPoint.X <= b.HalfWidth &&
		localPoint.Y >= -b.HalfHeight &&
		localPoint.Y <= b.HalfHeight &&
		localPoint.Z >= -b.HalfLength &&
		localPoint.Z <= b.HalfLength
}

// BoundingSphere returns the smallest [Sphere] that fully encompasses the box.
func (b Box) BoundingSphere() Sphere {
	return Sphere{
		Center: b.Center,
		Radius: dprec.Sqrt(b.HalfWidth*b.HalfWidth + b.HalfHeight*b.HalfHeight + b.HalfLength*b.HalfLength),
	}
}
