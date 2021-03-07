package command

import "github.com/mokiat/gomath/sprec"

type Clear struct {
	ClearColors  [8]OptionalClearColor
	ClearDepth   OptionalFloat32
	ClearStencil OptionalUint32
}

type ClearColor struct {
	Attachment int
	Color      sprec.Vec4
}
