package animation

import (
	"iter"
	"maps"

	"github.com/mokiat/gog/opt"
)

// NewRecording creates a new empty recording.
func NewRecording() *Recording {
	return &Recording{
		bindings: make(map[string]KeyframeSet),
	}
}

// Recording represents a pre-stored set of keyframmes that can be played back.
type Recording struct {
	name      string
	startTime float64
	endTime   float64
	loop      bool
	bindings  map[string]KeyframeSet
}

// Name returns the name of the recording.
func (r *Recording) Name() string {
	return r.name
}

// SetName sets the name of the recording.
func (r *Recording) SetName(name string) {
	r.name = name
}

// StartTime returns the time (in seconds) at which the recording starts.
func (r *Recording) StartTime() float64 {
	return r.startTime
}

// SetStartTime sets the time (in seconds) at which the recording starts.
func (r *Recording) SetStartTime(startTime float64) {
	r.startTime = startTime
}

// EndTime returns the time (in seconds) at which the recording ends.
func (r *Recording) EndTime() float64 {
	return r.endTime
}

// SetEndTime sets the time (in seconds) at which the recording ends.
func (r *Recording) SetEndTime(endTime float64) {
	r.endTime = endTime
}

// Loop returns whether the recording should be looped.
func (r *Recording) Loop() bool {
	return r.loop
}

// SetLoop sets whether the recording should be looped.
func (r *Recording) SetLoop(loop bool) {
	r.loop = loop
}

// Length returns the length of the recording in seconds.
func (r *Recording) Length() float64 {
	return r.EndTime() - r.StartTime()
}

func (r *Recording) Binding(name string) (KeyframeSet, bool) {
	if binding, ok := r.bindings[name]; ok {
		return binding, true
	}
	return KeyframeSet{}, false
}

// SetBinding sets the keyframes for the node with the specified name.
func (r *Recording) SetBinding(name string, keyframes KeyframeSet) {
	r.bindings[name] = keyframes
}

// RemoveBinding removes the binding for the node with the specified name.
func (r *Recording) RemoveBinding(name string) {
	delete(r.bindings, name)
}

// BoundNodes returns the names of all nodes that have keyframes in this
// recording.
func (r *Recording) BoundNodes() iter.Seq[string] {
	return maps.Keys(r.bindings)
}

// BindingTransform returns the transformation of the node with the specified
// name at the specified time position.
func (r *Recording) BindingTransform(name string, timestamp float64) NodeTransform {
	var result NodeTransform
	binding, ok := r.bindings[name]
	if !ok {
		return result
	}
	if len(binding.TranslationKeyframes) > 0 {
		result.Translation = opt.V(binding.Translation(timestamp))
	}
	if len(binding.RotationKeyframes) > 0 {
		result.Rotation = opt.V(binding.Rotation(timestamp))
	}
	if len(binding.ScaleKeyframes) > 0 {
		result.Scale = opt.V(binding.Scale(timestamp))
	}
	return result
}

// Playback creates a new animtion source that plays back the animation.
func (r *Recording) Playback() *Playback {
	return NewPlayback(r)
}
