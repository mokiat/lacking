package animation

import "github.com/mokiat/gomath/dprec"

// NewPlayback creates a new animation source using the specified
// Recording.
func NewPlayback(recording *Recording) *Playback {
	return &Playback{
		recording: recording,
		position:  recording.StartTime(),
		loop:      recording.loop,
	}
}

var _ Source = (*Playback)(nil)

// Playback represents an animation source that plays back an
// animation.
type Playback struct {
	recording *Recording
	position  float64
	loop      bool
}

// Loop returns whether the animation should loop.
func (p *Playback) Loop() bool {
	return p.loop
}

// SetLoop sets whether the animation should loop.
func (p *Playback) SetLoop(loop bool) *Playback {
	p.loop = loop
	p.SetPosition(p.Position()) // force clamp if needed
	return p
}

// Length returns the length of the animation in seconds.
func (p *Playback) Length() float64 {
	return p.recording.Length()
}

// Position returns the current position of the animation in seconds.
func (p *Playback) Position() float64 {
	return p.position
}

// SetPosition sets the current position of the animation in seconds.
func (p *Playback) SetPosition(position float64) {
	if p.loop {
		p.position = p.recording.StartTime() + dprec.Mod(position, p.recording.Length())
	} else {
		p.position = dprec.Clamp(position, p.recording.StartTime(), p.recording.EndTime())
	}
}

// NodeTransform returns the transformation of the node with the specified
// name at the current time position.
func (p *Playback) NodeTransform(name string) NodeTransform {
	return p.recording.BindingTransform(name, p.position)
}
