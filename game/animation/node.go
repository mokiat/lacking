package animation

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

const (
	minFraction = 0.0
	maxFraction = 0.999999999
)

// Node represents an animation logic.
type Node interface {

	// Rate returns the fraction of the animation length that advances each
	// second (fraction per second).
	Rate() float64

	// Fraction returns the amount of animation that has elapsed. In case of
	// looping, the value will wrap around.
	//
	// The returned value is in the range [0.0..1.0).
	Fraction() float64

	// SetFraction relocates the animation to the specified fractional position.
	//
	// NOTE: This resets the animation and accumulated delta is lost.
	SetFraction(fraction float64)

	// Advance moves the animation forward by the specified delta seconds.
	//
	// The synchronizationRate determines the amount of scaling on the seconds
	// that should be applied in order to be correctly synchronized with sibling
	// and parent nodes in case of synchronization.
	Advance(seconds, synchronizationRate float64)

	// BoneTransform returns the transformation of the specified bone. Keep in
	// mind that this is after a fixed interval update has been applied. If
	// this is called from within a dynamic update handler, the
	// BoneTransformInterpolation method should be used instead.
	BoneTransform(bone string) NodeTransform

	// BoneDeltaTransform returns the transformation that the bone will experience
	// throughout the next delta interval. This is used for root motion.
	BoneDeltaTransform(bone string, delta float64) NodeTransform
}

func wrapFraction(fraction float64) float64 {
	_, fraction = math.Modf(fraction)
	if fraction < 0.0 {
		fraction += 1.0
	}
	return clampFraction(fraction)
}

func clampFraction(fraction float64) float64 {
	return dprec.Clamp(fraction, minFraction, maxFraction)
}
