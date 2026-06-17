package shape2d

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/core/spatial/query2d"
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

	spatialID query2d.TreeItemID
	static    bool

	rejectGroup uint32
	sourceMask  uint32
	targetMask  uint32

	userData S
}

func (s *sceneShape[S]) matchesFilter(filter Filter) bool {
	if s.static && filter.SkipStatic {
		return false
	}
	if !s.static && filter.SkipDynamic {
		return false
	}
	if mask, ok := filter.Mask.Unwrap(); ok {
		if (s.sourceMask & mask) == 0 {
			return false
		}
	}
	return true
}

func shapesCanIntersect[S any](a, b *sceneShape[S]) bool {
	if a.objectIndex == b.objectIndex {
		return false
	}
	if !a.static && !b.static && a.objectIndex >= b.objectIndex {
		return false // prevent double checks for dynamic shapes
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
	shapeKindSegment
	shapeKindCircle
	shapeKindRectangle
	shapeKindPolygon
)

type shapeKind uint8

func (k shapeKind) String() string {
	switch k {
	case shapeKindNone:
		return "None"
	case shapeKindSegment:
		return "Segment"
	case shapeKindCircle:
		return "Circle"
	case shapeKindRectangle:
		return "Rectangle"
	case shapeKindPolygon:
		return "Polygon"
	default:
		return "Unknown"
	}
}

const invalidShapeRef = shapeRef(0) // has none shape kind

const tempShapeIndex = uint32(0xFFFFFFE) // reserved index for temporary shapes

func newShapeRef(kind shapeKind, index uint32) shapeRef {
	return shapeRef((index << 4) | (uint32(kind) & 0b1111))
}

func newTempShapeRef(kind shapeKind) shapeRef {
	return newShapeRef(kind, tempShapeIndex)
}

type shapeRef uint32

func (r shapeRef) String() string {
	return fmt.Sprintf("%s:%d", r.kind(), r.index())
}

func (r shapeRef) index() uint32 {
	return uint32(r) >> 4
}

func (r shapeRef) kind() shapeKind {
	return shapeKind(r & 0b1111)
}

func (r shapeRef) isTemporary() bool {
	return r.index() == tempShapeIndex
}

func newShapeRefPair(source, target shapeRef) shapeRefPair {
	return shapeRefPair(uint64(source)<<32 | uint64(target))
}

type shapeRefPair uint64

func (p shapeRefPair) flipped() shapeRefPair {
	return newShapeRefPair(p.target(), p.source())
}

func (p shapeRefPair) source() shapeRef {
	return shapeRef(uint32(p >> 32))
}

func (p shapeRefPair) target() shapeRef {
	return shapeRef(uint32(p & 0xFFFFFFFF))
}
