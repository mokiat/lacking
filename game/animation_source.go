package game

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

// AnimationSource represents a source of animation data.
type AnimationSource interface {

	// Length returns the length of the animation in seconds.
	Length() float64

	// Position returns the current position of the animation in seconds.
	Position() float64

	// SetPosition sets the current position of the animation in seconds.
	SetPosition(position float64)

	// NodeTransform returns the transformation of the node with the
	// specified name at the current time position.
	NodeTransform(name string) NodeTransform
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
