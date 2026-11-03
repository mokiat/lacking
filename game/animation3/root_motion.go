package animation3

// RootMotion extracts the motion of a single bone, typically the root bone
// of a character, from an animation tree. The extracted motion is meant to
// move the character itself, for example through a character controller, so
// that its movement in the world matches its animation.
type RootMotion struct {
	rootNode Node
	bone     int
}

// NewRootMotion creates a new [RootMotion] that extracts the motion of the
// bone with the specified name from the specified node, which is usually the
// root of an animation tree. The bone name is resolved against the specified
// skeleton once. If the skeleton has no bone with that name, the RootMotion
// extracts no motion.
func NewRootMotion(skeleton *Skeleton, node Node, boneName string) *RootMotion {
	bone, ok := skeleton.BoneIndex(boneName)
	if !ok {
		bone = -1
	}
	return &RootMotion{
		rootNode: node,
		bone:     bone,
	}
}

// DeltaTransform returns the motion that the bone experiences throughout the
// next delta seconds of playback, as reported by [Node.DeltaTransform], without
// changing the playback position. It is meant to be called before the node is
// advanced by the same delta. The result is expressed in the space of the
// bone's parent. If the bone was not found in the skeleton, the
// [IdentityTransform] is returned.
func (m *RootMotion) DeltaTransform(deltaSeconds float64) Transform {
	if m.bone == -1 {
		return IdentityTransform()
	}
	return m.rootNode.DeltaTransform(m.bone, deltaSeconds)
}
