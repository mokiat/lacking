package animation3

// ReverseNode is a [Node] that plays back another node in reverse. Its
// fraction mirrors that of the delegate node, so fraction 0.0 of a
// ReverseNode shows the end of the delegate and fraction 1.0 shows its start,
// and advancing it moves the delegate backwards in time.
//
// A ReverseNode is a transparent wrapper: it holds no playback state of its
// own and the synchronization setting is that of the delegate. Whether
// playback wraps around or stops at either end is decided by the delegate.
type ReverseNode struct {
	delegate Node
}

var _ Node = (*ReverseNode)(nil)

// NewReverseNode creates a new [ReverseNode] that plays back the specified
// delegate node in reverse. The ReverseNode takes control of the playback
// position of the delegate, which should not be advanced or relocated
// through other means afterwards.
func NewReverseNode(delegate Node) *ReverseNode {
	return &ReverseNode{
		delegate: delegate,
	}
}

// Reverse is a shorthand for [NewReverseNode], which is useful during node
// tree construction.
func Reverse(node Node) *ReverseNode {
	return NewReverseNode(node)
}

// Rate returns the rate of the delegate node. Playing a node in reverse does
// not change how fast it plays.
func (n *ReverseNode) Rate() float64 {
	return n.delegate.Rate()
}

// Fraction returns the playback position mirrored from the fraction of the
// delegate node, that is 1.0 minus the delegate fraction.
func (n *ReverseNode) Fraction() float64 {
	return 1.0 - n.delegate.Fraction()
}

// SetFraction relocates playback to the specified fraction, which places the
// delegate node at 1.0 minus that fraction. A fraction outside the range
// [0.0..1.0] is wrapped or clamped by the delegate, as it would be if the
// delegate were played forward.
func (n *ReverseNode) SetFraction(fraction float64) {
	n.delegate.SetFraction(1.0 - fraction)
}

// Advance moves playback forward by the specified delta seconds, which moves
// the delegate node backwards by the same amount. A negative delta moves the
// delegate forward.
func (n *ReverseNode) Advance(deltaSeconds float64) {
	n.delegate.Advance(-deltaSeconds)
}

// IsSynchronized returns whether the delegate node is synchronized.
func (n *ReverseNode) IsSynchronized() bool {
	return n.delegate.IsSynchronized()
}

// SetSynchronized configures whether the delegate node is synchronized.
func (n *ReverseNode) SetSynchronized(synchronized bool) {
	n.delegate.SetSynchronized(synchronized)
}

// Transform returns the transformation that the delegate node produces for
// the specified bone at its current playback position.
func (n *ReverseNode) Transform(bone int) Transform {
	return n.delegate.Transform(bone)
}

// DeltaTransform returns the motion that the specified bone experiences
// throughout the next delta seconds of reverse playback, which is the motion
// that the delegate node produces when measured delta seconds backwards in
// time.
func (n *ReverseNode) DeltaTransform(bone int, deltaSeconds float64) Transform {
	return n.delegate.DeltaTransform(bone, -deltaSeconds)
}
