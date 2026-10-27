package animation3

import (
	"github.com/mokiat/gomath/dprec"
)

// ClipNode is a [Node] that plays back a single [Clip]. Playback covers the
// time range of the clip, with the clip's start time at fraction 0.0 and its
// end time at fraction 1.0. A looping node wraps around when it reaches
// either end, whereas a non-looping node stops there.
type ClipNode struct {
	skeleton *Skeleton
	clip     *Clip
	channels []clipNodeChannel
	offset   float64
	length   float64
	position float64
	synchronizationAspect
	loop bool
}

var _ Node = (*ClipNode)(nil)

// NewClipNode creates a new [ClipNode] that plays back the specified clip on
// the bones of the specified skeleton, looping if loop is true. Clip channels
// are matched to skeleton bones by name: channels for bones that are not in
// the skeleton are ignored, and bones without a channel hold their rest
// transformation. The node starts at fraction 0.0 and is not synchronized.
func NewClipNode(skeleton *Skeleton, clip *Clip, loop bool) *ClipNode {
	return &ClipNode{
		skeleton: skeleton,
		clip:     clip,
		channels: extractChannels(skeleton, clip),
		offset:   clip.StartTime(),
		length:   clip.Length(),
		position: 0.0,
		loop:     loop,
	}
}

// Clip returns the clip that this node plays back.
func (n *ClipNode) Clip() *Clip {
	return n.clip
}

// Rate returns the reciprocal of the clip's [Clip.Length]. A clip with zero
// length has a rate of 0.0.
func (n *ClipNode) Rate() float64 {
	if n.length <= 0.0 {
		return 0.0
	}
	return 1.0 / n.length
}

// Fraction returns the playback position as a fraction of the clip's length.
// A clip with zero length always has a fraction of 0.0.
func (n *ClipNode) Fraction() float64 {
	if n.length <= 0.0 {
		return 0.0
	}
	return n.position / n.length
}

// SetFraction relocates playback to the specified fraction of the clip's
// length. A looping node wraps the fraction into the range [0.0..1.0),
// whereas a non-looping node clamps it into the range [0.0..1.0].
func (n *ClipNode) SetFraction(fraction float64) {
	if n.loop {
		fraction = wrapFraction(fraction)
	} else {
		fraction = clampFraction(fraction)
	}
	n.position = fraction * n.length
}

// Advance moves playback by the specified delta seconds, where one second
// corresponds to one unit of keyframe time. A looping node wraps around past
// either end of the clip, whereas a non-looping node stops at the end it
// reaches. Advancing a clip with zero length has no effect.
func (n *ClipNode) Advance(seconds float64) {
	if n.length <= 0.0 {
		return
	}
	n.position += seconds
	if n.loop {
		n.position = wrapTime(n.position, n.length)
	} else {
		n.position = dprec.Clamp(n.position, 0.0, n.length)
	}
}

// Transform returns the local transformation of the specified bone at the
// current playback position. Each component is interpolated between the two
// keyframes that surround the position: translation and scale linearly and
// rotation through normalized linear interpolation along the shorter arc.
// The latter does not keep a constant angular velocity between keyframes,
// which is only noticeable when keyframes are far apart in rotation. A
// position before the first or after the last keyframe of a component holds
// the value of that keyframe.
//
// Components that the clip does not animate for the bone are taken from the
// bone's [Skeleton.BoneRestTransform], so a bone that has no channel in the
// clip returns its rest transformation.
//
// The bone must be a valid index in the skeleton of the node.
func (n *ClipNode) Transform(bone int) Transform {
	channel := &n.channels[bone]

	hasTranslations := channel.HasTranslations()
	hasRotations := channel.HasRotations()
	hasScales := channel.HasScales()

	var result Transform
	if !hasTranslations || !hasRotations || !hasScales {
		result = n.skeleton.BoneRestTransform(bone)
	}
	if hasTranslations {
		result.Translation = channel.Translation(n.offset + n.position)
	}
	if hasRotations {
		result.Rotation = channel.Rotation(n.offset + n.position)
	}
	if hasScales {
		result.Scale = channel.Scale(n.offset + n.position)
	}
	return result
}

func (n *ClipNode) DeltaTransform(bone int, delta float64) Transform {
	panic("TODO")
}

// clipNodeChannel holds the keyframes that animate a single skeleton bone.
// A component with no keyframes is not animated.
type clipNodeChannel struct {
	translations KeyframeList[dprec.Vec3]
	rotations    KeyframeList[dprec.Quat]
	scales       KeyframeList[dprec.Vec3]
}

// extractChannels returns the channels of the clip indexed by skeleton bone.
// Bones that have no channel in the clip get an empty channel and channels
// for bones that are not in the skeleton are dropped.
func extractChannels(skeleton *Skeleton, clip *Clip) []clipNodeChannel {
	channels := make([]clipNodeChannel, skeleton.BoneCount())
	for i := range channels {
		if channel, ok := clip.Channel(skeleton.BoneName(i)); ok {
			channels[i] = clipNodeChannel{
				translations: channel.TranslationKeyframes,
				rotations:    channel.RotationKeyframes,
				scales:       channel.ScaleKeyframes,
			}
		}
	}
	return channels
}

// HasTranslations returns whether the channel animates the translation.
func (c *clipNodeChannel) HasTranslations() bool {
	return len(c.translations) > 0
}

// HasRotations returns whether the channel animates the rotation.
func (c *clipNodeChannel) HasRotations() bool {
	return len(c.rotations) > 0
}

// HasScales returns whether the channel animates the scale.
func (c *clipNodeChannel) HasScales() bool {
	return len(c.scales) > 0
}

// Translation returns the translation at the specified keyframe timestamp.
// The channel must have translation keyframes.
func (c *clipNodeChannel) Translation(timestamp float64) dprec.Vec3 {
	leftIndex, rightIndex := c.translations.binarySearch(timestamp)
	if leftIndex == rightIndex {
		return c.translations[leftIndex].Value
	}
	left := c.translations[leftIndex]
	right := c.translations[rightIndex]
	a := (timestamp - left.Timestamp) / (right.Timestamp - left.Timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, dprec.Clamp(a, 0.0, 1.0))
}

// Rotation returns the rotation at the specified keyframe timestamp.
// The channel must have rotation keyframes.
func (c *clipNodeChannel) Rotation(timestamp float64) dprec.Quat {
	leftIndex, rightIndex := c.rotations.binarySearch(timestamp)
	if leftIndex == rightIndex {
		return c.rotations[leftIndex].Value
	}
	left := c.rotations[leftIndex]
	right := c.rotations[rightIndex]
	a := (timestamp - left.Timestamp) / (right.Timestamp - left.Timestamp)
	return dprec.QuatNLerp(left.Value, right.Value, dprec.Clamp(a, 0.0, 1.0))
}

// Scale returns the scale at the specified keyframe timestamp.
// The channel must have scale keyframes.
func (c *clipNodeChannel) Scale(timestamp float64) dprec.Vec3 {
	leftIndex, rightIndex := c.scales.binarySearch(timestamp)
	if leftIndex == rightIndex {
		return c.scales[leftIndex].Value
	}
	left := c.scales[leftIndex]
	right := c.scales[rightIndex]
	a := (timestamp - left.Timestamp) / (right.Timestamp - left.Timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, dprec.Clamp(a, 0.0, 1.0))
}
