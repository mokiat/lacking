package animation

import (
	"github.com/mokiat/gomath/dprec"
)

// NewPlaybackNode creates a simple animation Node that plays back a given
// source directly.
func NewPlaybackNode(source Source, loop bool) *PlaybackNode {
	return &PlaybackNode{
		source:            source,
		previousTimestamp: 0.0,
		currentTimestamp:  0.0,
		loop:              loop,
	}
}

// PlaybackNode represents an animation source that plays back an
// animation.
type PlaybackNode struct {
	source            Source
	previousTimestamp float64
	currentTimestamp  float64
	loop              bool
}

var _ Node = (*PlaybackNode)(nil)

// Source returns the underlying animation source.
func (n *PlaybackNode) Source() Source {
	return n.source
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *PlaybackNode) Reset() {
	n.previousTimestamp = n.currentTimestamp
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *PlaybackNode) Rate() float64 {
	return 1.0 / n.source.Length()
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *PlaybackNode) Fraction() float64 {
	fraction := n.currentTimestamp / n.source.Length()
	if n.loop {
		return wrapFraction(fraction)
	}
	return clampFraction(fraction)
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *PlaybackNode) SetFraction(fraction float64) {
	if n.loop {
		fraction = wrapFraction(fraction)
	} else {
		fraction = clampFraction(fraction)
	}
	n.currentTimestamp = fraction * n.source.Length()
	n.Reset()
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *PlaybackNode) Advance(seconds, synchronizationRate float64) {
	n.currentTimestamp += seconds * synchronizationRate
	if !n.loop {
		fraction := clampFraction(n.currentTimestamp / n.source.Length())
		n.currentTimestamp = fraction * n.source.Length()
	}
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *PlaybackNode) BoneTransform(bone string) NodeTransform {
	return n.source.BoneTransformAt(bone, n.currentTimestamp)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (n *PlaybackNode) BoneTransformDelta(bone string) NodeTransform {
	return n.source.BoneTransformDelta(bone, n.previousTimestamp, n.currentTimestamp)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (n *PlaybackNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	timestamp := dprec.Mix(n.previousTimestamp, n.currentTimestamp, fraction)
	return n.source.BoneTransformAt(bone, timestamp)
}
