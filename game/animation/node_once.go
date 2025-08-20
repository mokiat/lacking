package animation

import "github.com/mokiat/gomath/dprec"

// NewOnceNode creates a node that can play an animation once at a time,
// per user request.
func NewOnceNode(primary, overlay Node) *OnceNode {
	return &OnceNode{
		primary: primary,
		overlay: overlay,
		active:  false,
	}
}

// OnceNode allows an animation to be played just once.
type OnceNode struct {
	primary         Node
	overlay         Node
	remaining       float64
	fadeInFraction  float64
	fadeOutFraction float64
	active          bool
}

var _ Node = (*OnceNode)(nil)

// FadeInFraction returns the amount of time (in fraction of the total
// animation) that it takes to fade into the overlay animation.
func (n *OnceNode) FadeInFraction() float64 {
	return n.fadeInFraction
}

// SetFadeInFraction sets the amount of time (in fraction of the total
// animation) that it takes to fade into the overlay animation.
func (n *OnceNode) SetFadeInFraction(fraction float64) {
	n.fadeInFraction = fraction
}

// FadeOutFraction returns the amount of time (in fraction of the total
// animation) that it takes to fade out of the overlay animation.
func (n *OnceNode) FadeOutFraction() float64 {
	return n.fadeOutFraction
}

// SetFadeOutFraction sets the amount of time (in fraction of the total
// animation) that it takes to fade out of the overlay animation.
func (n *OnceNode) SetFadeOutFraction(fraction float64) {
	n.fadeOutFraction = fraction
}

// Trigger rewinds and activates the animation to be played once.
func (n *OnceNode) Trigger() *OnceNode {
	n.remaining = 1.0
	n.overlay.SetFraction(0.0)
	n.active = true
	return n
}

// Finished returns whether the action has completed.
func (n *OnceNode) Finished() bool {
	return !n.active
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *OnceNode) Rate() float64 {
	return n.primary.Rate()
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *OnceNode) Fraction() float64 {
	return n.primary.Fraction()
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *OnceNode) SetFraction(fraction float64) {
	n.primary.SetFraction(fraction)
	if n.overlay.IsSynchronized() {
		n.overlay.SetFraction(n.primary.Fraction())
	}
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *OnceNode) Advance(seconds, synchronizationRate float64) {
	if n.primary.IsSynchronized() {
		n.primary.Advance(seconds, synchronizationRate)
	} else {
		n.primary.Advance(seconds, 1.0)
	}

	if n.overlay.IsSynchronized() {
		adjustedRate := n.primary.Rate() / n.overlay.Rate()
		n.overlay.Advance(seconds, synchronizationRate*adjustedRate)
		n.remaining -= seconds * synchronizationRate * adjustedRate
	} else {
		n.overlay.Advance(seconds, 1.0)
		n.remaining -= seconds
	}
	if n.remaining < 0.0 {
		n.active = false
	}
}

// IsSynchronized returns whether the node should be synchronized.
func (n *OnceNode) IsSynchronized() bool {
	return n.primary.IsSynchronized()
}

// SetSynchronized configures whether the node should be synchronized.
func (n *OnceNode) SetSynchronized(synchronized bool) {
	n.primary.SetSynchronized(synchronized)
}

// Synchronize is called each frame to allow a node to synchronized its
// children (depending on their setting).
//
// This will be called (and should be called on children) regardless if
// the current or any child node is synchronized or not.
func (n *OnceNode) Synchronize() {
	n.primary.Synchronize()

	if n.overlay.IsSynchronized() {
		n.overlay.SetFraction(n.primary.Fraction())
	}
	n.overlay.Synchronize()
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *OnceNode) BoneTransform(bone string) NodeTransform {
	primaryTransform := n.primary.BoneTransform(bone)
	if !n.active {
		return primaryTransform
	}
	overlayTransform := n.overlay.BoneTransform(bone)
	return BlendNodeTransforms(primaryTransform, overlayTransform, n.blendFactor(1.0-n.remaining))
}

// BoneDeltaTransform returns the transformation that the bone will experience
// throughout the next delta interval. This is used for root motion.
func (n *OnceNode) BoneDeltaTransform(bone string, delta float64) NodeTransform {
	primaryTransform := n.primary.BoneDeltaTransform(bone, delta)
	if !n.active {
		return primaryTransform
	}
	overlayTransform := n.overlay.BoneDeltaTransform(bone, delta)
	return BlendNodeTransforms(primaryTransform, overlayTransform, n.blendFactor(1.0-n.remaining))
}

func (n *OnceNode) blendFactor(transitionFraction float64) float64 {
	transitionFraction = dprec.Clamp(transitionFraction, 0.0, 1.0)
	if transitionFraction < n.fadeInFraction {
		return transitionFraction / n.fadeInFraction
	}
	if transitionFraction > 1.0-n.fadeOutFraction {
		return (1.0 - transitionFraction) / n.fadeOutFraction
	}
	return 1.0
}
