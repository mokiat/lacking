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
func (s *TempoNode) Speed() float64 {
	return s.speed
}

// SetSpeed sets the playback speed of the animation.
func (s *TempoNode) SetSpeed(speed float64) *TempoNode {
	s.speed = speed
	return s
}

// Rate returns the fraction of the animation length that advances each
// second.
func (s *TempoNode) Rate() float64 {
	return s.delegate.Rate() * s.speed
}

// Reset clears any update delta information, so that new interpolations can
// be tracked.
func (s *TempoNode) Reset() {
	s.delegate.Reset()
}

// Progress returns the current fraction of the animation that has
// advanced since the start.
//
// This value will always be in the range [0.0..1.0).
func (s *TempoNode) Progress() float64 {
	return s.delegate.Progress()
}

// SetProgress changes the current position of the animation to the
// specified fraction.
//
// It is possible to set this value above 1.0, and in fact is necessary
// during update, so that it can handle loops and interpolation correctly,
// as setting the value directly to the wrapped-around value might indicate
// a reverse animation or a fractional animation.
//
// Internally, once applied, the progress will be normalized to [0.0..1.0).
func (s *TempoNode) SetProgress(fraction float64) {
	s.delegate.SetProgress(fraction)
}

// BoneTransform returns the transformation of the specified bone. Keep in
// mind that this is after a fixed interval update has been applied. If
// this is called from within a dynamic update handler, the
// BoneTransformInterpolation method should be used instead.
func (s *TempoNode) BoneTransform(bone string) NodeTransform {
	return s.delegate.BoneTransform(bone)
}

// BoneTransformDelta returns the transformation that was applied to the
// specified bone since the last reset.
func (s *TempoNode) BoneTransformDelta(bone string) NodeTransform {
	return s.delegate.BoneTransformDelta(bone)
}

// BoneTransformInterpolation returns the transformation of the specified bone
// at the specified interpolation fraction.
func (s *TempoNode) BoneTransformInterpolation(bone string, fraction float64) NodeTransform {
	return s.delegate.BoneTransformInterpolation(bone, fraction)
}

// WithSpeed creates a new node that adjusts the speed of the underlying node.
func WithSpeed(node Node, speed float64) *TempoNode {
	result := NewTempoNode(node)
	result.SetSpeed(speed)
	return result
}
