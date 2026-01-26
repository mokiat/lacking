package audio

import "github.com/mokiat/gomath/sprec"

// Node represents a node in a chain of audio elements. Each node produces
// audio data which can be synthesized, processed, or played back.
type Node any

// UserNode represents an audio node that requires explicit resource management.
type UserNode interface {
	Node

	// Delete releases any resources associated with the node. After calling
	// this method, the node should not be used anymore as it may be reused
	// by the audio system.
	Delete()
}

// PlaybackNode represents an audio node that plays back audio data from a
// Media source.
type PlaybackNode interface {
	UserNode

	// Start starts the playback of the audio.
	//
	// The offset parameter specifies the position in seconds from which to start
	// the playback.
	Start(offset float32)

	// Stop stops the playback of the audio.
	Stop()

	// Resume resumes the playback of the audio.
	//
	// If the playback is already playing, this method has no effect.
	//
	// If the playback is stopped, this method has the same effect as Start with
	// an offset of 0.
	Resume()

	// Pause pauses the playback of the audio.
	//
	// If the playback is already paused or stopped, this method has no effect.
	Pause()

	// IsPlaying returns true if the playback is currently playing.
	IsPlaying() bool

	// IsLoop returns true if the playback is set to loop when it reaches the end
	// of the media.
	IsLoop() bool

	// SetLoop sets whether the playback should loop when it reaches the end of
	// the media.
	SetLoop(loop bool)

	// LoopStart returns the loop start position in seconds.
	LoopStart() float32

	// SetLoopStart sets the loop start position in seconds.
	SetLoopStart(loopStart float32)

	// LoopEnd returns the loop end position in seconds.
	LoopEnd() float32

	// SetLoopEnd sets the loop end position in seconds.
	SetLoopEnd(loopEnd float32)
}

// OscillatorNode represents an audio node that generates periodic waveforms.
type OscillatorNode interface {
	UserNode

	// Frequency returns the frequency of the oscillator in Hertz.
	Frequency() float32

	// SetFrequency sets the frequency of the oscillator in Hertz.
	SetFrequency(frequency float32)
}

// GainNode represents an audio node that applies a gain (volume adjustment) to
// the audio signal.
type GainNode interface {
	UserNode

	// Gain returns the gain factor applied to the audio signal.
	Gain() float32

	// SetGain sets the gain factor applied to the audio signal.
	SetGain(gain float32)
}

// PanNode represents an audio node that applies panning to the audio signal,
// distributing the signal between left and right channels.
type PanNode interface {
	UserNode

	// Pan returns the pan value, where -1.0 is full left, 0.0 is center, and
	// 1.0 is full right.
	Pan() float32

	// SetPan sets the pan value, where -1.0 is full left, 0.0 is center, and
	// 1.0 is full right.
	SetPan(pan float32)
}

// SpatialNode represents an audio node that provides spatial audio effects.
type SpatialNode interface {
	UserNode

	// Position returns the 3D position of the audio source.
	Position() sprec.Vec3

	// SetPosition sets the 3D position of the audio source.
	SetPosition(position sprec.Vec3)
}

// HighPassNode represents an audio node that applies a high-pass filter to
// the audio signal.
type HighPassNode interface {
	UserNode

	// CutoffFrequency returns the cutoff frequency of the high-pass filter in
	// Hertz.
	CutoffFrequency() float32

	// SetCutoffFrequency sets the cutoff frequency of the high-pass filter in
	// Hertz.
	SetCutoffFrequency(frequency float32)
}

// LowPassNode represents an audio node that applies a low-pass filter to
// the audio signal.
type LowPassNode interface {
	UserNode

	// CutoffFrequency returns the cutoff frequency of the low-pass filter in
	// Hertz.
	CutoffFrequency() float32

	// SetCutoffFrequency sets the cutoff frequency of the low-pass filter in
	// Hertz.
	SetCutoffFrequency(frequency float32)
}

// DelayNode represents an audio node that applies a delay effect to the audio
// signal.
type DelayNode interface {
	UserNode

	// DelayTime returns the delay time in seconds.
	DelayTime() float32

	// SetDelayTime sets the delay time in seconds.
	//
	// The maximum supported delay time may be limited by the implementation
	// but should be at least 1 second.
	SetDelayTime(delayTime float32)
}

// ReverbNode represents an audio node that applies a reverb effect to the
// audio signal.
type ReverbNode interface {
	UserNode

	// RoomSize returns the size of the virtual room for the reverb effect.
	RoomSize() float32

	// SetRoomSize sets the size of the virtual room for the reverb effect.
	SetRoomSize(size float32)
}

// CompressorNode represents an audio node that applies dynamic range
// compression to the audio signal.
type CompressorNode interface {
	UserNode

	// Threshold returns the threshold level in decibels.
	Threshold() float32

	// SetThreshold sets the threshold level in decibels.
	SetThreshold(threshold float32)
}

// ConnectorNode represents a no-op audio node that can be used to connect
// other nodes together without affecting the audio signal.
type ConnectorNode interface {
	UserNode
}
