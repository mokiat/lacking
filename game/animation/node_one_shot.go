package animation

import (
	"github.com/mokiat/gomath/dprec"
)

// OneShotNode is a [Node] that plays a base animation and, each time it is
// triggered, plays an overlay animation over it exactly once, such as a
// recoil or a blink on top of locomotion. Both nodes are expected to be of
// the same kind (see [Node.Transform]), which is also the kind of the
// result.
//
// The overlay is blended over the base bone by bone, according to a mask
// that assigns a weight to each bone, in the same way as by a
// [MaskedBlendNode]. The overlay can fade in when it starts and fade out
// when it ends, both over durations in seconds.
//
// The base node drives the playback of the OneShotNode: its rate and
// fraction are those of the base node. The overlay node is advanced in real
// time, and only while it plays, so that it starts from its beginning each
// time it is triggered. The synchronization settings of both the base and
// the overlay node have no effect. Whether the base node is kept in phase
// with the parent of the OneShotNode is decided by the synchronization
// setting of the OneShotNode itself.
type OneShotNode struct {
	synchronizationAspect
	base        Node
	overlay     Node
	boneWeights []float64

	fadeInDuration  float64
	fadeOutDuration float64

	playing   bool
	elapsed   float64 // seconds since the overlay was triggered
	remaining float64 // fraction of the overlay that is left to play

	// stopping indicates that the overlay was stopped explicitly, in which
	// case stopWeight is the weight that remains while it fades out.
	stopping   bool
	stopWeight float64
}

var _ Node = (*OneShotNode)(nil)

// NewOneShotNode creates a new [OneShotNode] that plays the specified base
// node and, when triggered, the specified overlay node over it, according to
// the specified bone weights keyed by bone name. Bone weights are resolved
// against the specified skeleton: bones without an entry have a weight of
// 0.0 and show the base animation, and entries for bones that are not in
// the skeleton are ignored. If the bone weights are nil, the overlay
// applies to all bones with a weight of 1.0. The node does not retain the
// specified map.
//
// The node starts out without playing the overlay and without fades. The
// OneShotNode takes control of the playback of both nodes, which should not
// be advanced or relocated through other means afterwards. The node is not
// synchronized.
func NewOneShotNode(skeleton *Skeleton, base, overlay Node, boneWeights map[string]float64) *OneShotNode {
	weights := make([]float64, skeleton.BoneCount())
	for i := range weights {
		if boneWeights == nil {
			weights[i] = 1.0
		} else {
			weights[i] = boneWeights[skeleton.BoneName(i)]
		}
	}
	return &OneShotNode{
		base:        base,
		overlay:     overlay,
		boneWeights: weights,
	}
}

// FadeInDuration returns the time in seconds that it takes for the overlay
// to fade in after it is triggered.
func (n *OneShotNode) FadeInDuration() float64 {
	return n.fadeInDuration
}

// SetFadeInDuration changes the time in seconds that it takes for the
// overlay to fade in after it is triggered. A duration of 0.0, as well as a
// negative one, shows the overlay at full weight immediately.
func (n *OneShotNode) SetFadeInDuration(seconds float64) {
	n.fadeInDuration = max(seconds, 0.0)
}

// FadeOutDuration returns the time in seconds that it takes for the overlay
// to fade out before it ends or after it is stopped.
func (n *OneShotNode) FadeOutDuration() float64 {
	return n.fadeOutDuration
}

// SetFadeOutDuration changes the time in seconds that it takes for the
// overlay to fade out before it ends or after it is stopped. A duration of
// 0.0, as well as a negative one, removes the overlay immediately.
//
// If the overlay is shorter than the fade in and fade out durations
// combined, the two fades meet and the overlay does not reach full weight.
func (n *OneShotNode) SetFadeOutDuration(seconds float64) {
	n.fadeOutDuration = max(seconds, 0.0)
}

// IsPlaying returns whether the overlay is playing, including while it
// fades in or out.
func (n *OneShotNode) IsPlaying() bool {
	return n.playing
}

// Trigger starts playing the overlay from its beginning, fading it in. The
// overlay plays until one cycle of it has elapsed, as determined by its
// rate, or until it is stopped.
//
// Triggering has no effect while the overlay is playing, unless it was
// stopped and is fading out, in which case the overlay starts over from its
// beginning and fades back in from its current weight.
//
// The overlay must have a rate greater than 0.0, otherwise the method
// panics.
func (n *OneShotNode) Trigger() {
	if n.playing && !n.stopping {
		return
	}
	if n.overlay.Rate() <= 0.0 {
		panic("overlay must have a rate greater than 0.0")
	}
	// The fade in continues from the weight that the overlay has at the
	// moment, which is above 0.0 only while it fades out after a stop.
	n.elapsed = n.fadeWeight() * n.fadeInDuration
	n.overlay.SetFraction(0.0)
	n.remaining = 1.0
	n.playing = true
	n.stopping = false
}

// Stop fades out the overlay from its current weight over the fade out
// duration, after which it stops playing. Stopping has no effect when the
// overlay is not playing or is already stopped.
func (n *OneShotNode) Stop() {
	if !n.playing || n.stopping {
		return
	}
	n.stopWeight = n.fadeWeight()
	n.stopping = true
	if n.fadeOutDuration == 0.0 {
		n.playing = false
	}
}

// Rate returns the rate of the base node.
func (n *OneShotNode) Rate() float64 {
	return n.base.Rate()
}

// Fraction returns the fraction of the base node.
func (n *OneShotNode) Fraction() float64 {
	return n.base.Fraction()
}

// SetFraction relocates the base node to the specified fraction. The overlay
// keeps its playback position.
func (n *OneShotNode) SetFraction(fraction float64) {
	n.base.SetFraction(fraction)
}

// Advance moves the base node by the specified delta seconds, as well as the
// overlay node if it is playing. The overlay stops playing once one cycle of
// it has elapsed or once it has faded out after being stopped.
func (n *OneShotNode) Advance(deltaSeconds float64) {
	n.base.Advance(deltaSeconds)
	if !n.playing {
		return
	}
	n.overlay.Advance(deltaSeconds)
	n.elapsed = max(n.elapsed+deltaSeconds, 0.0)
	n.remaining -= deltaSeconds * n.overlay.Rate()
	if n.stopping && n.fadeOutDuration > 0.0 {
		n.stopWeight -= deltaSeconds / n.fadeOutDuration
	}
	if n.remaining <= 0.0 || (n.stopping && n.stopWeight <= 0.0) {
		n.playing = false
	}
}

// Transform returns the transformation of the specified bone, blended
// through [BlendTransforms] from that of the base node towards that of the
// overlay node by the effective weight of the bone, which is its mask weight
// scaled by the current fade of the overlay. While the overlay is not
// playing, this is the transformation of the base node.
//
// The bone must be a valid index in the skeleton of the node.
func (n *OneShotNode) Transform(bone int) Transform {
	switch weight := n.boneWeight(bone); weight {
	case 0.0:
		return n.base.Transform(bone)
	case 1.0:
		return n.overlay.Transform(bone)
	default:
		return BlendTransforms(n.base.Transform(bone), n.overlay.Transform(bone), weight)
	}
}

// DeltaTransform returns the motion that the specified bone experiences
// throughout the next delta seconds, blended through [BlendTransforms] from
// that of the base node towards that of the overlay node by the effective
// weight of the bone. A bone with an effective weight of 0.0 moves as in the
// base node, so a mask that excludes the root bone leaves root motion
// entirely to the base node.
//
// For an effective weight between 0.0 and 1.0 the result approximates the
// motion of the blended transformation. It does not account for the change
// in the fade of the overlay throughout the delta seconds.
//
// The bone must be a valid index in the skeleton of the node.
func (n *OneShotNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	switch weight := n.boneWeight(bone); weight {
	case 0.0:
		return n.base.DeltaTransform(bone, deltaSeconds)
	case 1.0:
		return n.overlay.DeltaTransform(bone, deltaSeconds)
	default:
		return BlendTransforms(n.base.DeltaTransform(bone, deltaSeconds), n.overlay.DeltaTransform(bone, deltaSeconds), weight)
	}
}

// boneWeight returns the effective weight of the specified bone.
func (n *OneShotNode) boneWeight(bone int) float64 {
	return n.boneWeights[bone] * n.fadeWeight()
}

// fadeWeight returns the weight of the overlay as a whole, according to its
// fade in, its fade out and whether it was stopped.
func (n *OneShotNode) fadeWeight() float64 {
	if !n.playing {
		return 0.0
	}
	weight := 1.0
	if n.fadeInDuration > 0.0 {
		weight = min(weight, n.elapsed/n.fadeInDuration)
	}
	if rate := n.overlay.Rate(); n.fadeOutDuration > 0.0 && rate > 0.0 {
		remainingSeconds := n.remaining / rate
		weight = min(weight, remainingSeconds/n.fadeOutDuration)
	}
	if n.stopping {
		weight = min(weight, n.stopWeight)
	}
	return dprec.Clamp(weight, 0.0, 1.0)
}
