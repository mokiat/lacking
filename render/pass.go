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
	StencilClearValue int
}

// LoadOperation describes how the contents of a resource should be loaded.
type LoadOperation int

const (
	// LoadOperationDontCare means that the contents of the resource should not
	// be loaded.
	LoadOperationDontCare LoadOperation = iota

	// LoadOperationClear means that the contents of the resource should be
	// cleared.
	LoadOperationClear
)

// StoreOperation describes how the contents of a resource should be stored.
type StoreOperation int

const (
	// StoreOperationDontCare means that the contents of the resource should not
	// be stored.
	StoreOperationDontCare StoreOperation = iota

	// StoreOperationStore means that the contents of the resource should be
	// stored.
	StoreOperationStore
)

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
	X int

	// Y is the Y coordinate of the area.
	Y int

	// Width is the width of the area.
	Width int

	// Height is the height of the area.
	Height int
}
