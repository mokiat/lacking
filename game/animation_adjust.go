package game

// NewAdjustedAnimation creates a new adjusted animation source with the
// specified delegate.
func NewAdjustedAnimation(delegate AnimationSource) *AdjustedAnimation {
	return &AdjustedAnimation{
		delegate: delegate,
		speed:    1.0,
	}
}

var _ AnimationSource = (*AdjustedAnimation)(nil)

// AdjustedAnimation is a decorator for an animation source that allows
// adjusting the playback speed.
type AdjustedAnimation struct {
	delegate AnimationSource
	speed    float64
}

// Source returns the underlying animation source.
func (a *AdjustedAnimation) Source() AnimationSource {
	return a.delegate
}

// Speed returns the current playback speed of the animation.
// A value of 1.0 means that the animation is played at normal speed.
func (a *AdjustedAnimation) Speed() float64 {
	return a.speed
}

// SetSpeed sets the playback speed of the animation.
func (a *AdjustedAnimation) SetSpeed(speed float64) {
	a.speed = speed
}

// Length returns the length of the animation in seconds.
func (a *AdjustedAnimation) Length() float64 {
	return a.delegate.Length() / a.speed
}

// Position returns the current position of the animation in seconds.
func (a *AdjustedAnimation) Position() float64 {
	return a.delegate.Position() / a.speed
}

// SetPosition sets the current position of the animation in seconds.
func (a *AdjustedAnimation) SetPosition(position float64) {
	a.delegate.SetPosition(position * a.speed)
}

// NodeTransform returns the transformation of the node with the
// specified name at the current time position.
func (a *AdjustedAnimation) NodeTransform(name string) NodeTransform {
	return a.delegate.NodeTransform(name)
}
