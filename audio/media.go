package audio

import "time"

// MediaData represents the raw audio data and its associated metadata.
type MediaData struct {

	// Frames contains the decoded audio frames.
	Frames []Frame

	// SampleRate is the sample rate of the audio data.
	SampleRate int
}

// Media represents a playable audio sequence.
type Media interface {

	// Length returns the duration of the media.
	Length() time.Duration

	// Delete frees any resources used by this media.
	Delete()
}
