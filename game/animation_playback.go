package game

import "github.com/mokiat/gomath/dprec"

// NewAnimationPlayback creates a new AnimationSource using the specified
// Animation.
func NewAnimationPlayback(animation *Animation) *AnimationPlayback {
	return &AnimationPlayback{
		animation: animation,
		position:  animation.StartTime(),
		loop:      animation.loop,
	}
}

var _ AnimationSource = (*AnimationPlayback)(nil)

// AnimationPlayback represents an animation source that plays back an
// animation.
type AnimationPlayback struct {
	animation *Animation
	position  float64
	loop      bool
}

// Loop returns whether the animation should loop.
func (p *AnimationPlayback) Loop() bool {
	return p.loop
}

// SetLoop sets whether the animation should loop.
func (p *AnimationPlayback) SetLoop(loop bool) *AnimationPlayback {
	p.loop = loop
	p.SetPosition(p.Position()) // force clamp if needed
	return p
}

// Length returns the length of the animation in seconds.
func (p *AnimationPlayback) Length() float64 {
	return p.animation.Length()
}

// Position returns the current position of the animation in seconds.
func (p *AnimationPlayback) Position() float64 {
	return p.position
}

// SetPosition sets the current position of the animation in seconds.
func (p *AnimationPlayback) SetPosition(position float64) {
	if p.loop {
		p.position = p.animation.StartTime() + dprec.Mod(position, p.animation.Length())
	} else {
		p.position = dprec.Clamp(position, p.animation.StartTime(), p.animation.EndTime())
	}
}

// NodeTransform returns the transformation of the node with the specified
// name at the current time position.
func (p *AnimationPlayback) NodeTransform(name string) NodeTransform {
	return p.animation.BindingTransform(name, p.position)
}
