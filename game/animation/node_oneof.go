package animation

// NewOneOfNode creates an animation node that produces one specific animation
// from a set of animations.
func NewOneOfNode(animations map[string]Node) *OneOfNode {
	return &OneOfNode{
		animations:          animations,
		activeAnimationName: "",
		activeAnimation:     nil,
	}
}

var _ Node = (*OneOfNode)(nil)

// OneOfNode is an animation node that plays one specific animation
// from a set of animations. The active animation can be changed at any time.
type OneOfNode struct {
	animations          map[string]Node
	activeAnimationName string
	activeAnimation     Node
}

// ActiveAnimation returns the underlying node that is currently being used.
func (s *OneOfNode) ActiveAnimation() Node {
	return s.activeAnimation
}

// PickAnimation changes to the specified animation.
//
// The reset flag controls whether the new animation should start from the
// beginning.
func (s *OneOfNode) PickAnimation(name string, reset bool) {
	s.activeAnimationName = name
	s.activeAnimation = s.animations[name]
	if reset && (s.activeAnimation != nil) {
		s.activeAnimation.SetProgress(0.0)
		s.activeAnimation.Reset()
	}
}

// Rate returns the fraction of the animation length that advances each
// second.
func (s *OneOfNode) Rate() float64 {
	if s.activeAnimation == nil {
		return 1.0
	}
	return s.activeAnimation.Rate()
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (s *OneOfNode) Reset() {
	for _, node := range s.animations {
		node.Reset()
	}
}

// Progress returns the current fraction of the animation that has
// advanced since the start.
//
// This value will always be in the range [0.0..1.0).
func (s *OneOfNode) Progress() float64 {
	if s.activeAnimation == nil {
		return 0.0
	}
	return s.activeAnimation.Progress()
}

// SetProgress changes the current position of the animation to the
// specified fraction.
//
// It is possible to set this value above 1.0, and in fact is necessary
// during update, so that it can handle loops and interpolation correctly,
// as setting the value directly to the wrapped-around value might indicate
// a reverse animation or a fractional animation.
//
// Internally, once applied, the progress will be normalized to [0.0..1.0).
func (s *OneOfNode) SetProgress(fraction float64) {
	if s.activeAnimation != nil {
		s.activeAnimation.SetProgress(fraction)
	}
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (s *OneOfNode) BoneTransform(bone string) NodeTransform {
	if s.activeAnimation == nil {
		return NodeTransform{}
	}
	return s.activeAnimation.BoneTransform(bone)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (s *OneOfNode) BoneTransformDelta(bone string) NodeTransform {
	if s.activeAnimation == nil {
		return NodeTransform{}
	}
	return s.activeAnimation.BoneTransformDelta(bone)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (s *OneOfNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	if s.activeAnimation == nil {
		return NodeTransform{}
	}
	return s.activeAnimation.BoneTransformInterpolation(bone, fraction)
}
