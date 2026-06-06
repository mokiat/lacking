package audio

// Playback represents an instance of a media being played on a bus. It allows control over the playback state of
// the media.
type Playback interface {

	// Start begins playback of the media. If the media is already playing, this method has no effect.
	//
	// The at parameter specifies the starting point in seconds from the beginning of the media. If at is negative or
	// greater than the length of the media, it will be clamped to the valid range.
	Start(at float32)

	// Stop halts playback of the media. If the media is not playing, this method has no effect.
	Stop()

	// Pause temporarily halts playback of the media. If the media is not playing, this method has no effect.
	Pause()

	// Resume continues playback of the media if it was paused. If the media is not paused, this method has no effect.
	Resume()

	// Looping returns true if the media is set to loop, false otherwise.
	Looping() bool

	// SetLooping sets whether the media should loop when it reaches the end.
	SetLooping(loop bool)

	// LoopStart returns the starting point in seconds from the beginning of the media where looping should occur.
	//
	// Default value is 0.0 seconds.
	LoopStart() float32

	// SetLoopStart sets the starting point in seconds from the beginning of the media where looping should occur.
	SetLoopStart(loopStart float32)

	// LoopEnd returns the ending point in seconds from the beginning of the media where looping should occur.
	//
	// Default value is the length of the media.
	LoopEnd() float32

	// SetLoopEnd sets the ending point in seconds from the beginning of the media where looping should occur.
	SetLoopEnd(loopEnd float32)

	// Playing returns true if the media is currently playing, false otherwise.
	Playing() bool

	// PlaybackRate returns the current playback rate of the media.
	PlaybackRate() float32

	// SetPlaybackRate sets the playback rate of the media.
	SetPlaybackRate(rate float32)

	// Gain returns the current gain of the playback.
	//
	// Default value is 1.0, which means no change in volume.
	Gain() float32

	// SetGain sets the gain of the playback.
	SetGain(gain float32)

	// LowPassFilter returns the low-pass filter applied to this playback, if any. If no low-pass filter is applied,
	// this method returns nil.
	LowPassFilter() FrequencyFilter

	// HighPassFilter returns the high-pass filter applied to this playback, if any. If no high-pass filter is applied,
	// this method returns nil.
	HighPassFilter() FrequencyFilter

	// SetOnFinished sets a callback function that will be called when the media finishes playing naturally,
	// i.e. when it reaches the end and is not set to loop. It will not be called if playback is stopped
	// via Stop() or paused via Pause(), nor on each loop iteration.
	//
	// If looping is disabled (via SetLooping) while the media is playing, and the media subsequently
	// reaches its end, the callback will be called.
	SetOnFinished(onFinished func())

	// Release releases any resources associated with this playback instance. After calling this method, the playback
	// should not be used anymore.
	Release()
}

// SpatialPlayback represents a playback instance that also has spatial properties, allowing it to be positioned
// and oriented in 3D space for spatial audio effects.
type SpatialPlayback interface {
	Playback
	SpatialEmitter
}

// PlaybackSettings represents the settings for creating a playback instance.
type PlaybackSettings struct {

	// FireAndForget indicates whether the playback should automatically release its resources after it finishes playing.
	FireAndForget bool

	// UseLowPassFilter indicates whether a low-pass filter should be applied to the playback.
	UseLowPassFilter bool

	// UseHighPassFilter indicates whether a high-pass filter should be applied to the playback.
	UseHighPassFilter bool
}
