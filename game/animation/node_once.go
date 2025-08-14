package animation

import "github.com/mokiat/gomath/dprec"

// TODO: Rework node to use a form of blending with two inputs - one that
// is the main Node and a node to be played once. There can also be
// fade-in and fade-out settings that would internally use blending.

// NewOnceNode creates a node that can play an animation once at a time,
// per user request.
func NewOnceNode(delegate Node) *OnceNode {
	return &OnceNode{
		delegate: delegate,
		progress: 0.0,
		active:   false,
	}
}

// OnceNode allows an animation to be played just once.
type OnceNode struct {
	delegate Node
	progress float64
	active   bool
}

var _ Node = (*OnceNode)(nil)

// Trigger rewinds and activates the animation to be played once.
func (n *OnceNode) Trigger() *OnceNode {
	n.progress = 0.0
	n.delegate.SetFraction(0.0)
	n.active = true
	return n
}

func (n *OnceNode) Active() bool {
	return n.active
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *OnceNode) Reset() {
	n.delegate.Reset()
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *OnceNode) Rate() float64 {
	return n.delegate.Rate()
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *OnceNode) Fraction() float64 {
	return dprec.Clamp(n.progress, minFraction, maxFraction)
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *OnceNode) SetFraction(fraction float64) {
	n.progress = fraction
	n.delegate.SetFraction(fraction)
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *OnceNode) Advance(seconds, synchronizationRate float64) {
	// TODO: Consider n.progress = n.Fraction() + n.Rate() * seconds * synchronizationRate
	n.progress += n.Rate() * seconds * synchronizationRate
	n.delegate.Advance(seconds, synchronizationRate)
	if n.progress >= 1.0 {
		n.active = false
	}
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *OnceNode) BoneTransform(bone string) NodeTransform {
	if !n.active {
		return NodeTransform{}
	}
	return n.delegate.BoneTransform(bone)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (n *OnceNode) BoneTransformDelta(bone string) NodeTransform {
	if !n.active {
		return NodeTransform{}
	}
	return n.delegate.BoneTransformDelta(bone)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (n *OnceNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	if !n.active {
		return NodeTransform{}
	}
	return n.delegate.BoneTransformInterpolation(bone, fraction)
}
