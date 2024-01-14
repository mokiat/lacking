package render

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

// Capabilities describes the capabilities of the implementation.
type Capabilities struct {

	// Quality indicates the performance capabilities of the implementation.
	Quality Quality
}
