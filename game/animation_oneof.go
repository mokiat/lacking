package game

// NewOneOfAnimation creates an animation source that plays one specific
// animation from a set of animations.
func NewOneOfAnimation(animations map[string]AnimationSource) *OneOfAnimation {
	var anyAnimation AnimationSource
	for _, animation := range animations {
		anyAnimation = animation
		break
	}
	return &OneOfAnimation{
		animations:      animations,
		activeAnimation: anyAnimation,
	}
}

var _ AnimationSource = (*OneOfAnimation)(nil)

// OneOfAnimation is an animation source that plays one specific animation
// from a set of animations. The active animation can be changed at any time.
type OneOfAnimation struct {
	animations      map[string]AnimationSource
	activeAnimation AnimationSource
}

// PickAnimation changes to the specified animation.
func (a *OneOfAnimation) PickAnimation(name string) {
	a.activeAnimation = a.animations[name]
}

// Length returns the length of the currently active animation.
func (a *OneOfAnimation) Length() float64 {
	return a.activeAnimation.Length()
}

// Position returns the current position of the currently active animation.
func (a *OneOfAnimation) Position() float64 {
	return a.activeAnimation.Position()
}

// SetPosition sets the current position of the currently active animation.
//
// It can be useful to set the position to zero when changing to a non-looping
// animation so that it starts from the beginning.
func (a *OneOfAnimation) SetPosition(position float64) {
	a.activeAnimation.SetPosition(position)
}

// NodeTransform returns the transform of the specified node in the currently
// active animation.
func (a *OneOfAnimation) NodeTransform(name string) NodeTransform {
	return a.activeAnimation.NodeTransform(name)
}
