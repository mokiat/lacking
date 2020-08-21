package world

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/resource"
)

type Renderable struct {
	prev *Renderable
	next *Renderable

	Matrix sprec.Mat4
	Radius float32
	Model  *resource.Model
}

func (r *Renderable) attach(target *Renderable) {
	r.prev = target
	r.next = target.next

	if r.next != nil {
		r.next.prev = r
	}
	target.next = r
}

func (r *Renderable) detach() {
	prev := r.prev
	next := r.next
	if prev != nil {
		prev.next = next
	}
	if next != nil {
		next.prev = prev
	}
	r.prev = nil
	r.next = nil
}
