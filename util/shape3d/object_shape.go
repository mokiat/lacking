package shape3d

import "github.com/mokiat/lacking/util/mem"

type ShapeID struct {
	internalID mem.SparseID
}

func (i ShapeID) IsNil() bool {
	return i == (ShapeID{})
}

type objectShape struct {
	id          mem.SparseID
	objectID    mem.SparseID
	nextShapeID mem.SparseID

	actualID mem.SparseID
	kind     shapeKind
}

const (
	shapeKindSphere shapeKind = iota
	shapeKindBox
	shapeKindMesh
)

type shapeKind uint8
