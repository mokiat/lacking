package audio

import "github.com/mokiat/gomath/sprec"

const (
	// DefaultFrequency is the default frequency for OscillatorNode.
	DefaultFrequency = 440.0

	// DefaultGain is the default gain factor for GainNode, representing no
	// change to the audio signal.
	DefaultGain = 1.0

	// DefaultPan is the default pan value for PanNode, representing a
	// centered audio signal.
	DefaultPan = 0.0

	// DefaultCutoffFrequency is the default cutoff frequency for HighPassNode
	// and LowPassNode.
	DefaultCutoffFrequency = 350.0

	// DefaultDelay is the default delay time for DelayNode, representing no
	// delay.
	DefaultDelay = 0.0

	// DefaultRoomSize is the default room size for ReverbNode, representing a
	// small room.
	DefaultRoomSize = 0.3

	// DefaultDamping is the default damping factor for ReverbNode, representing
	// a moderate amount of damping.
	DefaultDamping = 0.5

	// DefaultDry is the default dry level for ReverbNode, representing the
	// original signal at full gain.
	DefaultDry = 1.0

	// DefaultWet is the default wet level for ReverbNode, representing the
	// reverberated signal at half gain.
	DefaultWet = 0.5

	// DefaultAttack is the default attack time for CompressorNode.
	DefaultAttack = 0.003

	// DefaultRelease is the default release time for CompressorNode.
	DefaultRelease = 0.25

	// DefaultRatio is the default compression ratio for CompressorNode.
	DefaultRatio = 12.0

	// DefaultKnee is the default knee width for CompressorNode.
	DefaultKnee = 30.0

	// DefaultThreshold is the default threshold level for CompressorNode.
	DefaultThreshold = -24.0
)

// Node represents a node in a chain of audio elements. Each node produces
// audio data which can be synthesized, processed, or played back.
type Node interface {
	_isAudioNode() // marker method
}

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

	// Resume resumes a paused playback from where it was paused.
	//
	// If the playback is already playing, this method has no effect. If the
	// playback is stopped rather than paused, this method starts from the
	// beginning (equivalent to Start(0)).
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
	//
	// Default value is 440.0 Hz (A4 note).
	Frequency() float32

	// SetFrequency sets the frequency of the oscillator in Hertz.
	SetFrequency(frequency float32)
}

// GainNode represents an audio node that applies a gain (volume adjustment) to
// the audio signal.
type GainNode interface {
	UserNode

	// Gain returns the gain factor applied to the audio signal.
	//
	// A value of 1.0 means no change, 0.0 is silence, and values greater than
	// 1.0 amplify the signal.
	//
	// Default value is 1.0.
	Gain() float32

	// SetGain sets the gain factor applied to the audio signal.
	//
	// The value must be non-negative.
	SetGain(gain float32)
}

// PanNode represents an audio node that applies panning to the audio signal,
// distributing the signal between left and right channels.
type PanNode interface {
	UserNode

	// Pan returns the pan value, where -1.0 is full left, 0.0 is center, and
	// 1.0 is full right.
	//
	// Default value is 0.0 (center).
	Pan() float32

	// SetPan sets the pan value, where -1.0 is full left, 0.0 is center, and
	// 1.0 is full right.
	SetPan(pan float32)
}

// SpatialNode represents an audio node that provides spatial audio effects.
// Implementations must apply inverse distance attenuation relative to the
// [SpatialListener] returned by [API.SpatialListener].
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
	//
	// Default value is 350.0 Hz.
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
	//
	// Default value is 350.0 Hz.
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
	//
	// Default value is 0.0 seconds (no delay).
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
	//
	// The value is in the range [0.0, 1.0]. Default value is 0.3.
	RoomSize() float32

	// SetRoomSize sets the size of the virtual room for the reverb effect.
	//
	// The value will be clamped to the range [0.0, 1.0].
	SetRoomSize(size float32)

	// Damping returns the damping factor of the reverb effect.
	//
	// Higher values cause high frequencies to decay faster, simulating
	// absorptive surfaces.
	//
	// Default value is 0.5.
	Damping() float32

	// SetDamping sets the damping factor of the reverb effect.
	//
	// The value will be clamped to the range [0.0, 1.0].
	SetDamping(damping float32)

	// Dry returns the dry level of the reverb effect.
	//
	// Default value is 1.0.
	Dry() float32

	// SetDry sets the dry level of the reverb effect.
	//
	// The value will be clamped to the range [0.0, 1.0].
	SetDry(dry float32)

	// Wet returns the wet level of the reverb effect.
	//
	// Default value is 0.5.
	Wet() float32

	// SetWet sets the wet level of the reverb effect.
	//
	// The value will be clamped to the range [0.0, 1.0].
	SetWet(wet float32)
}

// CompressorNode represents an audio node that applies dynamic range
// compression to the audio signal.
type CompressorNode interface {
	UserNode

	// Attack returns the attack time in seconds.
	//
	// Default value is 0.003 seconds.
	Attack() float32

	// SetAttack sets the attack time in seconds.
	//
	// The value will be clamped to the range [0.0, 1.0].
	SetAttack(attack float32)

	// Release returns the release time in seconds.
	//
	// Default value is 0.25 seconds.
	Release() float32

	// SetRelease sets the release time in seconds.
	//
	// The value will be clamped to the range [0.0, 1.0].
	SetRelease(release float32)

	// Ratio returns the compression ratio.
	//
	// Default value is 12.0.
	Ratio() float32

	// SetRatio sets the compression ratio.
	//
	// The value will be clamped to the range [1.0, 20.0].
	SetRatio(ratio float32)

	// Knee returns the knee width in decibels.
	//
	// Default value is 30.0 dB.
	Knee() float32

	// SetKnee sets the knee width in decibels.
	//
	// The value will be clamped to the range [0.0, 40.0].
	SetKnee(knee float32)

	// Threshold returns the threshold level in decibels.
	//
	// Default value is -24.0 dB.
	Threshold() float32

	// SetThreshold sets the threshold level in decibels.
	//
	// The value will be clamped to the range [-100.0, 0.0].
	SetThreshold(threshold float32)
}

// ConnectorNode represents a no-op audio node that can be used to connect
// other nodes together without affecting the audio signal.
type ConnectorNode interface {
	UserNode
}
