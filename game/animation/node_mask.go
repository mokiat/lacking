package animation

import "github.com/mokiat/gog/filter"

// NewMaskNode creates a new animation node that picks specific bones
// from the specified node.
func NewMaskNode(delegate Node, selection filter.Func[string]) *MaskNode {
	return &MaskNode{
		delegate:  delegate,
		selection: selection,
	}
}

// MaskNode is an animation source that picks specific bones
// from another animation source.
type MaskNode struct {
	delegate  Node
	selection filter.Func[string]
}

var _ Node = (*MaskNode)(nil)

// Rate returns the fraction of the animation length that advances each
// second.
func (n *MaskNode) Rate() float64 {
	return n.delegate.Rate()
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *MaskNode) Fraction() float64 {
	return n.delegate.Fraction()
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *MaskNode) SetFraction(fraction float64) {
	n.delegate.SetFraction(fraction)
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *MaskNode) Advance(seconds, synchronizationRate float64) {
	n.delegate.Advance(seconds, synchronizationRate)
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *MaskNode) BoneTransform(bone string) NodeTransform {
	if !n.selection(bone) {
		return NodeTransform{}
	}
	return n.delegate.BoneTransform(bone)
}

// BoneDeltaTransform returns the transformation that the bone will experience
// throughout the next delta interval. This is used for root motion.
func (n *MaskNode) BoneDeltaTransform(bone string, delta float64) NodeTransform {
	if !n.selection(bone) {
		return NodeTransform{}
	}
	return n.delegate.BoneDeltaTransform(bone, delta)
}
