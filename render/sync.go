package render

type FenceObject interface {
	_isFenceObject() bool // ensures interface uniqueness
}

type Fence interface {
	FenceObject
	Status() FenceStatus
	Delete()
}

const (
	FenceStatusNotReady FenceStatus = iota
	FenceStatusSuccess
	FenceStatusDeviceLost
)

type FenceStatus int
