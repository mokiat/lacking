package animation

import (
	"github.com/mokiat/gomath/dprec"
)

// ClipNode is a [Node] that plays back a single [Clip]. Playback covers the
// time range of the clip, with the clip's start time at fraction 0.0 and its
// end time at fraction 1.0. A looping node wraps around when it reaches
// either end, whereas a non-looping node stops there.
type ClipNode struct {
	skeleton         *Skeleton
	clip             *Clip
	channels         []Channel
	length           float64
	startTimestamp   float64
	currentTimestamp float64
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
		skeleton:         skeleton,
		clip:             clip,
		channels:         extractChannels(skeleton, clip),
		length:           clip.Length(),
		startTimestamp:   clip.StartTime(),
		currentTimestamp: 0.0,
		loop:             loop,
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
	return n.currentTimestamp / n.length
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
	n.currentTimestamp = fraction * n.length
}

// Advance moves playback by the specified delta seconds, where one second
// corresponds to one unit of keyframe time. A looping node wraps around past
// either end of the clip, whereas a non-looping node stops at the end it
// reaches. Advancing a clip with zero length has no effect.
func (n *ClipNode) Advance(deltaSeconds float64) {
	if n.length <= 0.0 {
		return
	}
	n.currentTimestamp += deltaSeconds
	if n.loop {
		n.currentTimestamp = wrapTime(n.currentTimestamp, n.length)
	} else {
		n.currentTimestamp = clampTime(n.currentTimestamp, n.length)
	}
}

// Transform returns the local transformation of the specified bone at the
// current playback position, as produced by the bone's [Channel] in the clip
// at the matching keyframe timestamp. A bone that has no channel in the clip
// returns its [Skeleton.BoneRestTransform].
//
// The bone must be a valid index in the skeleton of the node.
func (n *ClipNode) Transform(bone int) Transform {
	return n.transformAt(bone, n.currentTimestamp)
}

// DeltaTransform returns the motion that the specified bone experiences
// between the current playback position and the position delta seconds
// later, without changing the playback position. A negative delta measures
// the motion backwards in time. The returned transformation D is expressed in
// the space of the bone's parent and satisfies D * A = B, where A and B are
// the [Transform.Matrix] of the bone at the two positions.
//
// A looping node measures the motion across any number of wrap arounds as if
// each repetition of the clip continued from where the previous one ended. A
// non-looping node stops measuring at the end of the clip that it reaches. A
// clip with zero length, as well as a zero delta, produces the
// [IdentityTransform].
//
// The result is decomposed from a matrix and is approximate when the motion
// contains shear, which can happen when the bone's scale is non-uniform and
// changes between the two positions.
//
// The bone must be a valid index in the skeleton of the node.
func (n *ClipNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	if deltaSeconds == 0.0 || n.length <= 0.0 {
		return IdentityTransform()
	}

	from := n.currentTimestamp
	to := from + deltaSeconds

	if !n.loop {
		return Mat4Transform(n.segmentDelta(bone, from, to))
	}
	if deltaSeconds >= 0.0 {
		return Mat4Transform(n.forwardDelta(bone, from, to))
	} else {
		return Mat4Transform(n.reverseDelta(bone, from, to))
	}
}

func (n *ClipNode) forwardDelta(bone int, from, to float64) dprec.Mat4 {
	var loopDelta dprec.Mat4
	if to > n.length {
		loopDelta = n.segmentDelta(bone, 0.0, n.length)
	}

	result := dprec.IdentityMat4()
	for to > n.length {
		result = dprec.Mat4Prod(result, loopDelta)
		to -= n.length
	}
	return dprec.Mat4Prod(result, n.segmentDelta(bone, from, to))
}

func (n *ClipNode) reverseDelta(bone int, from, to float64) dprec.Mat4 {
	var loopDelta dprec.Mat4
	if to < 0.0 {
		loopDelta = n.segmentDelta(bone, n.length, 0.0)
	}

	result := dprec.IdentityMat4()
	for to < 0.0 {
		result = dprec.Mat4Prod(result, loopDelta)
		to += n.length
	}
	return dprec.Mat4Prod(result, n.segmentDelta(bone, from, to))
}

func (n *ClipNode) transformAt(bone int, timestamp float64) Transform {
	channel := n.channels[bone]

	if channel == nil {
		return n.skeleton.BoneRestTransform(bone)
	}

	return channel.Transform(n.startTimestamp + timestamp)
}

func (n *ClipNode) segmentDelta(bone int, from, to float64) dprec.Mat4 {
	from = clampTime(from, n.length)
	to = clampTime(to, n.length)

	fromMatrix := n.transformAt(bone, from).Matrix()
	toMatrix := n.transformAt(bone, to).Matrix()

	return dprec.Mat4Prod(toMatrix, dprec.InverseMat4(fromMatrix))
}

// extractChannels returns the channels of the clip indexed by skeleton bone.
// Bones that have no channel in the clip get a nil channel and channels for
// bones that are not in the skeleton are dropped.
func extractChannels(skeleton *Skeleton, clip *Clip) []Channel {
	channels := make([]Channel, skeleton.BoneCount())
	for i := range channels {
		if channel, ok := clip.Channel(skeleton.BoneName(i)); ok {
			channels[i] = channel
		}
	}
	return channels
}
