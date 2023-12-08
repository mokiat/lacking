package audio

// PlayInfo specifies conditions for the playback.
type PlayInfo struct {

	// Loop indicates whether the playback should loop.
	Loop bool

	// Gain indicates the amount of volume, where 1.0 is max and 0.0 is min.
	Gain float64

	// Pan indicates the sound panning, where -1.0 is left, 0.0 is center, and
	// 1.0 is right.
	Pan float64
}

type Playback interface {

	// Stop causes the playback to stop.
	Stop()
}
