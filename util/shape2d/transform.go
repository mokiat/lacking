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
type Transform struct {

	// Translation specifies the translation that the transformation applies.
	Translation dprec.Vec2

	// Rotation specifies the rotation angle that the transformation applies.
	Rotation dprec.Angle
}

// Apply returns the transformation of the specified vector.
func (t *Transform) Apply(v dprec.Vec2) dprec.Vec2 {
	cs := dprec.Cos(t.Rotation)
	sn := dprec.Sin(t.Rotation)
	return dprec.Vec2Sum(t.Translation, dprec.Vec2{
		X: cs*v.X - sn*v.Y,
		Y: sn*v.X + cs*v.Y,
	})
}

// Basis returns a BasisTransform that is equivalent to the Transform.
func (t Transform) Basis() BasisTransform {
	return BasisTransform{
		Rotation:    AngleBasisRotation(t.Rotation),
		Translation: t.Translation,
	}
}

// BasisTransform represents a shape transformation with pre-computed basis
// vectors.
type BasisTransform struct {
	Translation dprec.Vec2
	Rotation    BasisRotation
}

// TRBasisTransform returns a new BasisTransform that represents both a translation
// and a rotation.
func TRBasisTransform(translation dprec.Vec2, angle dprec.Angle) BasisTransform {
	return BasisTransform{
		Translation: translation,
		Rotation:    AngleBasisRotation(angle),
	}
}

// Apply returns the transformation of the specified vector.
func (t BasisTransform) Apply(v dprec.Vec2) dprec.Vec2 {
	return dprec.Vec2Sum(t.Translation, t.Rotation.Apply(v))
}

// AngleBasisRotation returns a BasisRotation that represents the specified
// angle rotation.
func AngleBasisRotation(angle dprec.Angle) BasisRotation {
	cs := dprec.Cos(angle)
	sn := dprec.Sin(angle)
	return BasisRotation{
		BasisX: dprec.NewVec2(cs, sn),
		BasisY: dprec.NewVec2(-sn, cs),
	}
}

// BasisRotation represents a shape rotation with pre-computed basis vectors.
type BasisRotation struct {

	// BasisX holds the X basis vector of the rotation.
	BasisX dprec.Vec2

	// BasisY holds the Y basis vector of the rotation.
	BasisY dprec.Vec2
}

// Angle returns the angle of the rotation.
func (r BasisRotation) Angle() dprec.Angle {
	return dprec.Atan2(r.BasisX.Y, r.BasisX.X)
}

// Apply returns the rotation of the specified vector.
func (r BasisRotation) Apply(v dprec.Vec2) dprec.Vec2 {
	return dprec.Vec2{
		X: r.BasisX.X*v.X + r.BasisY.X*v.Y,
		Y: r.BasisX.Y*v.X + r.BasisY.Y*v.Y,
	}
}
