package animation

import (
	"github.com/mokiat/gomath/dprec"
)

// NewBlendPairNode creates a new animation blending node that blends between
// a pair of sub-nodes.
func NewBlendPairNode(first, second Node) *BlendPairNode {
	if first == second {
		panic("the nodes need to be different")
	}
	return &BlendPairNode{
		first:       first,
		second:      second,
		progress:    0.0,
		blendFactor: 0.0,
	}
}

// BlendPairNode represents an animation node that blends two child
// animation nodes. The blending factor is determined by the factor
// field of the node.
type BlendPairNode struct {
	first        Node
	second       Node
	progress     float64
	blendFactor  float64
	synchronized bool
}

var _ Node = (*BlendPairNode)(nil)

// BlendFactor returns the blending factor of the node. A value of 0.0 means
// that the first blend node is used, a value of 1.0 means that the second
// blend node is used. The value is clamped to the range [0.0, 1.0].
func (n *BlendPairNode) BlendFactor() float64 {
	return n.blendFactor
}

// SetBlendFactor sets the blending factor of the node. The value is clamped
// to the range [0.0, 1.0].
func (n *BlendPairNode) SetBlendFactor(factor float64) *BlendPairNode {
	n.blendFactor = dprec.Clamp(factor, 0.0, 1.0)
	return n
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *BlendPairNode) Rate() float64 {
	return blendRates(n.first, n.second, n.blendFactor)
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *BlendPairNode) Fraction() float64 {
	return wrapFraction(n.progress)
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *BlendPairNode) SetFraction(fraction float64) {
	n.progress = wrapFraction(fraction)

	if n.first.IsSynchronized() {
		n.first.SetFraction(n.progress)
	}

	if n.second.IsSynchronized() {
		n.second.SetFraction(n.progress)
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
	n.progress = wrapFraction(n.progress)

	if n.first.IsSynchronized() {
		adjustedRate := rate / n.first.Rate()
		n.first.Advance(seconds, synchronizationRate*adjustedRate)
	} else {
		n.first.Advance(seconds, 1.0)
	}

	if n.second.IsSynchronized() {
		adjustedRate := rate / n.second.Rate()
		n.second.Advance(seconds, synchronizationRate*adjustedRate)
	} else {
		n.second.Advance(seconds, 1.0)
	}
}

// IsSynchronized returns whether the node should be synchronized.
func (n *BlendPairNode) IsSynchronized() bool {
	return n.synchronized
}

// SetSynchronized configures whether the node should be synchronized.
func (n *BlendPairNode) SetSynchronized(synchronized bool) {
	n.synchronized = synchronized
}

// Synchronize is called each frame to allow a node to synchronized its
// children (depending on their setting).
//
// This will be called (and should be called on children) regardless if
// the current or any child node is synchronized or not.
func (n *BlendPairNode) Synchronize() {
	if n.first.IsSynchronized() {
		n.first.SetFraction(n.progress)
	}
	n.first.Synchronize()
	if n.second.IsSynchronized() {
		n.second.SetFraction(n.progress)
	}
	n.second.Synchronize()
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *BlendPairNode) BoneTransform(bone string) NodeTransform {
	firstTransform := n.first.BoneTransform(bone)
	secondTransform := n.second.BoneTransform(bone)
	return BlendNodeTransforms(firstTransform, secondTransform, n.blendFactor)
}

// BoneDeltaTransform returns the transformation that the bone will experience
// throughout the next delta interval. This is used for root motion.
func (n *BlendPairNode) BoneDeltaTransform(bone string, delta float64) NodeTransform {
	firstTransform := n.first.BoneDeltaTransform(bone, delta)
	secondTransform := n.second.BoneDeltaTransform(bone, delta)
	return BlendNodeTransforms(firstTransform, secondTransform, n.blendFactor)
}
