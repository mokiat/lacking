package asset

import "github.com/mokiat/gomath/dprec"

// TODO: Figure out how to handle more intricate animations
// that are not just on nodes.
// Maybe have a set of keyframe types and some path mechanism
// to target a node or a property on the content of a node.

// Animation represents a sequence of keyframes that can be
// applied to a scene to animate it.
type Animation struct {

	// Name identifies this animation.
	Name string

	// StartTime is the timestamp in seconds at which this animation starts.
	StartTime float64

	// EndTime is the timestamp in seconds at which this animation ends.
	EndTime float64

	// Bindings is a list of keyframes that are applied to the scene.
	Bindings []AnimationBinding
}

// AnimationBinding represents a set of keyframes that are applied
// to a specific node in the scene.
type AnimationBinding struct {

	// NodeName is the name of the node that this binding applies to.
	NodeName string

	// TranslationKeyframes is a list of keyframes that animate the translation
	// of the node.
	TranslationKeyframes []AnimationKeyframe[dprec.Vec3]

	// RotationKeyframes is a list of keyframes that animate the rotation
	// of the node.
	RotationKeyframes []AnimationKeyframe[dprec.Quat]

	// ScaleKeyframes is a list of keyframes that animate the scale
	// of the node.
	ScaleKeyframes []AnimationKeyframe[dprec.Vec3]
}

// AnimationKeyframe represents a single keyframe in an animation.
type AnimationKeyframe[T any] struct {

	// Timestamp is the time in seconds at which this keyframe is applied.
	Timestamp float64

	// Value is the value that is applied at the given timestamp.
	Value T
}
