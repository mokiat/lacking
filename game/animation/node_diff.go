package animation

// NewDiffNode creates a node that returns the difference between two nodes.
func NewDiffNode(primary, secondary Node) *DiffNode {
	return &DiffNode{
		primary:   primary,
		secondary: secondary,
	}
}

// DiffNode returns the difference between two animations.
type DiffNode struct {
	primary   Node
	secondary Node
}

var _ Node = (*DiffNode)(nil)

// Rate returns the fraction of the animation length that advances each
// second.
func (n *DiffNode) Rate() float64 {
	return 1.0 // TODO: Figure this out!
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *DiffNode) Fraction() float64 {
	return n.primary.Fraction() // TODO: Figure this out.
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *DiffNode) SetFraction(fraction float64) {
	n.primary.SetFraction(fraction)
	n.secondary.SetFraction(fraction)

}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *DiffNode) Advance(seconds, synchronizationRate float64) {
	n.primary.Advance(seconds, synchronizationRate)
	n.secondary.Advance(seconds, synchronizationRate)
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *DiffNode) BoneTransform(bone string) NodeTransform {
	firstTransform := n.primary.BoneTransform(bone)
	secondTransform := n.primary.BoneTransform(bone)
	return DiffNodeTransforms(firstTransform, secondTransform)
}

// BoneDeltaTransform returns the transformation that the bone will experience
// throughout the next delta interval. This is used for root motion.
func (n *DiffNode) BoneDeltaTransform(bone string, delta float64) NodeTransform {
	firstTransform := n.primary.BoneDeltaTransform(bone, delta)
	secondTransform := n.primary.BoneDeltaTransform(bone, delta)
	return DiffNodeTransforms(firstTransform, secondTransform)
}
