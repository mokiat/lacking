package placement3d

import "fmt"

// InvalidShapeID indicates a shape that can never be part of the scene.
const InvalidShapeID = ShapeID(invalidShapeRef)

const invalidShapeRef = -1

type shapeRef int32

func newShapeRef(index int32, isMesh bool) shapeRef {
	result := uint32(index) << 1
	if isMesh {
		result |= 1
	}
	return shapeRef(index)
}

func (r shapeRef) isMesh() bool {
	return r%2 == 1
}

func (r shapeRef) index() int32 {
	return int32(r >> 1)
}

func (r shapeRef) String() string {
	return fmt.Sprintf("%d [%t]", r.index(), r.isMesh())
}

type shapeRefPair struct {
	source shapeRef
	target shapeRef
}

func newShapeRefPair(source, target shapeRef) shapeRefPair {
	return shapeRefPair{
		source: source,
		target: target,
	}
}

func (p shapeRefPair) flipped() shapeRefPair {
	return shapeRefPair{
		source: p.target,
		target: p.source,
	}
}
