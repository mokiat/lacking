package animation

// NewPoseNode creates an animation node that produces a single static
// bone pose.
func NewPoseNode(source Source, timestamp float64) *PoseNode {
	return &PoseNode{
		source:    source,
		timestamp: timestamp,
	}
}

// PoseNode represents an animation node that produces a single static
// bone pose.
type PoseNode struct {
	source    Source
	timestamp float64
}

var _ Node = (*PoseNode)(nil)

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *PoseNode) Reset() {}

// Rate returns the fraction of the animation length that advances each
// second (fraction per second).
func (n *PoseNode) Rate() float64 {
	return 1.0
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *PoseNode) Fraction() float64 {
	return 0.0
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *PoseNode) SetFraction(fraction float64) {}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *PoseNode) Advance(seconds, synchronizationRate float64) {}

// IsSynchronized returns whether the node should be synchronized.
func (n *PoseNode) IsSynchronized() bool {
	return false
}

// SetSynchronized configures whether the node should be synchronized.
func (n *PoseNode) SetSynchronized(synchronized bool) {
	panic("pose node cannot be synchronized")
}

// Synchronize is called each frame to allow a node to synchronized its
// children (depending on their setting).
//
// This will be called (and should be called on children) regardless if
// the current or any child node is synchronized or not.
func (n *PoseNode) Synchronize() {}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *PoseNode) BoneTransform(bone string) NodeTransform {
	return n.source.BoneTransformAt(bone, n.timestamp)
}

// BoneDeltaTransform returns the transformation that the bone will experience
// throughout the next delta interval. This is used for root motion.
func (n *PoseNode) BoneDeltaTransform(bone string, delta float64) NodeTransform {
	return NodeTransform{}
}
