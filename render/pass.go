package render

type LoadOperation int

const (
	LoadOperationDontCare LoadOperation = iota
	LoadOperationClear
)

type StoreOperation int

const (
	StoreOperationDontCare StoreOperation = iota
	StoreOperationStore
)

type RenderPassInfo struct {
	Framebuffer Framebuffer
	Viewport    Area

	DepthLoadOp     LoadOperation
	DepthStoreOp    StoreOperation
	DepthClearValue float32

	StencilLoadOp     LoadOperation
	StencilStoreOp    StoreOperation
	StencilClearValue int

	Colors [4]ColorAttachmentInfo
}

type ColorAttachmentInfo struct {
	LoadOp     LoadOperation
	StoreOp    StoreOperation
	ClearValue [4]float32
}

type Area struct {
	X      int
	Y      int
	Width  int
	Height int
}
