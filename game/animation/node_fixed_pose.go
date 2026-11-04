package animation

// FixedPoseNode is a [Node] that produces a static pose made of bone
// transformations that are specified at construction and cannot be changed
// afterwards.
//
// The node has no playback of its own: its rate and fraction are always 0.0,
// and advancing or relocating it has no effect. It can be configured as
// synchronized, though this has no effect on the pose that it produces.
type FixedPoseNode struct {
	synchronizationAspect
	transforms []Transform
}

var _ Node = (*FixedPoseNode)(nil)

// NewFixedPoseNode creates a new [FixedPoseNode] that holds the specified
// bone transformations, keyed by bone name, for the bones of the specified
// skeleton. Bones without an entry hold their
// [Skeleton.BoneRestTransform], and entries for bones that are not in the
// skeleton are ignored. The node does not retain the specified map.
func NewFixedPoseNode(skeleton *Skeleton, boneTransforms map[string]Transform) *FixedPoseNode {
	transforms := make([]Transform, skeleton.BoneCount())
	for i := range transforms {
		if t, ok := boneTransforms[skeleton.BoneName(i)]; ok {
			transforms[i] = t
		} else {
			transforms[i] = skeleton.BoneRestTransform(i)
		}
	}
	return &FixedPoseNode{
		transforms: transforms,
	}
}

// Rate returns 0.0, since a static pose has no meaningful length.
func (n *FixedPoseNode) Rate() float64 {
	return 0.0 // motionless node
}

// Fraction returns 0.0, since a static pose has no playback position.
func (n *FixedPoseNode) Fraction() float64 {
	return 0.0
}

// SetFraction has no effect, since a static pose has no playback position.
func (n *FixedPoseNode) SetFraction(fraction float64) {
	// nothing to change
}

// Advance has no effect, since a static pose does not move.
func (n *FixedPoseNode) Advance(deltaSeconds float64) {
	// can't advance this node
}

// Transform returns the transformation of the specified bone that was
// configured at construction, or the bone's [Skeleton.BoneRestTransform] if
// none was.
//
// The bone must be a valid index in the skeleton of the node.
func (n *FixedPoseNode) Transform(bone int) Transform {
	return n.transforms[bone]
}

// DeltaTransform returns the [IdentityTransform], since a static pose
// experiences no motion.
func (n *FixedPoseNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	return IdentityTransform()
}
