package audio

// Bus represents a flat audio bus that can be used to group sound sources together for collective control.
type Bus interface {

	// Gain returns the current gain of the bus.
	//
	// Default is 1.0, which means no change in volume.
	Gain() float32

	// SetGain sets the gain of the bus.
	SetGain(gain float32)

	// Reverb returns the reverb controls of the bus.
	//
	// If the bus was not created with reverb enabled, this will return nil.
	Reverb() Reverb

	// Compression returns the compression controls of the bus.
	//
	// If the bus was not created with compression enabled, this will return nil.
	Compression() Compression

	// Release releases any resources associated with the bus.
	//
	// All attached sound sources will be stopped.
	Release()
}

// BusSettings represents the settings for creating a new audio bus.
type BusSettings struct {

	// UseReverb indicates whether to enable reverb on the bus.
	//
	// Default is false.
	UseReverb bool

	// UseCompression indicates whether to enable compression on the bus.
	//
	// Default is false.
	UseCompression bool
}
