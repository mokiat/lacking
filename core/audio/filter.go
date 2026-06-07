package audio

// Reverb represents the settings for audio reverb on a bus.
type Reverb interface {

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

// Compression represents the settings for audio compression on a bus.
type Compression interface {

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

// FrequencyFilter represents a simple frequency filter.
type FrequencyFilter interface {

	// Frequency returns the cutoff frequency of the filter in hertz.
	Frequency() float32

	// SetFrequency sets the cutoff frequency of the filter in hertz.
	//
	// The value must be positive.
	SetFrequency(frequency float32)
}
