package internal

import "github.com/mokiat/lacking/ui"

const maxLayerDepth = 256

type Layer struct {
	depth    int
	previous *Layer
	next     *Layer

	Translation ui.Position
	ClipBounds  ui.Bounds
	SolidColor  ui.Color
	StrokeColor ui.Color
	StrokeSize  int
	Font        *Font
	FontSize    int
}

func (l *Layer) InheritFrom(other *Layer) {
	l.Translation = other.Translation
	l.ClipBounds = other.ClipBounds
	l.SolidColor = other.SolidColor
	l.StrokeColor = other.StrokeColor
	l.StrokeSize = other.StrokeSize
	l.Font = other.Font
}

func (l *Layer) Previous() *Layer {
	if l.previous == nil {
		panic("too many pops: no more layers")
	}
	return l.previous
}

func (l *Layer) Next() *Layer {
	if l.depth >= maxLayerDepth {
		panic("too many pushes: max layer depth reached")
	}
	if l.next == nil {
		l.next = &Layer{
			previous: l,
			depth:    l.depth + 1,
		}
	}
	l.next.InheritFrom(l)
	return l.next
}
