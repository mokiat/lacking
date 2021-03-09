package command

type ID struct {
	Type  Type
	Index int
}

type Type int

const (
	TypeChangeFramebuffer Type = 1 + iota
	TypeClearFramebuffer
	TypeChangeDepthConfig
	TypeChangeReorderConfig
	TypeRenderItem
)
