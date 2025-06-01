package animation

// NewAdjustedSource creates a new animation source with an adjusted
// playback speed.
func NewAdjustedSource(delegate Source) *AdjustedSource {
	return &AdjustedSource{
		delegate: delegate,
		speed:    1.0,
	}
}

var _ Source = (*AdjustedSource)(nil)

// AdjustedSource is a decorator for an animation source that allows
// adjusting the playback speed.
type AdjustedSource struct {
	delegate Source
	speed    float64
}

// Source returns the underlying animation source.
func (s *AdjustedSource) Source() Source {
	return s.delegate
}

// Speed returns the current playback speed of the animation.
// A value of 1.0 means that the animation is played at normal speed.
func (s *AdjustedSource) Speed() float64 {
	return s.speed
}

// SetSpeed sets the playback speed of the animation.
func (s *AdjustedSource) SetSpeed(speed float64) {
	s.speed = speed
}

// Length returns the length of the animation in seconds.
func (s *AdjustedSource) Length() float64 {
	return s.delegate.Length() / s.speed
}

// Position returns the current position of the animation in seconds.
func (s *AdjustedSource) Position() float64 {
	return s.delegate.Position() / s.speed
}

// SetPosition sets the current position of the animation in seconds.
func (s *AdjustedSource) SetPosition(position float64) {
	s.delegate.SetPosition(position * s.speed)
}

// NodeTransform returns the transformation of the node with the
// specified name at the current time position.
func (s *AdjustedSource) NodeTransform(name string) NodeTransform {
	return s.delegate.NodeTransform(name)
}

// WithSpeed creates a new adjusted animation source with the
// specified delegate and playback speed.
func WithSpeed(source Source, speed float64) Source {
	result := NewAdjustedSource(source)
	result.SetSpeed(speed)
	return result
}
