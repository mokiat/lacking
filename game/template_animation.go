package game

import (
	"github.com/mokiat/gog/opt"
)

// AnimationTemplate represents a template for an animation, likely
// loaded from an asset.
//
// NOTE: Once used to instantiate an animation, the template should not be
// modified as it may lead to unexpected behavior in instances.
type AnimationTemplate struct {

	// ID is the unique identifier of the animation template in the scope
	// of the asset file it was loaded from.
	ID uint32

	// Name is the name of the animation.
	Name string

	// StartTime is the time (in seconds) at which the animation starts.
	StartTime float64

	// EndTime is the time (in seconds) at which the animation ends.
	EndTime float64

	// Loop specifies whether the animation should loop.
	Loop bool

	// Bindings is a map of node names to their keyframe sets that are affected
	// by the animation.
	Bindings map[string]AnimationKeyframeSet
}

// AnimationInfo contains information needed to instantiate an Animation.
type AnimationInfo struct {

	// Template is the definition of the animation.
	Template *AnimationTemplate

	// ClipStart, if specified, overrides the start time of the animation.
	ClipStart opt.T[float64]

	// ClipEnd, if specified, overrides the end time of the animation.
	ClipEnd opt.T[float64]

	// Loop, if specified, overrides the loop setting of the animation.
	Loop opt.T[bool]
}

// InstantiateAnimation creates an instance of an Animation based on the
// provided info.
func (s *Scene) InstantiateAnimation(info AnimationInfo) Identifiable[*Animation] {
	template := info.Template
	animation := &Animation{
		name:      template.Name,
		startTime: info.ClipStart.ValueOrDefault(template.StartTime),
		endTime:   info.ClipEnd.ValueOrDefault(template.EndTime),
		loop:      info.Loop.ValueOrDefault(template.Loop),
		bindings:  template.Bindings,
	}
	return Identifiable[*Animation]{
		ID:    template.ID,
		Value: animation,
	}
}

// AnimationSetTemplate represents a template for a set of animations.
//
// NOTE: Once used to instantiate an animation set, the template should not be
// modified as it may lead to unexpected behavior in instances.
type AnimationSetTemplate struct {

	// Animations is a list of animation templates that are part of this set.
	Animations []AnimationTemplate
}

// AnimationSetInfo contains information needed to instantiate an
// AnimationSetTemplate.
type AnimationSetInfo struct {

	// Template is the definition of the animation set.
	Template *AnimationSetTemplate
}

// AnimationSet represents a set of animations that can be instantiated
// and used in a Scene.
type AnimationSet struct {

	// Animations is a list of animations that are part of this set.
	Animations IdentifiableList[*Animation]
}

// InstantiateAnimationSet creates an instance of an AnimationSet based on the
// provided info.
func (s *Scene) InstantiateAnimationSet(info AnimationSetInfo) *AnimationSet {
	template := info.Template

	animations := make([]Identifiable[*Animation], len(template.Animations))
	for i := range template.Animations {
		animationTemplate := &template.Animations[i]
		animations[i] = s.InstantiateAnimation(AnimationInfo{
			Template: animationTemplate,
		})
	}

	return &AnimationSet{
		Animations: animations,
	}
}
