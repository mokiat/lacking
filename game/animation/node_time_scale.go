package animation

import "math"

// TimeScaleNode is a [Node] that plays back another node with its time
// scaled by a multiplier. A time scale of 2.0 plays the delegate twice as
// fast, 0.5 plays it at half speed and 0.0 freezes it at its current
// position. A negative time scale plays the delegate backwards.
//
// A TimeScaleNode is a transparent wrapper: it holds no playback position of
// its own, so its fraction is that of the delegate, and the synchronization
// setting is that of the delegate. Whether playback wraps around or stops at
// either end is decided by the delegate.
type TimeScaleNode struct {
	delegate  Node
	timeScale float64
}

var _ Node = (*TimeScaleNode)(nil)

// NewTimeScaleNode creates a new [TimeScaleNode] that plays back the
// specified delegate node with a time scale of 1.0, which leaves its
// playback unchanged. The TimeScaleNode takes control of advancing the
// delegate, which should not be advanced through other means afterwards.
func NewTimeScaleNode(delegate Node) *TimeScaleNode {
	return &TimeScaleNode{
		delegate:  delegate,
		timeScale: 1.0,
	}
}

// ScaleTime is a shorthand for creating a [TimeScaleNode] through
// [NewTimeScaleNode] and configuring it with the specified time scale, which
// is useful during node tree construction.
func ScaleTime(node Node, timeScale float64) *TimeScaleNode {
	result := NewTimeScaleNode(node)
	result.SetTimeScale(timeScale)
	return result
}

// TimeScale returns the multiplier that is applied to time before it reaches
// the delegate node.
func (n *TimeScaleNode) TimeScale() float64 {
	return n.timeScale
}

// SetTimeScale changes the multiplier that is applied to time before it
// reaches the delegate node. See [TimeScaleNode] for the meaning of the
// different values.
func (n *TimeScaleNode) SetTimeScale(timeScale float64) {
	n.timeScale = timeScale
}

// Rate returns the rate of the delegate node multiplied by the absolute value
// of the time scale. The direction of playback does not affect the rate, so a
// negative time scale produces the same rate as its positive counterpart. A
// time scale of 0.0 produces a rate of 0.0, since playback does not progress.
func (n *TimeScaleNode) Rate() float64 {
	return n.delegate.Rate() * math.Abs(n.timeScale)
}

// Fraction returns the playback position of the delegate node.
func (n *TimeScaleNode) Fraction() float64 {
	return n.delegate.Fraction()
}

// SetFraction relocates the delegate node to the specified fraction. The time
// scale has no effect on this operation.
func (n *TimeScaleNode) SetFraction(fraction float64) {
	n.delegate.SetFraction(fraction)
}

// Advance moves the delegate node by the specified delta seconds multiplied
// by the time scale.
func (n *TimeScaleNode) Advance(deltaSeconds float64) {
	n.delegate.Advance(deltaSeconds * n.timeScale)
}

// IsSynchronized returns whether the delegate node is synchronized.
func (n *TimeScaleNode) IsSynchronized() bool {
	return n.delegate.IsSynchronized()
}

// SetSynchronized configures whether the delegate node is synchronized.
func (n *TimeScaleNode) SetSynchronized(synchronized bool) {
	n.delegate.SetSynchronized(synchronized)
}

// Transform returns the transformation that the delegate node produces for
// the specified bone at its current playback position.
func (n *TimeScaleNode) Transform(bone int) Transform {
	return n.delegate.Transform(bone)
}

// DeltaTransform returns the motion that the specified bone experiences
// throughout the next delta seconds of scaled playback, which is the motion
// that the delegate node produces throughout delta seconds multiplied by the
// time scale.
func (n *TimeScaleNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	return n.delegate.DeltaTransform(bone, deltaSeconds*n.timeScale)
}
