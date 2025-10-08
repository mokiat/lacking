package hierarchy2

import "github.com/mokiat/gomath/dprec"

const initialTimestamp int32 = -1

type NodeID struct {
	index    int32
	revision uint32
}

var NilNodeID = NodeID{0, 0}

type node struct {
	revision  uint32
	timestamp int32
	name      string

	parentIndex       int32
	firstChildIndex   int32
	lastChildIndex    int32
	leftSiblingIndex  int32
	rightSiblingIndex int32

	previousPosition  dprec.Vec3
	previousRotation  dprec.Quat
	previousScale     dprec.Vec3
	previousAbsMatrix dprec.Mat4

	currentPosition  dprec.Vec3
	currentRotation  dprec.Quat
	currentScale     dprec.Vec3
	currentAbsMatrix dprec.Mat4
}
