package animation

// NewTempoNode creates a new animation node that can adjust the speed of the
// underlying animation.
func NewTempoNode(delegate Node) *TempoNode {
	return &TempoNode{
		delegate: delegate,
		speed:    1.0,
	}
}

var _ Node = (*TempoNode)(nil)

// TempoNode is a decorator for an animation source that allows
// adjusting the playback speed.
type TempoNode struct {
	delegate Node
	speed    float64
}

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

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (n *TempoNode) Reset() {
	n.delegate.Reset()
}

// Rate returns the fraction of the animation length that advances each
// second.
func (n *TempoNode) Rate() float64 {
	return n.delegate.Rate() * n.speed
}

// Seek relocates the animation to the specified position (fractional).
//
// NOTE: This resets the animation and accumulated delta is lost.
func (n *TempoNode) Seek(fraction float64) {
	n.delegate.Seek(fraction)
}

// Advance moves the animation forward by the specified delta seconds.
//
// The synchronizationRate determines the amount of scaling on the seconds
// that should be applied in order to be correctly synchronized with sibling
// and parent nodes in case of synchronization.
func (n *TempoNode) Advance(seconds, synchronizationRate float64) {
	n.delegate.Advance(seconds*n.speed, synchronizationRate)
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (n *TempoNode) BoneTransform(bone string) NodeTransform {
	return n.delegate.BoneTransform(bone)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (n *TempoNode) BoneTransformDelta(bone string) NodeTransform {
	return n.delegate.BoneTransformDelta(bone)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (n *TempoNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	return n.delegate.BoneTransformInterpolation(bone, fraction)
}

// WithSpeed creates a new node that adjusts the speed of the underlying node.
func WithSpeed(node Node, speed float64) *TempoNode {
	result := NewTempoNode(node)
	result.SetSpeed(speed)
	return result
}
