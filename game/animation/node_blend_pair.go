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
		first:              first,
		second:             second,
		progress:           0.0,
		factor:             0.0,
		firstSynchronized:  true,
		secondSynchronized: true,
	}
}

var _ Node = (*BlendPairNode)(nil)

// BlendPairNode represents an animation node that blends two child
// animation nodes. The blending factor is determined by the factor
// field of the node.
type BlendPairNode struct {
	first              Node
	second             Node
	progress           float64
	factor             float64
	firstSynchronized  bool
	secondSynchronized bool
}

// FirstSynchronized returns whether the the first blend node will be
// synchronized with the tree hierarchy.
func (s *BlendPairNode) FirstSynchronized() bool {
	return s.firstSynchronized
}

// SetFirstSynchronized sets whether the first blend node should be
// synchronized with the tree hierarchy.
func (s *BlendPairNode) SetFirstSynchronized(synchronized bool) *BlendPairNode {
	s.firstSynchronized = synchronized
	return s
}

// SecondSynchronized returns whether the the second blend node will be
// synchronized with the tree hierarchy.
func (s *BlendPairNode) SecondSynchronized() bool {
	return s.secondSynchronized
}

// SetSecondSynchronized sets whether the second blend node should be
// synchronized with the tree hierarchy.
func (s *BlendPairNode) SetSecondSynchronized(synchronized bool) *BlendPairNode {
	s.secondSynchronized = synchronized
	return s
}

// Factor returns the blending factor of the node. A value of 0.0 means
// that the first blend node is used, a value of 1.0 means that the second
// blend node is used. The value is clamped to the range [0.0, 1.0].
func (s *BlendPairNode) Factor() float64 {
	return s.factor
}

// SetFactor sets the blending factor of the node. The value is clamped
// to the range [0.0, 1.0].
func (s *BlendPairNode) SetFactor(factor float64) *BlendPairNode {
	s.factor = dprec.Clamp(factor, 0.0, 1.0)
	return s
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (s *BlendPairNode) Reset() {
	_, fraction := math.Modf(s.progress)
	s.Seek(fraction)

	s.first.Reset()
	s.second.Reset()
}

// Rate returns the fraction of the animation length that advances each
// second.
func (s *BlendPairNode) Rate() float64 {
	switch {
	case s.firstSynchronized && s.secondSynchronized:
		firstRate := s.first.Rate()
		secondRate := s.second.Rate()
		// NOTE: The rates are flipped in the denominator on purpose. This is how
		// the math ends up if you derive this from length blending.
		return firstRate * secondRate / dprec.Mix(secondRate, firstRate, s.factor)
	case s.firstSynchronized:
		return s.first.Rate()
	case s.secondSynchronized:
		return s.second.Rate()
	default:
		return 1.0
	}
}

// Seek relocates the animation to the specified position (fractional).
//
// NOTE: This resets the animation and accumulated delta is lost.
func (s *BlendPairNode) Seek(fraction float64) {
	s.progress = fraction

	if s.firstSynchronized {
		s.first.Seek(s.progress)
	}

	if s.secondSynchronized {
		s.second.Seek(s.progress)
	}
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (s *BlendPairNode) Advance(seconds, synchronizationRate float64) {
	rate := s.Rate()
	s.progress += rate * seconds * synchronizationRate

	if s.firstSynchronized {
		adjustedRate := rate / s.first.Rate()
		s.first.Advance(seconds, synchronizationRate*adjustedRate)
	} else {
		s.first.Advance(seconds, 1.0) // drop syncrhonization
	}

	if s.secondSynchronized {
		adjustedRate := rate / s.second.Rate()
		s.second.Advance(seconds, synchronizationRate*adjustedRate)
	} else {
		s.second.Advance(seconds, 1.0) // drop synchronization
	}
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (s *BlendPairNode) BoneTransform(bone string) NodeTransform {
	firstTransform := s.first.BoneTransform(bone)
	secondTransform := s.second.BoneTransform(bone)
	return BlendNodeTransforms(firstTransform, secondTransform, s.factor)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (s *BlendPairNode) BoneTransformDelta(bone string) NodeTransform {
	firstTransform := s.first.BoneTransformDelta(bone)
	secondTransform := s.second.BoneTransformDelta(bone)
	return BlendNodeTransforms(firstTransform, secondTransform, s.factor)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (s *BlendPairNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	firstTransform := s.first.BoneTransformInterpolation(bone, fraction)
	secondTransform := s.second.BoneTransformInterpolation(bone, fraction)
	return BlendNodeTransforms(firstTransform, secondTransform, s.factor)
}
