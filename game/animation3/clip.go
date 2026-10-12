package animation3

import (
	"iter"
	"maps"

	"github.com/mokiat/gomath/dprec"
)

// Clip represents a pre-recorded animation: a set of keyframe channels, one
// per animated bone, covering the time range between the start and the end
// time.
//
// A Clip is immutable once created and can be shared by any number of
// consumers.
type Clip struct {
	name      string
	startTime float64
	endTime   float64
	channels  map[string]ClipChannel
}

// NewClip creates a new [Clip] with the specified name, time range and
// keyframe channels keyed by bone name. The start and end times are keyframe
// timestamps.
//
// The Clip takes ownership of the channels map and the keyframe lists it
// references; the caller must not modify them afterwards.
func NewClip(name string, startTime, endTime float64, channels map[string]ClipChannel) *Clip {
	return &Clip{
		name:      name,
		startTime: startTime,
		endTime:   endTime,
		channels:  channels,
	}
}

// Name returns the name of the clip.
func (c *Clip) Name() string {
	return c.name
}

// StartTime returns the keyframe timestamp at which the clip starts.
func (c *Clip) StartTime() float64 {
	return c.startTime
}

// EndTime returns the keyframe timestamp at which the clip ends.
func (c *Clip) EndTime() float64 {
	return c.endTime
}

// Length returns the length of the clip, measured in the same units as the
// keyframe timestamps. A clip whose end time is before its start time has a
// length of 0.0.
func (c *Clip) Length() float64 {
	return max(0.0, c.endTime-c.startTime)
}

// Channel returns the keyframe channel for the bone with the specified name
// and whether such a channel exists. The keyframe lists of the returned
// channel are shared with the clip and must not be modified.
func (c *Clip) Channel(name string) (ClipChannel, bool) {
	channel, ok := c.channels[name]
	return channel, ok
}

// HasChannel returns whether the clip has a keyframe channel for the bone
// with the specified name.
func (c *Clip) HasChannel(name string) bool {
	_, ok := c.channels[name]
	return ok
}

// ChannelNamesIter returns an iterator over the names of all bones that have a
// keyframe channel in this clip. The iteration order is unspecified.
func (c *Clip) ChannelNamesIter() iter.Seq[string] {
	return maps.Keys(c.channels)
}

// Crop returns a new [Clip] with the specified name and time range and the
// same channels as this clip. The start and end times are keyframe
// timestamps, not offsets relative to this clip's start time. Keyframes are
// not trimmed, so those outside the new range remain accessible through
// [Clip.Channel].
func (c *Clip) Crop(name string, startTime, endTime float64) *Clip {
	return NewClip(
		name,
		startTime,
		endTime,
		c.channels,
	)
}

// ClipChannel holds the keyframes that animate a single bone of a [Clip].
// A component with no keyframes is not animated by the channel.
type ClipChannel struct {

	// TranslationKeyframes animate the translation of the bone.
	TranslationKeyframes KeyframeList[dprec.Vec3]

	// RotationKeyframes animate the rotation of the bone.
	RotationKeyframes KeyframeList[dprec.Quat]

	// ScaleKeyframes animate the scale of the bone.
	ScaleKeyframes KeyframeList[dprec.Vec3]
}
