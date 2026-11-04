package animation

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

// Node represents a single node in an animation tree. A node keeps track of
// its own playback position and produces the [Transform] of each bone at that
// position. Nodes can be combined into trees, where parent nodes derive their
// output from that of their children.
type Node interface {

	// Rate returns the fraction of the animation length that elapses each
	// second (fraction per second). A node with no meaningful length (for
	// example a static pose) returns 0.0.
	//
	// The rate is similar to a frequency: for an animation that plays at
	// normal speed it equals 1.0 / length, where length is the duration of
	// the animation in seconds. For example, a two second animation has a
	// rate of 0.5.
	Rate() float64

	// Fraction returns the amount of animation that has elapsed, in the
	// range [0.0..1.0]. For a looping node, 0.0 and 1.0 denote the same
	// position, so a fraction of 1.0 does not mean that playback has
	// finished.
	Fraction() float64

	// SetFraction relocates the animation to the specified fractional
	// position.
	SetFraction(fraction float64)

	// Advance moves the animation forward by the specified delta seconds.
	// A negative delta moves the animation backwards.
	//
	// A parent node is responsible for advancing its children. How and
	// whether each child is advanced, and whether a synchronized child is
	// kept at the fraction of the parent, depends on the kind of parent
	// node.
	Advance(deltaSeconds float64)

	// IsSynchronized returns whether the node should be kept in phase with
	// its parent.
	IsSynchronized() bool

	// SetSynchronized configures whether the node should be kept in phase
	// with its parent.
	SetSynchronized(synchronized bool)

	// Transform returns the transformation of the specified bone at the
	// current fractional position of the node. Bones that the node does not
	// animate return a fallback that depends on the kind of node:
	//
	//   - An absolute node produces the local transformation of the bone,
	//     relative to its parent. Bones that it does not animate return the
	//     bind pose transformation of the bone.
	//   - An additive node produces a difference that is meant to be applied
	//     on top of an absolute transformation through [AddTransforms].
	//     Bones that it does not animate return the [IdentityTransform].
	Transform(bone int) Transform

	// DeltaTransform returns the transformation that the specified bone
	// experiences throughout the next delta seconds. This is used for root
	// motion extraction. Bones that the node does not animate experience no
	// motion and return the [IdentityTransform].
	DeltaTransform(bone int, deltaSeconds float64) Transform
}

// Synchronized is a helper function that configures a node to be
// synchronized and returns it. This is useful during node tree construction.
func Synchronized[T Node](node T) T {
	node.SetSynchronized(true)
	return node
}

// synchronizationAspect is meant to be embedded in [Node] implementations to
// provide the [Node.IsSynchronized] and [Node.SetSynchronized] methods. Nodes
// start out as not synchronized.
type synchronizationAspect struct {
	synchronized bool
}

// IsSynchronized returns whether the node should be kept in phase with its
// parent.
func (s *synchronizationAspect) IsSynchronized() bool {
	return s.synchronized
}

// SetSynchronized configures whether the node should be kept in phase with
// its parent.
func (s *synchronizationAspect) SetSynchronized(synchronized bool) {
	s.synchronized = synchronized
}

// blendRates returns the rate of a blend of the two nodes at the specified
// weight, where a weight of 0.0 corresponds to the first node and a weight
// of 1.0 to the second one. Only synchronized nodes with a rate greater than
// 0.0 take part. If both take part, the cycle lengths of the nodes, rather
// than their rates, are blended linearly. If only one takes part, its rate
// is returned unchanged. If none takes part, the blend has no meaningful
// length and the result is 0.0.
func blendRates(first, second Node, weight float64) float64 {
	firstRate := first.Rate()
	firstOK := first.IsSynchronized() && firstRate > 0.0
	secondRate := second.Rate()
	secondOK := second.IsSynchronized() && secondRate > 0.0
	switch {
	case firstOK && secondOK:
		// The rates are swapped in the Mix on purpose. Blending the lengths
		// 1/firstRate and 1/secondRate linearly and taking the reciprocal
		// yields this expression.
		return firstRate * secondRate / dprec.Mix(secondRate, firstRate, weight)
	case firstOK:
		return firstRate
	case secondOK:
		return secondRate
	default:
		return 0.0
	}
}

// clampFraction clamps a fraction into the range [0.0..1.0].
func clampFraction(fraction float64) float64 {
	return dprec.Clamp(fraction, 0.0, 1.0)
}

// wrapFraction wraps a fraction into the range [0.0..1.0).
func wrapFraction(fraction float64) float64 {
	_, fraction = math.Modf(fraction)
	if fraction < 0.0 {
		fraction += 1.0
		// A tiny negative fraction can round up to exactly 1.0.
		if fraction >= 1.0 {
			fraction = 0.0
		}
	}
	return fraction
}

// clampTime clamps a time into the range [0..length].
func clampTime(t, length float64) float64 {
	return dprec.Clamp(t, 0.0, length)
}

// wrapTime wraps a time value into the range [0.0..length).
func wrapTime(t, length float64) float64 {
	t = math.Mod(t, length)
	if t < 0.0 {
		t += length
		// A tiny negative time can round up to exactly length.
		if t >= length {
			t = 0.0
		}
	}
	return t
}
