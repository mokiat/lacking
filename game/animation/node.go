package animation

// Node represents an animation logic.
type Node interface {

	// Reset clears any update delta information, so that new interpolations can
	// be tracked.
	Reset()

	// Rate returns the fraction of the animation length that advances each
	// second (fraction per second).
	Rate() float64

	// Seek relocates the animation to the specified position (fractional).
	//
	// NOTE: This resets the animation and accumulated delta is lost.
	Seek(fraction float64)

	// Advance moves the animation forward by the specified delta seconds.
	//
	// The synchronizationRate determines the amount of scaling on the seconds
	// that should be applied in order to be correctly synchronized with sibling
	// and parent nodes in case of synchronization.
	Advance(seconds, synchronizationRate float64)

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
