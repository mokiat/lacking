package mdl

import "github.com/mokiat/lacking/core/audio"

// Sound is a sound asset, such as a sound effect or a piece of music,
// represented as decoded stereo PCM [audio.Frame] values that are played
// back at a given sample rate.
type Sound struct {
	*Object
	sampleRate int
	frames     []audio.Frame
}

// NewSound creates a new empty [Sound] with a sample rate of 44100 Hz.
func NewSound() *Sound {
	return &Sound{
		Object:     NewObject(),
		sampleRate: 44100,
	}
}

// SampleRate returns the number of frames per second at which the sound
// is meant to be played back.
func (s *Sound) SampleRate() int {
	return s.sampleRate
}

// SetSampleRate changes the sample rate of the sound. The existing frames
// are not resampled.
func (s *Sound) SetSampleRate(rate int) {
	s.sampleRate = rate
}

// Frames returns the PCM frames of the sound. The returned slice is not
// a copy and should not be modified.
func (s *Sound) Frames() []audio.Frame {
	return s.frames
}

// SetFrames replaces the PCM frames of the sound. The sound retains a
// reference to the provided slice.
func (s *Sound) SetFrames(f []audio.Frame) {
	s.frames = f
}

// AppendFrame adds a single frame to the end of the sound.
func (s *Sound) AppendFrame(f audio.Frame) {
	s.frames = append(s.frames, f)
}

// AppendFrames adds the specified frames to the end of the sound.
func (s *Sound) AppendFrames(frames ...audio.Frame) {
	s.frames = append(s.frames, frames...)
}
