package shape2d

import "github.com/mokiat/gomath/dprec"

// IdentityTransform returns a new Transform that represents the origin.
func IdentityTransform() Transform {
	return Transform{
		Translation: dprec.ZeroVec2(),
		Rotation:    dprec.Radians(0.0),
	}
}

// TranslationTransform returns a new Transform that represents a translation.
func TranslationTransform(translation dprec.Vec2) Transform {
	return Transform{
		Translation: translation,
		Rotation:    dprec.Radians(0.0),
	}
}

// RotationTransform returns a new Transform that represents a rotation.
func RotationTransform(rotation dprec.Angle) Transform {
	return Transform{
		Translation: dprec.ZeroVec2(),
		Rotation:    rotation,
	}
}

// TRTransform returns a new Transform that represents both a translation
// and a rotation.
func TRTransform(translation dprec.Vec2, rotation dprec.Angle) Transform {
	return Transform{
		Translation: translation,
		Rotation:    rotation,
	}
}

// ChainedTransform returns the Transform that is the result of combining
// two Transforms together.
func ChainedTransform(parent, child Transform) Transform {
	return Transform{
		Translation: parent.Apply(child.Translation),
		Rotation:    parent.Rotation + child.Rotation,
	}
}

// Transform represents a shape transformation.
//
// Note: This package assumes that Y points down (i.e. that XY in 2D correspond
// to XZ in 3D). Rotation is still counter-clockwise.
type Transform struct {

	// Translation specifies the translation that the transformation applies.
	Translation dprec.Vec2

	// Rotation specifies the rotation angle that the transformation applies.
	Rotation dprec.Angle
}

// Apply returns the transformation of the specified vector.
func (t *Transform) Apply(v dprec.Vec2) dprec.Vec2 {
	cs := dprec.Cos(-t.Rotation)
	sn := dprec.Sin(-t.Rotation)
	return dprec.Vec2Sum(t.Translation, dprec.Vec2{
		X: cs*v.X - sn*v.Y,
		Y: sn*v.X + cs*v.Y,
	})
}
