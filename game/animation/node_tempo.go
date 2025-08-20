package animation

// NewTempoNode creates a new animation node that can adjust the speed of the
// underlying animation.
func NewTempoNode(delegate Node) *TempoNode {
	return &TempoNode{
		delegate: delegate,
		speed:    1.0,
	}
}

// TempoNode is a decorator for an animation source that allows
// adjusting the playback speed.
type TempoNode struct {
	delegate Node
	speed    float64
}

var _ Node = (*TempoNode)(nil)

// Speed returns the current playback speed of the animation.
// A value of 1.0 means that the animation is played at normal speed.
func (n *TempoNode) Speed() float64 {
	return n.speed
}

// SetSpeed sets the playback speed of the animation.
func (n *TempoNode) SetSpeed(speed float64) *TempoNode {
	n.speed = speed
	return n
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *TempoNode) Rate() float64 {
	return n.delegate.Rate() * n.speed
}

// Fraction returns the amount of animation that has elapsed. In case of
// looping, the value will wrap around.
//
// The returned value is in the range [0.0..1.0).
func (n *TempoNode) Fraction() float64 {
	return n.delegate.Fraction()
}

// SetFraction relocates the animation to the specified fractional position.
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *TempoNode) SetFraction(fraction float64) {
	n.delegate.SetFraction(fraction)
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *TempoNode) Advance(seconds, synchronizationRate float64) {
	n.delegate.Advance(seconds*n.speed, synchronizationRate)
}

// IsSynchronized returns whether the node should be synchronized.
func (n *TempoNode) IsSynchronized() bool {
	return n.delegate.IsSynchronized()
}

// SetSynchronized configures whether the node should be synchronized.
func (n *TempoNode) SetSynchronized(synchronized bool) {
	n.delegate.SetSynchronized(synchronized)
}

// Synchronize is called each frame to allow a node to synchronized its
// children (depending on their setting).
//
// This will be called (and should be called on children) regardless if
// the current or any child node is synchronized or not.
func (n *TempoNode) Synchronize() {
	n.delegate.Synchronize()
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *TempoNode) BoneTransform(bone string) NodeTransform {
	return n.delegate.BoneTransform(bone)
}

// BoneDeltaTransform returns the transformation that the bone will experience
// throughout the next delta interval. This is used for root motion.
func (n *TempoNode) BoneDeltaTransform(bone string, delta float64) NodeTransform {
	return n.delegate.BoneDeltaTransform(bone, delta*n.speed)
}

// WithSpeed creates a new node that adjusts the speed of the underlying node.
func WithSpeed(node Node, speed float64) *TempoNode {
	result := NewTempoNode(node)
	result.SetSpeed(speed)
	return result
}
