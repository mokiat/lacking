package ui

import "github.com/mokiat/gomath/sprec"

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
