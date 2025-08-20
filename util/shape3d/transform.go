package shape3d

import "github.com/mokiat/gomath/dprec"

// IdentityTransform returns a new Transform that represents the origin.
func IdentityTransform() Transform {
	return Transform{
		Translation: dprec.ZeroVec3(),
		Rotation:    dprec.IdentityQuat(),
	}
}

// TranslationTransform returns a new Transform that represents a translation.
func TranslationTransform(translation dprec.Vec3) Transform {
	return Transform{
		Translation: translation,
		Rotation:    dprec.IdentityQuat(),
	}
}

// RotationTransform returns a new Transform that represents a rotation.
func RotationTransform(rotation dprec.Quat) Transform {
	return Transform{
		Translation: dprec.ZeroVec3(),
		Rotation:    rotation,
	}
}

// TRTransform returns a new Transform that represents both a translation
// and a rotation.
func TRTransform(translation dprec.Vec3, rotation dprec.Quat) Transform {
	return Transform{
		Translation: translation,
		Rotation:    rotation,
	}
}

// ChainedTransform returns the Transform that is the result of combining
// two Transforms together.
func ChainedTransform(parent, child Transform) Transform {
	return Transform{
		Translation: dprec.Vec3Sum(
			parent.Translation,
			dprec.QuatVec3Rotation(parent.Rotation, child.Translation),
		),
		Rotation: dprec.QuatProd(parent.Rotation, child.Rotation),
	}
}

// Transform represents a shape transformation.
type Transform struct {

	// Translation specifies the translation that the transformation applies.
	Translation dprec.Vec3

	// Rotation specifies the rotation that the transformation applies.
	Rotation dprec.Quat
}

// Apply returns the transformation of the specified vector.
func (t *Transform) Apply(v dprec.Vec3) dprec.Vec3 {
	return dprec.Vec3Sum(t.Translation, dprec.QuatVec3Rotation(t.Rotation, v))
}
