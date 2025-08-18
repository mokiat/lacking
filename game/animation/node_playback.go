package animation

// NewPlaybackNode creates a simple animation Node that plays back a given
// source directly.
func NewPlaybackNode(source Source, loop bool) *PlaybackNode {
	return &PlaybackNode{
		source:           source,
		currentTimestamp: 0.0,
		loop:             loop,
	}
}

// PlaybackNode represents an animation source that plays back an
// animation.
type PlaybackNode struct {
	source           Source
	currentTimestamp float64
	loop             bool
}

var _ Node = (*PlaybackNode)(nil)

// Source returns the underlying animation source.
func (n *PlaybackNode) Source() Source {
	return n.source
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

// BoneDeltaTransform returns the transformation that the bone will experience
// throughout the next delta interval. This is used for root motion.
func (n *PlaybackNode) BoneDeltaTransform(bone string, delta float64) NodeTransform {
	return n.source.BoneTransformDelta(bone, n.currentTimestamp, n.currentTimestamp+delta)
}
