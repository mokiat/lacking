package animation3

// AdditivePoseNode is an additive [Node] (see [Node.Transform]) that produces
// a static set of bone differences that are specified at construction and
// cannot be changed afterwards. It is meant to be applied on top of an
// absolute node, for example through an [AddNode], in order to offset a few
// bones by a fixed amount, such as bending the spine to aim up or down.
//
// The node has no playback of its own: its rate and fraction are always 0.0,
// and advancing or relocating it has no effect. It can be configured as
// synchronized, though this has no effect on the pose that it produces.
type AdditivePoseNode struct {
	synchronizationAspect
	transforms []Transform
}

var _ Node = (*AdditivePoseNode)(nil)

// NewAdditivePoseNode creates a new [AdditivePoseNode] that holds the
// specified bone differences, keyed by bone name, for the bones of the
// specified skeleton. Bones without an entry hold the [IdentityTransform],
// so they are left unchanged when the node is applied, and entries for bones
// that are not in the skeleton are ignored. The node does not retain the
// specified map.
func NewAdditivePoseNode(skeleton *Skeleton, boneTransforms map[string]Transform) *AdditivePoseNode {
	transforms := make([]Transform, skeleton.BoneCount())
	for i := range transforms {
		if t, ok := boneTransforms[skeleton.BoneName(i)]; ok {
			transforms[i] = t
		} else {
			transforms[i] = IdentityTransform()
		}
	}
	return &AdditivePoseNode{
		transforms: transforms,
	}
}

// Rate returns 0.0, since a static pose has no meaningful length.
func (n *AdditivePoseNode) Rate() float64 {
	return 0.0 // motionless node
}

// Fraction returns 0.0, since a static pose has no playback position.
func (n *AdditivePoseNode) Fraction() float64 {
	return 0.0
}

// SetFraction has no effect, since a static pose has no playback position.
func (n *AdditivePoseNode) SetFraction(fraction float64) {
	// nothing to change
}

// Advance has no effect, since a static pose does not move.
func (n *AdditivePoseNode) Advance(deltaSeconds float64) {
	// can't advance this node
}

// Transform returns the difference of the specified bone that was configured
// at construction, or the [IdentityTransform] if none was.
//
// The bone must be a valid index in the skeleton of the node.
func (n *AdditivePoseNode) Transform(bone int) Transform {
	return n.transforms[bone]
}

// DeltaTransform returns the [IdentityTransform], since a static pose
// experiences no motion.
func (n *AdditivePoseNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	return IdentityTransform()
}
