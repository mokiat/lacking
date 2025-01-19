package game

import (
	"maps"
	"slices"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// AnimationDefinitionInfo contains the information required to define an
// animation.
type AnimationDefinitionInfo struct {

	// Name is the name of the animation.
	Name string

	// StartTime is the time (in seconds) at which the animation starts.
	StartTime float64

	// EndTime is the time (in seconds) at which the animation ends.
	EndTime float64

	// Loop specifies whether the animation should loop.
	Loop bool

	// Bindings is a list of node bindings that are affected by the animation.
	Bindings []AnimationBindingDefinitionInfo
}

// AnimationBindingDefinitionInfo contains the information required to define
// an animation node binding.
type AnimationBindingDefinitionInfo struct {

	// NodeName is the name of the node that is affected by the animation.
	NodeName string

	// TranslationKeyframes is a list of keyframes that define the translation
	// of the node.
	TranslationKeyframes KeyframeList[dprec.Vec3]

	// RotationKeyframes is a list of keyframes that define the rotation of the
	// node.
	RotationKeyframes KeyframeList[dprec.Quat]

	// ScaleKeyframes is a list of keyframes that define the scale of the node.
	ScaleKeyframes KeyframeList[dprec.Vec3]
}

// AnimationDefinition represents a definition of an animation.
type AnimationDefinition struct {
	name      string
	startTime float64
	endTime   float64
	loop      bool
	bindings  map[string]animationBinding
}

// Name returns the name of the animation.
func (d *AnimationDefinition) Name() string {
	return d.name
}

// StartTime returns the time (in seconds) at which the animation starts.
func (d *AnimationDefinition) StartTime() float64 {
	return d.startTime
}

// EndTime returns the time (in seconds) at which the animation ends.
func (d *AnimationDefinition) EndTime() float64 {
	return d.endTime
}

// Loop returns whether the animation should loop.
func (d *AnimationDefinition) Loop() bool {
	return d.loop
}

// NodeNames returns the names of the nodes that are animated by the animation.
func (d *AnimationDefinition) NodeNames() []string {
	return slices.Collect(maps.Keys(d.bindings))
}

// AnimationInfo represents an instantiation of an animation instance.
type AnimationInfo struct {

	// Definition is the definition of the animation.
	Definition *AnimationDefinition

	// ClipStart, if specified, overrides the start time of the animation.
	ClipStart opt.T[float64]

	// ClipEnd, if specified, overrides the end time of the animation.
	ClipEnd opt.T[float64]

	// Loop, if specified, overrides the loop setting of the animation.
	Loop opt.T[bool]
}

// Animation represents an instantiation of a keyframe animation.
type Animation struct {
	name      string
	startTime float64
	endTime   float64
	loop      bool
	bindings  map[string]animationBinding
}

// Name returns the name of the animation.
func (a *Animation) Name() string {
	return a.name
}

// StartTime returns the time (in seconds) at which the animation starts.
func (a *Animation) StartTime() float64 {
	return a.startTime
}

// EndTime returns the time (in seconds) at which the animation ends.
func (a *Animation) EndTime() float64 {
	return a.endTime
}

// Loop returns whether the animation should loop.
func (a *Animation) Loop() bool {
	return a.loop
}

// Length returns the length of the animation in seconds.
func (a *Animation) Length() float64 {
	return a.EndTime() - a.StartTime()
}

// BindingTransform returns the transformation of the node with the specified
// name at the specified time position.
func (a *Animation) BindingTransform(name string, timestamp float64) NodeTransform {
	var result NodeTransform
	binding, ok := a.bindings[name]
	if !ok {
		return result
	}
	if len(binding.translationKeyframes) > 0 {
		result.Translation = opt.V(binding.Translation(timestamp))
	}
	if len(binding.rotationKeyframes) > 0 {
		result.Rotation = opt.V(binding.Rotation(timestamp))
	}
	if len(binding.scaleKeyframes) > 0 {
		result.Scale = opt.V(binding.Scale(timestamp))
	}
	return result
}

// Playback creates a new AnimationPlayback that plays back the animation.
func (a *Animation) Playback() *AnimationPlayback {
	return NewAnimationPlayback(a)
}

type animationBinding struct {
	translationKeyframes KeyframeList[dprec.Vec3]
	rotationKeyframes    KeyframeList[dprec.Quat]
	scaleKeyframes       KeyframeList[dprec.Vec3]
}

func (b animationBinding) Translation(timestamp float64) dprec.Vec3 {
	left, right, t := b.translationKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}

func (b animationBinding) Rotation(timestamp float64) dprec.Quat {
	left, right, t := b.rotationKeyframes.Keyframe(timestamp)
	return dprec.QuatSlerp(left.Value, right.Value, t)
}

func (b animationBinding) Scale(timestamp float64) dprec.Vec3 {
	left, right, t := b.scaleKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}
