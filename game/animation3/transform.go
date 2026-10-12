package animation3

import (
	"github.com/mokiat/gomath/dprec"
)

// Transform represents the local transformation of a single bone, relative
// to its parent.
type Transform struct {

	// Translation indicates the translation of the bone.
	Translation dprec.Vec3

	// Rotation indicates the rotation of the bone.
	Rotation dprec.Quat

	// Scale indicates the per-axis scale of the bone.
	Scale dprec.Vec3
}

// IdentityTransform returns a [Transform] with a zero translation, an identity
// rotation and a unit scale. It leaves points unchanged and is the additive
// neutral: adding it through [AddTransforms] leaves the other transform
// unchanged.
func IdentityTransform() Transform {
	return Transform{
		Translation: dprec.ZeroVec3(),
		Rotation:    dprec.IdentityQuat(),
		Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
	}
}

// BlendTransforms blends two transforms using the specified factor. A factor
// of 0.0 returns the first transform and a factor of 1.0 returns the second
// one. The factor is not clamped and values outside that range extrapolate.
//
// Translation and scale are blended linearly per axis. Rotation is blended
// through spherical linear interpolation along the shorter arc, which produces
// a constant angular velocity as the factor changes. Rotations are expected to
// be unit quaternions.
func BlendTransforms(first, second Transform, factor float64) Transform {
	return Transform{
		Translation: dprec.Vec3Lerp(first.Translation, second.Translation, factor),
		Rotation:    dprec.QuatSlerp(first.Rotation, second.Rotation, factor),
		Scale:       dprec.Vec3Lerp(first.Scale, second.Scale, factor),
	}
}

// AddTransforms combines two transforms by applying the second one as an
// additive layer on top of the first one.
//
// Translation is added per axis and scale is multiplied per axis. Rotation is
// composed by applying the rotation of the second transform on top of the
// rotation of the first. Because rotation composition is not commutative, the
// order of the arguments matters. Rotations are expected to be unit
// quaternions.
//
// Adding the [IdentityTransform] on either side leaves the other transform
// unchanged.
func AddTransforms(first, second Transform) Transform {
	return Transform{
		Translation: dprec.Vec3Sum(first.Translation, second.Translation),
		Rotation:    dprec.QuatProd(second.Rotation, first.Rotation),
		Scale: dprec.NewVec3(
			first.Scale.X*second.Scale.X,
			first.Scale.Y*second.Scale.Y,
			first.Scale.Z*second.Scale.Z,
		),
	}
}

// DiffTransforms returns the additive difference of the first transform
// relative to the second one. It is the inverse of [AddTransforms], in the
// sense that adding the result on top of the second transform reproduces
// the first one:
//
//	AddTransforms(second, DiffTransforms(first, second)) == first
//
// Translation is subtracted per axis and scale is divided per axis. Rotation
// is the rotation that needs to be applied on top of the rotation of the
// second transform in order to reach the rotation of the first. Rotations are
// expected to be unit quaternions. Scale components of the second transform
// should not be zero, otherwise the resulting scale is not finite.
//
// Diffing a transform against itself yields the [IdentityTransform].
func DiffTransforms(first, second Transform) Transform {
	return Transform{
		Translation: dprec.Vec3Diff(first.Translation, second.Translation),
		Rotation:    dprec.QuatProd(first.Rotation, dprec.InverseQuat(second.Rotation)),
		Scale: dprec.NewVec3(
			first.Scale.X/second.Scale.X,
			first.Scale.Y/second.Scale.Y,
			first.Scale.Z/second.Scale.Z,
		),
	}
}

// Matrix returns the affine transformation matrix that corresponds to this
// [Transform]. When applied to a point, the matrix first scales it, then
// rotates it and finally translates it.
func (t Transform) Matrix() dprec.Mat4 {
	return dprec.TRSMat4(
		t.Translation,
		t.Rotation,
		t.Scale,
	)
}
