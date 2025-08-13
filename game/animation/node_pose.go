package animation

func NewPoseNode(source Source, timestamp float64) *PoseNode {
	return &PoseNode{
		source:    source,
		timestamp: timestamp,
	}
}

type PoseNode struct {
	source    Source
	timestamp float64
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *PoseNode) Reset() {}

// Rate returns the fraction of the animation length that advances each
// second (fraction per second).
func (n *PoseNode) Rate() float64 {
	return 1.0
}

// Seek relocates the animation to the specified position (fractional).
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *PoseNode) Seek(fraction float64) {}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *PoseNode) Advance(seconds, synchronizationRate float64) {}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *PoseNode) BoneTransform(bone string) NodeTransform {
	return n.source.BoneTransformAt(bone, n.timestamp)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (n *PoseNode) BoneTransformDelta(bone string) NodeTransform {
	return NodeTransform{}
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (n *PoseNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	return n.source.BoneTransformAt(bone, n.timestamp)
}
