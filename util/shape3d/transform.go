package shape3d

import "github.com/mokiat/gomath/dprec"

// Transform represents a shape transformation.
type Transform struct {
	Translation dprec.Vec3
	Rotation    dprec.Quat
}

// Apply returns the transformation of the specified vector.
func (t *Transform) Apply(v dprec.Vec3) dprec.Vec3 {
	return dprec.Vec3Sum(t.Translation, dprec.QuatVec3Rotation(t.Rotation, v))
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
