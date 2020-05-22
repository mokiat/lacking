package graphics

import "github.com/mokiat/gomath/sprec"

type DepthFunc int

const (
	DepthFuncLess DepthFunc = iota
	DepthFuncLessOrEqual
)

func createSequence(items []Item) Sequence {
	return Sequence{
		items: items,
	}
}

type Sequence struct {
	items          []Item
	itemStartIndex int
	itemEndIndex   int

	SourceFramebuffer     *Framebuffer
	TargetFramebuffer     *Framebuffer
	BlitFramebufferColor  bool
	BlitFramebufferDepth  bool
	BlitFramebufferSmooth bool
	BackgroundColor       sprec.Vec4
	TestDepth             bool
	ClearColor            bool
	ClearDepth            bool
	WriteDepth            bool
	DepthFunc             DepthFunc
	ProjectionMatrix      sprec.Mat4
	ViewMatrix            sprec.Mat4
}

func (s *Sequence) BeginItem() *Item {
	if s.itemEndIndex == len(s.items) {
		panic("max number of render items reached")
	}
	item := &s.items[s.itemEndIndex]
	s.itemEndIndex++
	item.reset()
	return item
}

func (s *Sequence) EndItem(item *Item) {
}

func (s *Sequence) reset(index int) {
	s.itemStartIndex = index
	s.itemEndIndex = index

	s.SourceFramebuffer = nil
	s.TargetFramebuffer = nil
	s.BlitFramebufferColor = false
	s.BlitFramebufferDepth = false
	s.BlitFramebufferSmooth = false
	s.TestDepth = true
	s.ClearColor = false
	s.ClearDepth = false
	s.WriteDepth = true
	s.DepthFunc = DepthFuncLess
}

func (s *Sequence) itemsView() []Item {
	return s.items[s.itemStartIndex:s.itemEndIndex]
}
