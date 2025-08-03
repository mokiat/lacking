package animation

// Node represents an animation logic.
type Node interface {

	// Rate returns the fraction of the animation length that advances each
	// second.
	Rate() float64

	// Reset clears any update delta information, so that new interpolations can
	// be tracked.
	Reset()

	// Progress returns the current fraction of the animation that has
	// advanced since the start.
	//
	// This value will always be in the range [0.0..1.0).
	Progress() float64

	// SetProgress changes the current position of the animation to the
	// specified fraction.
	//
	// It is possible to set this value above 1.0, and in fact is necessary
	// during update, so that it can handle loops and interpolation correctly,
	// as setting the value directly to the wrapped-around value might indicate
	// a reverse animation or a fractional animation.
	//
	// Internally, once applied, the progress will be normalized to [0.0..1.0).
	SetProgress(fraction float64)

	// BoneTransform returns the transformation of the specified bone. Keep in
	// mind that this is after a fixed interval update has been applied. If
	// this is called from within a dynamic update handler, the
	// BoneTransformInterpolation method should be used instead.
	BoneTransform(bone string) NodeTransform

	// BoneTransformDelta returns the transformation that was applied to the
	// specified bone since the last reset.
	BoneTransformDelta(bone string) NodeTransform

	// BoneTransformInterpolation returns the transformation of the specified bone
	// at the specified interpolation fraction.
	BoneTransformInterpolation(bone string, fraction float64) NodeTransform
}

// AdvanceNode is a helper function that moves the progress of a given node
// with the specified delta fraction.
func AdvanceNode(node Node, delta float64) {
	node.SetProgress(node.Progress() + delta)
}
