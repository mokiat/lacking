package animation

// NewAddNode creates a new node that returns the sum of two animations.
func NewAddNode(primary, secondary Node) *AddNode {
	return &AddNode{
		primary:   primary,
		secondary: secondary,
	}
}

// AddNode returns the sum of two animations.
type AddNode struct {
	primary   Node
	secondary Node
}

var _ Node = (*AddNode)(nil)

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *AddNode) Reset() {
	n.primary.Reset()
	n.secondary.Reset()
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *AddNode) Rate() float64 {
	return 1.0 // TODO: Figure this out.
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *AddNode) Fraction() float64 {
	return n.primary.Fraction() // TODO: Figure this out
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *AddNode) SetFraction(fraction float64) {
	n.primary.SetFraction(fraction)
	n.secondary.SetFraction(fraction)
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *AddNode) Advance(seconds, synchronizationRate float64) {
	n.primary.Advance(seconds, synchronizationRate)
	n.secondary.Advance(seconds, synchronizationRate)
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *AddNode) BoneTransform(bone string) NodeTransform {
	firstTransform := n.primary.BoneTransform(bone)
	secondTransform := n.primary.BoneTransform(bone)
	return AddNodeTransforms(firstTransform, secondTransform)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (n *AddNode) BoneTransformDelta(bone string) NodeTransform {
	firstTransform := n.primary.BoneTransformDelta(bone)
	secondTransform := n.secondary.BoneTransformDelta(bone)
	return AddNodeTransforms(firstTransform, secondTransform)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (n *AddNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	firstTransform := n.primary.BoneTransformInterpolation(bone, fraction)
	secondTransform := n.secondary.BoneTransformInterpolation(bone, fraction)
	return AddNodeTransforms(firstTransform, secondTransform)
}
