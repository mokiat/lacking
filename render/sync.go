package render

// FenceMarker marks a type as being a Fence.
type FenceMarker interface {
	_isFenceType()
}

// Fence is the interface that provides the API with the ability
// to synchronize with the GPU.
type Fence interface {
	FenceMarker
	Resource

	// Status returns the current status of the fence.
	Status() FenceStatus
}

const (
	// FenceStatusNotReady indicates that the GPU has not reached the fence yet.
	FenceStatusNotReady FenceStatus = iota

	// FenceStatusSuccess indicates that the GPU has processed all commands
	// up to the fence.
	FenceStatusSuccess
)

// FenceStatus represents the status of a Fence.
type FenceStatus uint8

// String returns a string representation of the FenceStatus.
func (s FenceStatus) String() string {
	switch s {
	case FenceStatusNotReady:
		return "NOT_READY"
	case FenceStatusSuccess:
		return "SUCCESS"
	default:
		return "UNKNOWN"
	}
}
