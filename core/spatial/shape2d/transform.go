package shape2d

import "github.com/mokiat/gomath/dprec"

// Transform represents a rigid-body transformation in 2D space, composed of a
// rotation followed by a translation. Applying it to a point first rotates the
// point and then offsets it by the translation.
type Transform struct {
	// Translation is the offset applied after rotation.
	Translation dprec.Vec2
	// Rotation is the orientation change applied before translation.
	Rotation Rotation
}

// IdentityTransform returns a transform that leaves points unchanged.
func IdentityTransform() Transform {
	return Transform{
		Translation: dprec.ZeroVec2(),
		Rotation:    IdentityRotation(),
	}
}

// TranslationTransform returns a transform that only offsets points by the
// specified translation and applies no rotation.
func TranslationTransform(translation dprec.Vec2) Transform {
	return Transform{
		Translation: translation,
		Rotation:    IdentityRotation(),
	}
}

// RotationTransform returns a transform that only rotates points by the
// specified rotation and applies no translation.
func RotationTransform(rotation Rotation) Transform {
	return Transform{
		Translation: dprec.ZeroVec2(),
		Rotation:    rotation,
	}
}

// TRTransform returns a transform that rotates points by the specified rotation
// and then offsets them by the specified translation.
func TRTransform(translation dprec.Vec2, rotation Rotation) Transform {
	return Transform{
		Translation: translation,
		Rotation:    rotation,
	}
}

// ChainedTransform returns the composition of a parent transform and a child
// transform, as found when resolving a child's transform relative to its
// parent. The child is applied first and the parent second, so that
// ChainedTransform(parent, child).Apply(p) equals parent.Apply(child.Apply(p)).
func ChainedTransform(parent, child Transform) Transform {
	return Transform{
		Translation: parent.Apply(child.Translation),
		Rotation:    ChainedRotation(parent.Rotation, child.Rotation),
	}
}

// Inverse returns the transform that undoes this transform. Applying a transform
// and then its inverse (in either order) leaves points unchanged.
func (t Transform) Inverse() Transform {
	invRotation := t.Rotation.Inverse()
	invTranslation := invRotation.Apply(dprec.InverseVec2(t.Translation))
	return Transform{
		Translation: invTranslation,
		Rotation:    invRotation,
	}
}

// Apply transforms the given point by rotating it and then offsetting it by the
// translation.
func (t Transform) Apply(point dprec.Vec2) dprec.Vec2 {
	rotated := t.Rotation.Apply(point)
	return dprec.Vec2Sum(rotated, t.Translation)
}
