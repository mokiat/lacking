package animation

import "github.com/mokiat/gog"

// NewOneOfNode creates an animation node that produces one specific animation
// from a set of animations.
func NewOneOfNode[T comparable](animations map[T]Node) *OneOfNode[T] {
	return &OneOfNode[T]{
		animations:         animations,
		activeAnimationKey: gog.Zero[T](),
		activeAnimation:    nil,
	}
}

// OneOfNode is an animation node that plays one specific animation
// from a set of animations. The active animation can be changed at any time.
type OneOfNode[T comparable] struct {
	animations         map[T]Node
	activeAnimationKey T
	activeAnimation    Node
}

var _ Node = (*OneOfNode[struct{}])(nil)

// ActiveAnimation returns the animation key that is playing.
func (n *OneOfNode[T]) ActiveAnimation() T {
	return n.activeAnimationKey
}

// PickAnimation changes to the specified animation.
//
// The rewind flag controls whether the new animation should start from the
// beginning.
func (n *OneOfNode[T]) PickAnimation(key T, rewind bool) {
	n.activeAnimationKey = key
	n.activeAnimation = n.animations[key]
	if rewind && (n.activeAnimation != nil) {
		n.activeAnimation.SetFraction(0.0)
	}
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *OneOfNode[T]) Reset() {
	for _, node := range n.animations {
		node.Reset()
	}
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *OneOfNode[T]) Rate() float64 {
	if n.activeAnimation == nil {
		return 1.0
	}
	return n.activeAnimation.Rate()
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *OneOfNode[T]) Fraction() float64 {
	if n.activeAnimation != nil {
		return n.activeAnimation.Fraction()
	}
	return 0.0
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *OneOfNode[T]) SetFraction(fraction float64) {
	if n.activeAnimation != nil {
		n.activeAnimation.SetFraction(fraction)
	}
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *OneOfNode[T]) Advance(seconds, synchronizationRate float64) {
	if n.activeAnimation != nil {
		n.activeAnimation.Advance(seconds, synchronizationRate)
	}
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *OneOfNode[T]) BoneTransform(bone string) NodeTransform {
	if n.activeAnimation == nil {
		return NodeTransform{}
	}
	return n.activeAnimation.BoneTransform(bone)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (n *OneOfNode[T]) BoneTransformDelta(bone string) NodeTransform {
	if n.activeAnimation == nil {
		return NodeTransform{}
	}
	return n.activeAnimation.BoneTransformDelta(bone)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (n *OneOfNode[T]) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	if n.activeAnimation == nil {
		return NodeTransform{}
	}
	return n.activeAnimation.BoneTransformInterpolation(bone, fraction)
}
