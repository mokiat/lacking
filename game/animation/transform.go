package animation

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// NodeTransform represents the transformation of a node.
type NodeTransform struct {

	// Translation, if specified, indicates the translation of the node.
	Translation opt.T[dprec.Vec3]

	// Rotation, if specified, indicates the rotation of the node.
	Rotation opt.T[dprec.Quat]

	// Scale, if specified, indicates the scale of the node.
	Scale opt.T[dprec.Vec3]
}

// InverseNodeTransform returns the inverse of a node transform.
func InverseNodeTransform(transform NodeTransform) NodeTransform {
	var result NodeTransform
	if transform.Translation.Specified {
		result.Translation = opt.V(dprec.InverseVec3(transform.Translation.Value))
	}
	if transform.Rotation.Specified {
		result.Rotation = opt.V(dprec.InverseQuat(transform.Rotation.Value))
	}
	if transform.Scale.Specified {
		scale := transform.Scale.Value
		result.Scale = opt.V(dprec.NewVec3(1.0/scale.X, 1.0/scale.Y, 1.0/scale.Z))
	}
	return result
}

// BlendNodeTransforms blends two node transformations using the specified
// factor. A factor of 0.0 means that the first transformation is used, a
// factor of 1.0 means that the second transformation is used.
func BlendNodeTransforms(first, second NodeTransform, factor float64) NodeTransform {
	return NodeTransform{
		Translation: combineLinear(first.Translation, second.Translation, factor),
		Rotation:    combineSpherical(first.Rotation, second.Rotation, factor),
		Scale:       combineLinear(first.Scale, second.Scale, factor),
	}
}

// AddNodeTransforms combines two transforms into a single one.
func AddNodeTransforms(first, second NodeTransform) NodeTransform {
	return NodeTransform{
		Translation: addLinear(first.Translation, second.Translation),
		Rotation:    addSpherical(first.Rotation, second.Rotation),
		Scale:       addLinear(first.Scale, second.Scale),
	}
}

// DiffNodeTransforms combines two transforms into a single one.
func DiffNodeTransforms(first, second NodeTransform) NodeTransform {
	return NodeTransform{
		Translation: diffLinear(first.Translation, second.Translation),
		Rotation:    diffSpherical(first.Rotation, second.Rotation),
		Scale:       diffLinear(first.Scale, second.Scale),
	}
}

func combineLinear(first, second opt.T[dprec.Vec3], amount float64) opt.T[dprec.Vec3] {
	switch {
	case first.Specified && second.Specified:
		return opt.V(dprec.Vec3Lerp(first.Value, second.Value, amount))
	case first.Specified:
		return first
	case second.Specified:
		return second
	default:
		return opt.Unspecified[dprec.Vec3]()
	}
}

func combineSpherical(first, second opt.T[dprec.Quat], amount float64) opt.T[dprec.Quat] {
	switch {
	case first.Specified && second.Specified:
		return opt.V(dprec.QuatSlerp(first.Value, second.Value, amount))
	case first.Specified:
		return first
	case second.Specified:
		return second
	default:
		return opt.Unspecified[dprec.Quat]()
	}
}

func addLinear(first, second opt.T[dprec.Vec3]) opt.T[dprec.Vec3] {
	switch {
	case first.Specified && second.Specified:
		return opt.V(dprec.Vec3Sum(first.Value, second.Value))
	case first.Specified:
		return opt.V(first.Value)
	case second.Specified:
		return opt.V(second.Value)
	default:
		return opt.Unspecified[dprec.Vec3]()
	}
}

func addSpherical(first, second opt.T[dprec.Quat]) opt.T[dprec.Quat] {
	switch {
	case first.Specified && second.Specified:
		return opt.V(dprec.QuatProd(second.Value, first.Value))
	case first.Specified:
		return opt.V(first.Value)
	case second.Specified:
		return opt.V(second.Value)
	default:
		return opt.Unspecified[dprec.Quat]()
	}
}

func diffLinear(first, second opt.T[dprec.Vec3]) opt.T[dprec.Vec3] {
	switch {
	case first.Specified && second.Specified:
		return opt.V(dprec.Vec3Diff(first.Value, second.Value))
	default:
		return opt.Unspecified[dprec.Vec3]()
	}
}

func diffSpherical(first, second opt.T[dprec.Quat]) opt.T[dprec.Quat] {
	switch {
	case first.Specified && second.Specified:
		return opt.V(dprec.QuatDiff(first.Value, second.Value, true))
	default:
		return opt.Unspecified[dprec.Quat]()
	}
}
