package audio

// API abstracts the underlying audio system, providing a consistent interface for audio manipulation and playback.
//
// All methods must be called from the UI thread.
type API interface {

	// CreateBus creates a new flat audio bus.
	CreateBus(settings BusSettings) Bus

	// MasterBus returns the master bus for the audio system.
	MasterBus() MasterBus

	// SpatialListener returns the spatial listener used for 3D audio.
	SpatialListener() SpatialListener
}

// MasterBus represents the master bus for the audio system, controlling the overall output of all audio.
type MasterBus interface {

	// Gain returns the master gain for the audio system.
	//
	// Default value is 1.0.
	Gain() float32

	// SetGain sets the master gain for the audio system.
	//
	// The value must be non-negative.
	SetGain(gain float32)

	// Compression returns the global compression controls for the audio system.
	Compression() Compression
}
