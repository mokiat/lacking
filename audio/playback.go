package audio

import "github.com/mokiat/gog/opt"

// PlayInfo specifies conditions for the playback.
type PlayInfo struct {

	// Loop indicates whether the playback should loop.
	Loop bool

	// Gain indicates the amount of volume, where 1.0 is max and 0.0 is min.
	//
	// If not specified, the default value is 1.0.
	Gain opt.T[float64]

	// Pan indicates the sound panning, where -1.0 is left, 0.0 is center, and
	// 1.0 is right.
	Pan float64
}

// Playback represents the audio playback of a media file.
type Playback interface {

	// Stop causes the playback to stop.
	Stop()
}
