package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

func newCanvasState() *canvasState {
	return &canvasState{
		topLayer: &canvasLayer{},

		projectionMatrix: sprec.IdentityMat4(),
	}
}

type canvasState struct {
	whiteMask render.Texture

	topLayer     *canvasLayer
	currentLayer *canvasLayer

	projectionMatrix sprec.Mat4
}

func (s *canvasState) onCreate(api render.API) {
	s.whiteMask = api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           1,
		Height:          1,
		Filtering:       render.FilterModeNearest,
		Wrapping:        render.WrapModeClamp,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
		Data:            []byte{0xFF, 0xFF, 0xFF, 0xFF},
	})
}

func (s *canvasState) onDestroy() {
	defer s.whiteMask.Release()
}

type canvasLayer struct {
	depth    int
	previous *canvasLayer
	next     *canvasLayer

	Transform  sprec.Mat4
	ClipBounds Bounds
}

func (l *canvasLayer) InheritFrom(other *canvasLayer) {
	l.Transform = other.Transform
	l.ClipBounds = other.ClipBounds
}

func (l *canvasLayer) Previous() *canvasLayer {
	if l.previous == nil {
		panic("too many pops: no more layers")
	}
	return l.previous
}

func (l *canvasLayer) Next() *canvasLayer {
	if l.depth >= maxLayerDepth {
		panic("too many pushes: max layer depth reached")
	}
	if l.next == nil {
		l.next = &canvasLayer{
			previous: l,
			depth:    l.depth + 1,
		}
	}
	l.next.InheritFrom(l)
	return l.next
}

func uiColorToVec(color Color) sprec.Vec4 {
	return sprec.NewVec4(
		float32(color.R)/255.0,
		float32(color.G)/255.0,
		float32(color.B)/255.0,
		float32(color.A)/255.0,
	)
}
