package render

// Limits describes the limits of the implementation.
type Limits interface {

	// Quality returns the supported quality level by the implementation.
	// This is usually based on the performance characteristics of the
	// underlying hardware.
	//
	// This is meant to be used as a hint for the application to decide
	// on the level of detail to use for rendering.
	Quality() Quality

	// UniformBufferOffsetAlignment returns the alignment requirement
	// for uniform buffer offsets.
	UniformBufferOffsetAlignment() int
}

// Quality is an enumeration of the supported render quality levels.
type Quality int

const (
	// QualityLow indicates that the implementation is running on hardware
	// with low performance capabilities.
	QualityLow Quality = iota

	// QualityMedium indicates that the implementation is running on hardware
	// with medium performance capabilities.
	QualityMedium

	// QualityHigh indicates that the implementation is running on hardware
	// with high performance capabilities.
	QualityHigh
)
