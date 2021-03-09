package command

import "github.com/mokiat/gomath/sprec"

type ClearFramebuffer struct {
	Colors  ClearColorRange
	Depth   OptionalFloat32
	Stencil OptionalUint32
}

type ClearColorRange struct {
	Offset int
	Count  int
}

type ClearColor struct {
	Attachment int
	Color      sprec.Vec4
}
