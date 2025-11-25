package animation

// NewDiffNode creates a node that returns the difference between two nodes.
func NewDiffNode(primary, overlay Node) *DiffNode {
	return &DiffNode{
		primary: primary,
		overlay: overlay,
	}
}

// DiffNode returns the difference between two animations.
type DiffNode struct {
	primary Node
	overlay Node

	synchronized bool
}

var _ Node = (*DiffNode)(nil)

// Rate returns the fraction of the animation length that advances each
// second.
func (n *DiffNode) Rate() float64 {
	return n.primary.Rate()
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *DiffNode) Fraction() float64 {
	return n.primary.Fraction()
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *DiffNode) SetFraction(fraction float64) {
	n.primary.SetFraction(fraction)
	if n.overlay.IsSynchronized() {
		n.overlay.SetFraction(fraction)
	}
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *DiffNode) Advance(seconds, synchronizationRate float64) {
	if n.primary.IsSynchronized() {
		n.primary.Advance(seconds, synchronizationRate)
	} else {
		n.primary.Advance(seconds, 1.0)
	}
	if n.overlay.IsSynchronized() {
		adjustedRate := n.primary.Rate() / n.overlay.Rate()
		n.overlay.Advance(seconds, synchronizationRate*adjustedRate)
	} else {
		n.overlay.Advance(seconds, 1.0)
	}
}

// IsSynchronized returns whether the node should be synchronized.
func (n *DiffNode) IsSynchronized() bool {
	return n.synchronized
}

// SetSynchronized configures whether the node should be synchronized.
func (n *DiffNode) SetSynchronized(synchronized bool) {
	n.synchronized = synchronized
}

// Synchronize is called each frame to allow a node to synchronized its
// children (depending on their setting).
//
// This will be called (and should be called on children) regardless if
// the current or any child node is synchronized or not.
func (n *DiffNode) Synchronize() {
	n.primary.Synchronize()
	n.overlay.Synchronize()
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
