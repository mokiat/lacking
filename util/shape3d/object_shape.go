package shape3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/util/spatial"
)

type ShapeID shapeRef

type ShapeInfo struct {
	RejectGroup uint32
	SourceMask  opt.T[uint32]
	TargetMask  opt.T[uint32]
}

type Shape struct { // TODO: make private
	objectIndex uint32
	nextShape   shapeRef

	spatialID spatial.CompactOctreeItemID
	static    bool

	rejectGroup uint32
	sourceMask  uint32
	targetMask  uint32
}

func shapesCanIntersect(a, b *Shape) bool {
	if a.objectIndex == b.objectIndex {
		return false
	}
	if a.rejectGroup != 0 && (a.rejectGroup == b.rejectGroup) {
		return false
	}
	if ((a.sourceMask & b.targetMask) == 0) && ((a.targetMask & b.sourceMask) == 0) {
		return false
	}
	return true
}

const (
	shapeKindNone shapeKind = iota
	shapeKindSphere
	shapeKindBox
	shapeKindMesh
)

type shapeKind uint8

const invalidShapeRef = shapeRef(0) // has none shape kind

func newShapeRef(kind shapeKind, index uint32) shapeRef {
	return shapeRef((index << 4) | (uint32(kind) & 0b1111))
}

type shapeRef uint32

func (r shapeRef) Index() uint32 {
	return uint32(r) >> 4
}

func (r shapeRef) Kind() shapeKind {
	return shapeKind(r & 0b1111)
}
