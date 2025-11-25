package animation

// NewOverlayNode creates an animation node that overlays one animation on
// top of another.
func NewOverlayNode(primary, overlay Node) *OverlayNode {
	return &OverlayNode{
		primary: primary,
		overlay: overlay,
	}
}

// OverlayNode is an animation node that overlays one animation on top
// of another.
type OverlayNode struct {
	primary Node
	overlay Node

	synchronized bool
}

var _ Node = (*OverlayNode)(nil)

// Rate returns the fraction of the animation length that advances each
// second (fraction per second).
func (n *OverlayNode) Rate() float64 {
	return n.primary.Rate()
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *OverlayNode) Fraction() float64 {
	return n.primary.Fraction()
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *OverlayNode) SetFraction(fraction float64) {
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
func (n *OverlayNode) Advance(seconds, synchronizationRate float64) {
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
func (n *OverlayNode) IsSynchronized() bool {
	return n.synchronized
}

// SetSynchronized configures whether the node should be synchronized.
func (n *OverlayNode) SetSynchronized(synchronized bool) {
	n.synchronized = synchronized
}

// Synchronize is called each frame to allow a node to synchronized its
// children (depending on their setting).
//
// This will be called (and should be called on children) regardless if
// the current or any child node is synchronized or not.
func (n *OverlayNode) Synchronize() {
	n.primary.Synchronize()
	n.overlay.Synchronize()
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *OverlayNode) BoneTransform(bone string) NodeTransform {
	originalTransform := n.primary.BoneTransform(bone)
	overlayTransform := n.overlay.BoneTransform(bone)
	return FirstNodeTransform(overlayTransform, originalTransform)
}

// BoneDeltaTransform returns the transformation that the bone will experience
// throughout the next delta interval. This is used for root motion.
func (n *OverlayNode) BoneDeltaTransform(bone string, delta float64) NodeTransform {
	originalTransform := n.primary.BoneDeltaTransform(bone, delta)
	overlayTransform := n.overlay.BoneDeltaTransform(bone, delta)
	return FirstNodeTransform(overlayTransform, originalTransform)
}
