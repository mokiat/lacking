package animation

// ClipPoseNode is a [Node] that produces the static pose that a [Clip] has at
// a specific timestamp. The timestamp is relative to the clip's
// [Clip.StartTime] and can be changed at any time through
// [ClipPoseNode.SetTimestamp].
//
// The node has no playback of its own: its rate and fraction are always 0.0,
// and advancing or relocating it has no effect. It can be configured as
// synchronized, though this has no effect on the pose that it produces.
type ClipPoseNode struct {
	synchronizationAspect
	skeleton         *Skeleton
	clip             *Clip
	channels         []Channel
	length           float64
	startTimestamp   float64
	currentTimestamp float64
}

var _ Node = (*ClipPoseNode)(nil)

// NewClipPoseNode creates a new [ClipPoseNode] that samples the specified
// clip at the specified timestamp on the bones of the specified skeleton.
// The timestamp is handled as by [ClipPoseNode.SetTimestamp]. Clip channels
// are matched to skeleton bones by name: channels for bones that are not in
// the skeleton are ignored, and bones without a channel hold their rest
// transformation.
func NewClipPoseNode(skeleton *Skeleton, clip *Clip, timestamp float64) *ClipPoseNode {
	return &ClipPoseNode{
		skeleton:         skeleton,
		clip:             clip,
		channels:         extractChannels(skeleton, clip),
		length:           clip.Length(),
		startTimestamp:   clip.StartTime(),
		currentTimestamp: clampTime(timestamp, clip.Length()),
	}
}

// Clip returns the clip that this node samples.
func (n *ClipPoseNode) Clip() *Clip {
	return n.clip
}

// Timestamp returns the position within the clip at which the pose is
// sampled, measured from the clip's [Clip.StartTime].
func (n *ClipPoseNode) Timestamp() float64 {
	return n.currentTimestamp
}

// SetTimestamp changes the position within the clip at which the pose is
// sampled, measured from the clip's [Clip.StartTime]. The timestamp is
// clamped into the range [0.0..length], where length is the clip's
// [Clip.Length].
func (n *ClipPoseNode) SetTimestamp(timestamp float64) {
	n.currentTimestamp = clampTime(timestamp, n.length)
}

// Rate returns 0.0, since a static pose has no meaningful length.
func (n *ClipPoseNode) Rate() float64 {
	return 0.0 // motionless
}

// Fraction returns 0.0, since a static pose has no playback position.
func (n *ClipPoseNode) Fraction() float64 {
	return 0.0
}

// SetFraction has no effect, since a static pose has no playback position.
// Use [ClipPoseNode.SetTimestamp] to change the sampled pose.
func (n *ClipPoseNode) SetFraction(fraction float64) {
	// nothing to change
}

// Advance has no effect, since a static pose does not move.
func (n *ClipPoseNode) Advance(deltaSeconds float64) {
	// can't advance this node
}

// Transform returns the local transformation of the specified bone at the
// current timestamp, as produced by the bone's [Channel] in the clip. A bone
// that has no channel in the clip returns its [Skeleton.BoneRestTransform].
//
// The bone must be a valid index in the skeleton of the node.
func (n *ClipPoseNode) Transform(bone int) Transform {
	channel := n.channels[bone]

	if channel == nil {
		return n.skeleton.BoneRestTransform(bone)
	}

	return channel.Transform(n.startTimestamp + n.currentTimestamp)
}

// DeltaTransform returns the [IdentityTransform], since a static pose
// experiences no motion.
func (n *ClipPoseNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	return IdentityTransform()
}
