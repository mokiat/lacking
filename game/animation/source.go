package animation

// Source represents a source of animation data.
type Source interface {

	// Length returns the length of the animation in seconds.
	Length() float64

	// BoneTransformDelta returns the transformation that occurred for the
	// specified bone between the from and to animation timestamps .
	//
	// This is mostly used for root motion.
	BoneTransformDelta(bone string, fromTimestamp, toTimestamp float64) NodeTransform

	// BoneTransformAt returns the transformation for the specified bone
	// at the specified animation timestamp.
	BoneTransformAt(bone string, timestamp float64) NodeTransform
}
