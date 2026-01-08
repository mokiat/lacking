package audio

import "github.com/mokiat/gomath/sprec"

// Node represents a node in a chain of audio elements. Each node produces
// audio data which can be synthesized, processed, or played back.
type Node any

// UserNode represents an audio node that requires explicit resource management.
type UserNode interface {
	Node

	// Delete releases any resources associated with the node. After calling
	// this method, the node should not be used anymore.
	Delete()
}

// PlaybackNode represents an audio node that plays back audio data from a
// Media source.
type PlaybackNode interface {
	UserNode

	// Loop returns true if the playback is set to loop when it reaches the end
	// of the media.
	Loop() bool

	// Done returns true if the playback has reached the end of the media
	// and is not looping.
	Done() bool
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
