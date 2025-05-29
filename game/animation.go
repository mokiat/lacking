package game

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// Animation represents an instantiation of a keyframe animation.
type Animation struct {
	name      string
	startTime float64
	endTime   float64
	loop      bool
	bindings  map[string]AnimationKeyframeSet
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
	if len(binding.TranslationKeyframes) > 0 {
		result.Translation = opt.V(binding.Translation(timestamp))
	}
	if len(binding.RotationKeyframes) > 0 {
		result.Rotation = opt.V(binding.Rotation(timestamp))
	}
	if len(binding.ScaleKeyframes) > 0 {
		result.Scale = opt.V(binding.Scale(timestamp))
	}
	return result
}

// Playback creates a new AnimationPlayback that plays back the animation.
func (a *Animation) Playback() *AnimationPlayback {
	return NewAnimationPlayback(a)
}

type AnimationKeyframeSet struct {
	TranslationKeyframes KeyframeList[dprec.Vec3]
	RotationKeyframes    KeyframeList[dprec.Quat]
	ScaleKeyframes       KeyframeList[dprec.Vec3]
}

func (b AnimationKeyframeSet) Translation(timestamp float64) dprec.Vec3 {
	left, right, t := b.TranslationKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}

func (b AnimationKeyframeSet) Rotation(timestamp float64) dprec.Quat {
	left, right, t := b.RotationKeyframes.Keyframe(timestamp)
	return dprec.QuatSlerp(left.Value, right.Value, t)
}

func (b AnimationKeyframeSet) Scale(timestamp float64) dprec.Vec3 {
	left, right, t := b.ScaleKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}
