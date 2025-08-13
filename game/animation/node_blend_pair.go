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
func (n *BlendPairNode) FirstSynchronized() bool {
	return n.firstSynchronized
}

// SetFirstSynchronized sets whether the first blend node should be
// synchronized with the tree hierarchy.
func (n *BlendPairNode) SetFirstSynchronized(synchronized bool) *BlendPairNode {
	n.firstSynchronized = synchronized
	return n
}

// SecondSynchronized returns whether the the second blend node will be
// synchronized with the tree hierarchy.
func (n *BlendPairNode) SecondSynchronized() bool {
	return n.secondSynchronized
}

// SetSecondSynchronized sets whether the second blend node should be
// synchronized with the tree hierarchy.
func (n *BlendPairNode) SetSecondSynchronized(synchronized bool) *BlendPairNode {
	n.secondSynchronized = synchronized
	return n
}

// Factor returns the blending factor of the node. A value of 0.0 means
// that the first blend node is used, a value of 1.0 means that the second
// blend node is used. The value is clamped to the range [0.0, 1.0].
func (n *BlendPairNode) Factor() float64 {
	return n.factor
}

// SetFactor sets the blending factor of the node. The value is clamped
// to the range [0.0, 1.0].
func (n *BlendPairNode) SetFactor(factor float64) *BlendPairNode {
	n.factor = dprec.Clamp(factor, 0.0, 1.0)
	return n
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *BlendPairNode) Reset() {
	_, fraction := math.Modf(n.progress)
	n.Seek(fraction)

	n.first.Reset()
	n.second.Reset()
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *BlendPairNode) Rate() float64 {
	switch {
	case n.firstSynchronized && n.secondSynchronized:
		firstRate := n.first.Rate()
		secondRate := n.second.Rate()
		// NOTE: The rates are flipped in the denominator on purpose. This is how
		// the math ends up if you derive this from length blending.
		return firstRate * secondRate / dprec.Mix(secondRate, firstRate, n.factor)
	case n.firstSynchronized:
		return n.first.Rate()
	case n.secondSynchronized:
		return n.second.Rate()
	default:
		return 1.0
	}
}

// Seek relocates the animation to the specified position (fractional).
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *BlendPairNode) Seek(fraction float64) {
	n.progress = fraction

	if n.firstSynchronized {
		n.first.Seek(n.progress)
	}

	if n.secondSynchronized {
		n.second.Seek(n.progress)
	}
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *BlendPairNode) Advance(seconds, synchronizationRate float64) {
	rate := n.Rate()
	n.progress += rate * seconds * synchronizationRate

	if n.firstSynchronized {
		adjustedRate := rate / n.first.Rate()
		n.first.Advance(seconds, synchronizationRate*adjustedRate)
	} else {
		n.first.Advance(seconds, 1.0) // drop syncrhonization
	}

	if n.secondSynchronized {
		adjustedRate := rate / n.second.Rate()
		n.second.Advance(seconds, synchronizationRate*adjustedRate)
	} else {
		n.second.Advance(seconds, 1.0) // drop synchronization
	}
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *BlendPairNode) BoneTransform(bone string) NodeTransform {
	firstTransform := n.first.BoneTransform(bone)
	secondTransform := n.second.BoneTransform(bone)
	return BlendNodeTransforms(firstTransform, secondTransform, n.factor)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (n *BlendPairNode) BoneTransformDelta(bone string) NodeTransform {
	firstTransform := n.first.BoneTransformDelta(bone)
	secondTransform := n.second.BoneTransformDelta(bone)
	return BlendNodeTransforms(firstTransform, secondTransform, n.factor)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (n *BlendPairNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	firstTransform := n.first.BoneTransformInterpolation(bone, fraction)
	secondTransform := n.second.BoneTransformInterpolation(bone, fraction)
	return BlendNodeTransforms(firstTransform, secondTransform, n.factor)
}
