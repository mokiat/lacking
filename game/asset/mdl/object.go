package mdl

import "sync/atomic"

var freeID atomic.Uint32

func NewObject() *Object {
	return &Object{
		id: freeID.Add(1),
	}
}

type Object struct {
	id uint32
}

func (o *Object) ID() uint32 {
	return o.id
}
