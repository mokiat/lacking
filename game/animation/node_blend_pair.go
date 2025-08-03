package animation

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

// NewBlendPairNode creates a new animation blending node that blends between
// a pair of sub-nodes.
func NewBlendPairNode(first, second Node) *BlendPairNode {
	if first == second {
		panic("the nodes need to be different")
	}
	return &BlendPairNode{
		first:        first,
		second:       second,
		progress:     0.0,
		factor:       0.0,
		synchronized: true,
	}
}

var _ Node = (*BlendPairNode)(nil)

// BlendPairNode represents an animation node that blends two child
// animation nodes. The blending factor is determined by the factor
// field of the node.
type BlendPairNode struct {
	first        Node
	second       Node
	progress     float64
	factor       float64
	synchronized bool
}

// First returns the first blend node.
func (s *BlendPairNode) First() Node {
	return s.first
}

// Second returns the second blend node.
func (s *BlendPairNode) Second() Node {
	return s.second
}

// Synchronized returns whether the two blend nodes should be synchronized.
func (s *BlendPairNode) Synchronized() bool {
	return s.synchronized
}

// SetSynchronized sets whether the two blend nodes should be synchronized.
func (s *BlendPairNode) SetSynchronized(synchronized bool) {
	s.synchronized = synchronized
}

// Factor returns the blending factor of the node. A value of 0.0 means
// that the first blend node is used, a value of 1.0 means that the second
// blend node is used. The value is clamped to the range [0.0, 1.0].
func (s *BlendPairNode) Factor() float64 {
	return s.factor
}

// SetFactor sets the blending factor of the node. The value is clamped
// to the range [0.0, 1.0].
func (s *BlendPairNode) SetFactor(factor float64) {
	s.factor = dprec.Clamp(factor, 0.0, 1.0)
}

// Rate returns the fraction of the animation length that advances each
// second.
func (s *BlendPairNode) Rate() float64 {
	if s.synchronized {
		firstRate := s.first.Rate()
		secondRate := s.second.Rate()
		// NOTE: The rates are flipped in the denominator on purpose. This is how
		// the math ends up if you derive this from lengths.
		return firstRate * secondRate / dprec.Mix(secondRate, firstRate, s.factor)
	} else {
		return s.first.Rate()
	}
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (s *BlendPairNode) Reset() {
	// Normalize.
	s.SetProgress(s.progress)
	// Reset stored delta
	s.first.Reset()
	s.second.Reset()
}

// Progress returns the current fraction of the animation that has
// advanced since the start.
//
// This value will always be in the range [0.0..1.0).
func (s *BlendPairNode) Progress() float64 {
	_, fraction := math.Modf(s.progress)
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
func (s *BlendPairNode) SetProgress(fraction float64) {
	s.progress = fraction
	s.first.SetProgress(fraction)
	s.second.SetProgress(fraction)
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (s *BlendPairNode) BoneTransform(bone string) NodeTransform {
	switch {
	case s.factor < 0.000001: // optimization
		return s.first.BoneTransform(bone)
	case s.factor > 0.999999: // optimization
		return s.second.BoneTransform(bone)
	default:
		firstTransform := s.first.BoneTransform(bone)
		secondTransform := s.second.BoneTransform(bone)
		return BlendNodeTransforms(firstTransform, secondTransform, s.factor)
	}
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (s *BlendPairNode) BoneTransformDelta(bone string) NodeTransform {
	switch {
	case s.factor < 0.000001: // quick solution
		return s.first.BoneTransformDelta(bone)
	case s.factor > 0.999999: // quick solution
		return s.second.BoneTransformDelta(bone)
	default:
		firstTransform := s.first.BoneTransformDelta(bone)
		secondTransform := s.second.BoneTransformDelta(bone)
		return BlendNodeTransforms(firstTransform, secondTransform, s.factor)
	}
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (s *BlendPairNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	switch {
	case s.factor < 0.000001: // optimization
		return s.first.BoneTransformInterpolation(bone, fraction)
	case s.factor > 0.999999: // optimization
		return s.second.BoneTransformInterpolation(bone, fraction)
	default:
		firstTransform := s.first.BoneTransformInterpolation(bone, fraction)
		secondTransform := s.second.BoneTransformInterpolation(bone, fraction)
		return BlendNodeTransforms(firstTransform, secondTransform, s.factor)
	}
}
