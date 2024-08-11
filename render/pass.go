package render

// RenderPassInfo describes the information needed to begin a new render pass.
type RenderPassInfo struct {

	// Framebuffer is the Framebuffer that the render pass will render to.
	Framebuffer Framebuffer

	// Viewport describes the area of the Framebuffer that the render pass will
	// render to.
	Viewport Area

	// Colors describes the color attachments that the render pass will render
	// to.
	Colors [4]ColorAttachmentInfo

	// DepthLoadOp describes how the contents of the depth attachment should be
	// loaded.
	DepthLoadOp LoadOperation

	// DepthStoreOp describes how the contents of the depth attachment should be
	// stored.
	DepthStoreOp StoreOperation

	// DepthClearValue is the value that should be used to clear the depth
	// attachment when DepthLoadOp is LoadOperationClear.
	DepthClearValue float32

	// StencilLoadOp describes how the contents of the stencil attachment should
	// be loaded.
	StencilLoadOp LoadOperation

	// StencilStoreOp describes how the contents of the stencil attachment
	// should be stored.
	StencilStoreOp StoreOperation

	// StencilClearValue is the value that should be used to clear the stencil
	// attachment when StencilLoadOp is LoadOperationClear.
	StencilClearValue uint32
}

const (
	// LoadOperationLoad means that the contents of the resource should be
	// made available as is.
	LoadOperationLoad LoadOperation = iota

	// LoadOperationClear means that the contents of the resource should be
	// cleared.
	LoadOperationClear
)

// LoadOperation describes how the contents of a resource should be loaded.
type LoadOperation int8

// String returns a string representation of the LoadOperation.
func (o LoadOperation) String() string {
	switch o {
	case LoadOperationLoad:
		return "LOAD"
	case LoadOperationClear:
		return "CLEAR"
	default:
		return "UNKNOWN"
	}
}

const (
	// StoreOperationDiscard means that the contents of the resource may be
	// discarded if that would improve performance.
	StoreOperationDiscard StoreOperation = iota

	// StoreOperationStore means that the contents of the resource should be
	// preserved.
	StoreOperationStore
)

// StoreOperation describes how the contents of a resource should be stored.
type StoreOperation int8

// String returns a string representation of the StoreOperation.
func (o StoreOperation) String() string {
	switch o {
	case StoreOperationDiscard:
		return "DISCARD"
	case StoreOperationStore:
		return "STORE"
	default:
		return "UNKNOWN"
	}
}

// ColorAttachmentInfo describes how a color attachment should be handled.
type ColorAttachmentInfo struct {

	// LoadOp describes how the contents of the color attachment should be
	// loaded.
	LoadOp LoadOperation

	// StoreOp describes how the contents of the color attachment should be
	// stored.
	StoreOp StoreOperation

	// ClearValue is the value that should be used to clear the color attachment
	// when LoadOp is LoadOperationClear.
	ClearValue [4]float32
}

// Area describes a rectangular area to be used for rendering.
type Area struct {

	// X is the X coordinate of the area.
	X uint32

	// Y is the Y coordinate of the area.
	Y uint32

	// Width is the width of the area.
	Width uint32

	// Height is the height of the area.
	Height uint32
}
