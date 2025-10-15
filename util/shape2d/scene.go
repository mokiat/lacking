package shape2d

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
)

// SceneInfo contains information needed to create an optimal scene.
type SceneInfo struct {

	// Size specifies the dimension (from side to side) of the scene.
	// Inserting an object outside these bounds has undefined behavior.
	//
	// If not specified, a default size of 4096 units is used.
	Size opt.T[float64]

	// MaxDepth specifies the maximum depth of the internal spatial
	// partitioning structure.
	//
	// If not specified, a default max depth of 8 is used.
	MaxDepth opt.T[uint32]

	// InitialNodeCapacity is a hint as to the number of nodes that will be
	// needed to place all items. Usually one would find that number empirically.
	// This allows the quadtree to preallocate memory and avoid dynamic allocations.
	//
	// By default the initial capacity is 4096.
	InitialNodeCapacity opt.T[uint32]

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the quadtree. This allows the quadtree to preallocate
	// memory and avoid dynamic allocations during insertion.
	//
	// By default the initial capacity is 1024.
	InitialItemCapacity opt.T[uint32]
}

// NewScene creates a new scene.
func NewScene[O, S any](info SceneInfo) *Scene[O, S] {
	cubeOctreeSettings := CompactTreeSettings(info)

	return &Scene[O, S]{
		freeObjectIndices:    ds.NewStack[uint32](256), // ~ 1 KiB
		freeCircleIndices:    ds.NewStack[uint32](256), // ~ 1 KiB
		freeRectangleIndices: ds.NewStack[uint32](256), // ~ 1 KiB
		freePolygonIndices:   ds.NewStack[uint32](256), // ~ 1 KiB

		objects:    make([]sceneObject[O], 0, 128),
		circles:    make([]sceneCircleShape[S], 0, 128),
		rectangles: make([]sceneRectangleShape[S], 0, 128),
		polygons:   make([]scenePolygonShape[S], 0, 128),

		circleTree:    NewCompactTree[uint32](cubeOctreeSettings),
		rectangleTree: NewCompactTree[uint32](cubeOctreeSettings),
		polygonTree:   NewCompactTree[uint32](cubeOctreeSettings),

		checks: make([]indexPair, 0, 1024),
	}
}

// Scene represents a 2D scene where objects made of shapes can be added.
type Scene[T, S any] struct {
	freeObjectIndices    *ds.Stack[uint32]
	freeCircleIndices    *ds.Stack[uint32]
	freeRectangleIndices *ds.Stack[uint32]
	freePolygonIndices   *ds.Stack[uint32]

	objects    []sceneObject[T]
	circles    []sceneCircleShape[S]
	rectangles []sceneRectangleShape[S]
	polygons   []scenePolygonShape[S]

	circleTree    *CompactTree[uint32]
	rectangleTree *CompactTree[uint32]
	polygonTree   *CompactTree[uint32]

	checks []indexPair
}

const invalidIndexPair = indexPair(0xFFFFFFFFFFFFFFFF)

type indexPair uint64

func (p indexPair) srcIndex() uint32 {
	return uint32(p >> 32)
}

func (p indexPair) tgtIndex() uint32 {
	return uint32(p & 0xFFFFFFFF)
}
