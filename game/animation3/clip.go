package animation3

import (
	"iter"
	"maps"
)

// Clip represents a pre-recorded animation: a set of [Channel] values, one
// per animated bone, covering the time range between the start and the end
// time.
//
// A Clip is immutable once created and can be shared by any number of
// consumers.
type Clip struct {
	name      string
	startTime float64
	endTime   float64
	channels  map[string]Channel
}

// NewClip creates a new [Clip] with the specified name, time range and
// channels keyed by bone name. The start and end times are keyframe
// timestamps.
//
// The Clip takes ownership of the channels map; the caller must not modify it
// afterwards. The channels themselves are immutable and can be shared with
// other clips.
func NewClip(name string, startTime, endTime float64, channels map[string]Channel) *Clip {
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

// Channel returns the channel for the bone with the specified name and
// whether such a channel exists.
func (c *Clip) Channel(name string) (Channel, bool) {
	channel, ok := c.channels[name]
	return channel, ok
}

// HasChannel returns whether the clip has a channel for the bone with the
// specified name.
func (c *Clip) HasChannel(name string) bool {
	_, ok := c.channels[name]
	return ok
}

// ChannelNamesIter returns an iterator over the names of all bones that have a
// channel in this clip. The iteration order is unspecified.
func (c *Clip) ChannelNamesIter() iter.Seq[string] {
	return maps.Keys(c.channels)
}

// Crop returns a new [Clip] with the specified name and time range and the
// same channels as this clip. The start and end times are keyframe
// timestamps, not offsets relative to this clip's start time. The channels
// are shared and not trimmed, so animation data outside the new range remains
// accessible through [Clip.Channel].
func (c *Clip) Crop(name string, startTime, endTime float64) *Clip {
	return NewClip(
		name,
		startTime,
		endTime,
		c.channels,
	)
}
