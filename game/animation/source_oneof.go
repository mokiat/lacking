package animation

// NewOneOfSource creates an animation source that plays one specific
// animation from a set of animations.
func NewOneOfSource(animations map[string]Source) *OneOfSource {
	var anyAnimation Source
	for _, animation := range animations {
		anyAnimation = animation
		break
	}
	return &OneOfSource{
		animations:      animations,
		activeAnimation: anyAnimation,
	}
}

var _ Source = (*OneOfSource)(nil)

// OneOfSource is an animation source that plays one specific animation
// from a set of animations. The active animation can be changed at any time.
type OneOfSource struct {
	animations      map[string]Source
	activeAnimation Source
}

// PickAnimation changes to the specified animation.
func (s *OneOfSource) PickAnimation(name string) {
	s.activeAnimation = s.animations[name]
}

// Length returns the length of the currently active animation.
func (s *OneOfSource) Length() float64 {
	return s.activeAnimation.Length()
}

// Position returns the current position of the currently active animation.
func (s *OneOfSource) Position() float64 {
	return s.activeAnimation.Position()
}

// SetPosition sets the current position of the currently active animation.
//
// It can be useful to set the position to zero when changing to a non-looping
// animation so that it starts from the beginning.
func (s *OneOfSource) SetPosition(position float64) {
	s.activeAnimation.SetPosition(position)
}

// NodeTransform returns the transform of the specified node in the currently
// active animation.
func (s *OneOfSource) NodeTransform(name string) NodeTransform {
	return s.activeAnimation.NodeTransform(name)
}
