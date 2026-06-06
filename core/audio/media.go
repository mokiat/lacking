package audio

// MediaData represents the raw audio data and its associated metadata.
type MediaData struct {

	// Frames contains the decoded audio frames.
	Frames []Frame

	// SampleRate is the sample rate of the audio data.
	SampleRate int
}

// Media represents an audio media object that can be played back or manipulated.
type Media interface {

	// Length returns the duration of the media in seconds.
	Length() float64

	// Release releases any resources associated with the media. After calling this method,
	// the media should not be used anymore.
	//
	// Existing playback is not affected.
	Release()
}
