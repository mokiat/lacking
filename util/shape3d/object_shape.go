package shape3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/util/spatial"
)

// InvalidShapeID indicates a shape that can never be part of the scene.
const InvalidShapeID = ShapeID(invalidShapeRef)

// ShapeID is a reference to a shape in the scene.
type ShapeID shapeRef

// ShapeInfo contains information needed to create a new shape in the scene.
type ShapeInfo[S any] struct {

	// RejectGroup becomes active if a value larger than zero is specified.
	// Shapes that share the same reject group are not checked for intersection.
	RejectGroup uint32

	// SourceMask specifies the layers in which this shape is positioned.
	SourceMask opt.T[uint32]

	// TargetMask specifies the layers with which this shape can intersect.
	TargetMask opt.T[uint32]

	// UserData allows one to attach custom user data to a shape.
	UserData S
}

type sceneShape[S any] struct {
	objectIndex uint32
	nextShape   shapeRef

	spatialID spatial.CompactOctreeItemID
	static    bool

	rejectGroup uint32
	sourceMask  uint32
	targetMask  uint32

	userData S
}

func shapesCanIntersect[S any](a, b *sceneShape[S]) bool {
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

func (r shapeRef) index() uint32 {
	return uint32(r) >> 4
}

func (r shapeRef) kind() shapeKind {
	return shapeKind(r & 0b1111)
}
