package animation

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

// NewPlaybackNode creates a simple animation Node that plays back a given
// source directly.
func NewPlaybackNode(source Source) *PlaybackNode {
	return &PlaybackNode{
		source:            source,
		previousTimestamp: 0.0,
		currentTimestamp:  0.0,
	}
}

var _ Node = (*PlaybackNode)(nil)

// PlaybackNode represents an animation source that plays back an
// animation.
type PlaybackNode struct {
	source            Source
	previousTimestamp float64
	currentTimestamp  float64
}

// Source returns the underlying animation source.
func (p *PlaybackNode) Source() Source {
	return p.source
}

// Rate returns the fraction of the animation length that advances each
// second.
func (p *PlaybackNode) Rate() float64 {
	return 1.0 / p.source.Length()
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (p *PlaybackNode) Reset() {
	p.SetProgress(p.Progress()) // normalize
	p.previousTimestamp = p.currentTimestamp
}

// Progress returns the current fraction of the animation that has
// advanced since the start.
//
// This value will always be in the range [0.0..1.0).
func (p *PlaybackNode) Progress() float64 {
	_, fraction := math.Modf(p.currentTimestamp / p.source.Length())
	if fraction < 0.0 {
		fraction += 1.0
	}
	return fraction
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
func (p *PlaybackNode) SetProgress(fraction float64) {
	p.currentTimestamp = fraction * p.source.Length()
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (p *PlaybackNode) BoneTransform(bone string) NodeTransform {
	return p.source.BoneTransformAt(bone, p.currentTimestamp)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (p *PlaybackNode) BoneTransformDelta(bone string) NodeTransform {
	return p.source.BoneTransformDelta(bone, p.previousTimestamp, p.currentTimestamp)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (p *PlaybackNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	timestamp := dprec.Mix(p.previousTimestamp, p.currentTimestamp, fraction)
	return p.source.BoneTransformAt(bone, timestamp)
}
