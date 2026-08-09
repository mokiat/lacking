package placement3d

import (
	"iter"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/query3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// SceneSettings contains information needed to create an optimal scene.
type SceneSettings struct {

	// Size specifies the dimension (from side to side) of the scene.
	// Inserting an item outside these bounds has undefined behavior.
	Size opt.T[float64]

	// MaxDepth controls the maximum depth that the underlying octree can reach.
	MaxDepth opt.T[uint32]

	// InitialNodeCapacity is a hint as to the number of nodes that will be
	// needed to place all items. Usually one would find that number empirically.
	// This allows the octree to preallocate memory and avoid dynamic allocations.
	InitialNodeCapacity opt.T[uint32]

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the octree. This allows the octree to preallocate
	// memory and avoid dynamic allocations during insertion.
	InitialItemCapacity opt.T[uint32]
}

// Scene represents a 3D scene into which movable objects (built from convex
// shapes) and immovable terrains (built from concave shapes) can be placed and
// tested for intersection.
//
// The type parameters specify the user data attached to each kind of entity:
// O for objects, T for terrains, and S for the shapes of both.
//
// A scene is not safe for concurrent use. Furthermore, the intersection
// queries share internal scratch buffers, so a query must not be started from
// within the callback of another query.
type Scene[O, T, S any] struct {
	solver *gjk3d.Solver

	objectShapeTree  *query3d.Octree[int32]
	terrainShapeTree *query3d.Octree[int32]

	freeObjectIndices       *ds.Stack[int32]
	freeObjectShapeIndices  *ds.Stack[int32]
	freeTerrainIndices      *ds.Stack[int32]
	freeTerrainShapeIndices *ds.Stack[int32]

	objects       []objectState[O]
	objectShapes  []objectShapeState[S]
	terrains      []terrainState[T]
	terrainShapes []terrainShapeState[S]

	objectShapeCandidates  []int32
	terrainShapeCandidates []int32

	tempGJKSource gjk3d.Shape
	tempGJKTarget gjk3d.Shape
}

// NewScene creates a new scene.
func NewScene[O, T, S any](settings SceneSettings) *Scene[O, T, S] {
	treeSettings := query3d.OctreeSettings(settings)

	return &Scene[O, T, S]{
		solver: gjk3d.NewSolver(),

		objectShapeTree:  query3d.NewOctree[int32](treeSettings),
		terrainShapeTree: query3d.NewOctree[int32](treeSettings),

		freeObjectIndices:       ds.EmptyStack[int32](),
		freeObjectShapeIndices:  ds.EmptyStack[int32](),
		freeTerrainIndices:      ds.EmptyStack[int32](),
		freeTerrainShapeIndices: ds.EmptyStack[int32](),

		objects:       make([]objectState[O], 0),
		objectShapes:  make([]objectShapeState[S], 0),
		terrains:      make([]terrainState[T], 0),
		terrainShapes: make([]terrainShapeState[S], 0),

		objectShapeCandidates:  make([]int32, 0),
		terrainShapeCandidates: make([]int32, 0),

		tempGJKSource: gjk3d.Shape{
			Points: make([]dprec.Vec3, 0, 8),
		},
		tempGJKTarget: gjk3d.Shape{
			Points: make([]dprec.Vec3, 0, 8),
		},
	}
}

// CreateObject creates a new object.
func (s *Scene[O, T, S]) CreateObject(info ObjectInfo[O]) ObjectID {
	transform := shape3d.Transform{
		Translation: info.Position.ValueOrDefault(dprec.ZeroVec3()),
		Rotation: shape3d.RotationFromQuat(
			info.Rotation.ValueOrDefault(dprec.IdentityQuat()),
		),
	}

	index := s.allocateObject()
	s.objects[index] = objectState[O]{
		transform:       transform,
		firstShapeIndex: nilIndex,
		lastShapeIndex:  nilIndex,
		userData:        info.UserData,
	}
	return ObjectID(index)
}

// DeleteObject deletes an object.
func (s *Scene[O, T, S]) DeleteObject(objID ObjectID) {
	index := int32(objID)
	object := &s.objects[index]
	object.userData = gog.Zero[O]() // in case of pointer
	s.eachObjectShape(object, func(shapeIndex int32, _ *objectShapeState[S]) {
		s.detachObjectShape(shapeIndex)
	})
	s.releaseObject(index)
}

// GetObjectUserData returns the user data associated with the given object.
func (s *Scene[O, T, S]) GetObjectUserData(objID ObjectID) O {
	index := int32(objID)
	object := &s.objects[index]
	return object.userData
}

// SetObjectUserData assigns the specified user data to the object.
func (s *Scene[O, T, S]) SetObjectUserData(objID ObjectID, userData O) {
	index := int32(objID)
	object := &s.objects[index]
	object.userData = userData
}

// GetObjectTransform returns the given object's transform.
func (s *Scene[O, T, S]) GetObjectTransform(objID ObjectID) shape3d.Transform {
	index := int32(objID)
	object := &s.objects[index]
	return object.transform
}

// SetObjectTransform relocates the given object.
func (s *Scene[O, T, S]) SetObjectTransform(objID ObjectID, transform shape3d.Transform) {
	index := int32(objID)
	object := &s.objects[index]
	object.transform = transform

	s.eachObjectShape(object, func(_ int32, shape *objectShapeState[S]) {
		shape.update(transform)
		area := shape3d.AABBFromSphere(shape.wsBSphere)
		s.objectShapeTree.Update(shape.spatialID, area)
	})
}

// GetObjectForShape returns the ID of the object that the given shape is
// attached to.
func (s *Scene[O, T, S]) GetObjectForShape(shapeID ObjectShapeID) ObjectID {
	index := int32(shapeID)
	shape := &s.objectShapes[index]
	return ObjectID(shape.objectIndex)
}

// AttachSphere creates a sphere shape and attaches it to the object to be
// used for intersection tests.
//
// The sphere is specified in the local space of the object and moves along
// with it.
func (s *Scene[O, T, S]) AttachSphere(objID ObjectID, info SphereInfo[S]) ObjectShapeID {
	index := int32(objID)

	sphere := info.Sphere
	transform := shape3d.Transform{
		Translation: sphere.Center,
		Rotation:    shape3d.IdentityRotation(),
	}

	return s.attachObjectShape(index, info.Filtering, objectShapeRepresentation{
		lsBSphere:   sphere,
		wsBSphere:   sphere,
		lsTransform: transform,
		wsTransform: transform,
		kind:        objectShapeKindSphere,
		points: []dprec.Vec3{ // TODO: Consider reusing from a buffer.
			dprec.ZeroVec3(),
		},
		skinRadius: sphere.Radius,
	}, info.UserData)
}

// AttachBox creates a box shape and attaches it to the object to be used for
// intersection tests.
//
// The box is specified in the local space of the object and moves along with
// it.
func (s *Scene[O, T, S]) AttachBox(objID ObjectID, info BoxInfo[S]) ObjectShapeID {
	index := int32(objID)

	box := info.Box
	transform := shape3d.Transform{
		Translation: info.Box.Center,
		Rotation:    info.Box.Rotation,
	}
	bSphere := box.BoundingSphere()
	halfWidth := box.HalfWidth
	halfHeight := box.HalfHeight
	halfLength := box.HalfLength

	return s.attachObjectShape(index, info.Filtering, objectShapeRepresentation{
		lsBSphere:   bSphere,
		wsBSphere:   bSphere,
		lsTransform: transform,
		wsTransform: transform,
		kind:        objectShapeKindBox,
		points: []dprec.Vec3{ // TODO: Consider reusing from a buffer.
			dprec.NewVec3(-halfWidth, -halfHeight, -halfLength),
			dprec.NewVec3(halfWidth, -halfHeight, -halfLength),
			dprec.NewVec3(halfWidth, halfHeight, -halfLength),
			dprec.NewVec3(-halfWidth, halfHeight, -halfLength),
			dprec.NewVec3(-halfWidth, -halfHeight, halfLength),
			dprec.NewVec3(halfWidth, -halfHeight, halfLength),
			dprec.NewVec3(halfWidth, halfHeight, halfLength),
			dprec.NewVec3(-halfWidth, halfHeight, halfLength),
		},
		skinRadius: 0.0,
	}, info.UserData)
}

// DeleteObjectShape deletes a shape from an object. The object is not
// deleted and continues to exist in the scene.
func (s *Scene[O, T, S]) DeleteObjectShape(shapeID ObjectShapeID) {
	index := int32(shapeID)
	s.detachObjectShape(index)
}

// GetObjectShapeUserData returns the user data associated with the given
// object shape.
func (s *Scene[O, T, S]) GetObjectShapeUserData(shapeID ObjectShapeID) S {
	index := int32(shapeID)
	shape := &s.objectShapes[index]
	return shape.userData
}

// SetObjectShapeUserData assigns the specified user data to the object shape.
func (s *Scene[O, T, S]) SetObjectShapeUserData(shapeID ObjectShapeID, userData S) {
	index := int32(shapeID)
	shape := &s.objectShapes[index]
	shape.userData = userData
}

// EachSphere iterates over all sphere shapes in the scene that match the mask
// and yields them, in world space, to the provided callback. Iteration stops
// early if the callback returns false.
//
// Note that a zero mask matches no shape at all. Use [FullMask] to iterate
// over every sphere in the scene.
func (s *Scene[O, T, S]) EachSphere(mask Mask, yield func(shape3d.Sphere) bool) {
	for index := range s.objectShapes {
		shape := &s.objectShapes[index]
		if shape.spatialID == query3d.InvalidTreeItemID {
			continue
		}
		if shape.kind != objectShapeKindSphere {
			continue
		}
		if !shape.satisfiesMask(mask) {
			continue
		}
		if !yield(shape.toSphere()) {
			return
		}
	}
}

// SphereIter returns an iterator over all sphere shapes in the scene that
// match the mask, as described by [Scene.EachSphere].
func (s *Scene[O, T, S]) SphereIter(mask Mask) iter.Seq[shape3d.Sphere] {
	return func(yield func(shape3d.Sphere) bool) {
		s.EachSphere(mask, yield)
	}
}

// EachBox iterates over all box shapes in the scene that match the mask and
// yields them, in world space, to the provided callback. Iteration stops early
// if the callback returns false.
//
// Note that a zero mask matches no shape at all. Use [FullMask] to iterate
// over every box in the scene.
func (s *Scene[O, T, S]) EachBox(mask Mask, yield func(shape3d.Box) bool) {
	for index := range s.objectShapes {
		shape := &s.objectShapes[index]
		if shape.spatialID == query3d.InvalidTreeItemID {
			continue
		}
		if shape.kind != objectShapeKindBox {
			continue
		}
		if !shape.satisfiesMask(mask) {
			continue
		}
		if !yield(shape.toBox()) {
			return
		}
	}
}

// BoxIter returns an iterator over all box shapes in the scene that match the
// mask, as described by [Scene.EachBox].
func (s *Scene[O, T, S]) BoxIter(mask Mask) iter.Seq[shape3d.Box] {
	return func(yield func(shape3d.Box) bool) {
		s.EachBox(mask, yield)
	}
}

// CreateTerrain creates a new terrain.
//
// A terrain has no transform of its own. It merely groups the concave shapes
// that are attached to it, which are specified in world space.
func (s *Scene[O, T, S]) CreateTerrain(info TerrainInfo[T]) TerrainID {
	index := s.allocateTerrain()
	s.terrains[index] = terrainState[T]{
		firstShapeIndex: nilIndex,
		lastShapeIndex:  nilIndex,
		userData:        info.UserData,
	}
	return TerrainID(index)
}

// DeleteTerrain deletes a terrain, along with all of the shapes that are
// attached to it.
func (s *Scene[O, T, S]) DeleteTerrain(terrainID TerrainID) {
	index := int32(terrainID)
	terrain := &s.terrains[index]
	terrain.userData = gog.Zero[T]() // in case of pointer
	s.eachTerrainShape(terrain, func(shapeIndex int32, _ *terrainShapeState[S]) {
		s.detachTerrainShape(shapeIndex)
	})
	s.releaseTerrain(index)
}

// GetTerrainUserData returns the user data associated with the given terrain.
func (s *Scene[O, T, S]) GetTerrainUserData(terrainID TerrainID) T {
	index := int32(terrainID)
	terrain := &s.terrains[index]
	return terrain.userData
}

// SetTerrainUserData assigns the specified user data to the terrain.
func (s *Scene[O, T, S]) SetTerrainUserData(terrainID TerrainID, userData T) {
	index := int32(terrainID)
	terrain := &s.terrains[index]
	terrain.userData = userData
}

// GetTerrainForShape returns the ID of the terrain that the given shape is
// attached to.
func (s *Scene[O, T, S]) GetTerrainForShape(shapeID TerrainShapeID) TerrainID {
	index := int32(shapeID)
	shape := &s.terrainShapes[index]
	return TerrainID(shape.terrainIndex)
}

// AttachMesh creates a mesh shape and attaches it to the terrain to be used
// for intersection tests.
//
// The mesh is specified in world space, as terrains have no transform of their
// own, and cannot be relocated afterwards.
//
// The mesh specified through [MeshInfo.Mesh] must not be empty, otherwise this
// function panics.
func (s *Scene[O, T, S]) AttachMesh(terrainID TerrainID, info MeshInfo[S]) TerrainShapeID {
	index := int32(terrainID)

	mesh := info.Mesh
	bSphere := mesh.BoundingSphere()
	aabb := mesh.BoundingAABB()

	return s.attachTerrainShape(index, info.Filtering, terrainShapeRepresentation{
		wsBSphere:   bSphere,
		wsAABB:      aabb,
		wsTriangles: mesh.Triangles,
	}, info.UserData)
}

// DeleteTerrainShape deletes a shape from a terrain. The terrain is not
// deleted and continues to exist in the scene.
func (s *Scene[O, T, S]) DeleteTerrainShape(shapeID TerrainShapeID) {
	index := int32(shapeID)
	s.detachTerrainShape(index)
}

// GetTerrainShapeUserData returns the user data associated with the given
// terrain shape.
func (s *Scene[O, T, S]) GetTerrainShapeUserData(shapeID TerrainShapeID) S {
	index := int32(shapeID)
	shape := &s.terrainShapes[index]
	return shape.userData
}

// SetTerrainShapeUserData assigns the specified user data to the terrain
// shape.
func (s *Scene[O, T, S]) SetTerrainShapeUserData(shapeID TerrainShapeID, userData S) {
	index := int32(shapeID)
	shape := &s.terrainShapes[index]
	shape.userData = userData
}

// CollectSegmentObjectIntersections collects all intersections of the segment
// with the object shapes in the scene that match the mask.
//
// The reported contacts have no source, since the segment is not part of the
// scene. Their Depth is the fraction of the segment that lies beyond the
// contact point, as described by [shape3d.Contact].
func (s *Scene[O, T, S]) CollectSegmentObjectIntersections(segment shape3d.Segment, mask Mask, yield ObjectContactCallback) {
	s.objectShapeCandidates = s.objectShapeCandidates[:0]
	s.objectShapeTree.QuerySegment(segment, func(index int32) bool {
		s.objectShapeCandidates = append(s.objectShapeCandidates, index)
		return true
	})
	s.collectSegmentObject(segment, mask, yield)
}

// CheckSegmentObjectIntersection returns the intersection of the segment with
// the object shape that it enters first, if any.
func (s *Scene[O, T, S]) CheckSegmentObjectIntersection(segment shape3d.Segment, mask Mask) (ObjectContact, bool) {
	var collection DeepestObjectContact
	s.CollectSegmentObjectIntersections(segment, mask, collection.AddContact)
	return collection.Contact()
}

// CollectSegmentTerrainIntersections collects all intersections of the segment
// with the terrain shapes in the scene that match the mask. At most one
// contact is reported per terrain shape.
//
// The reported contacts have no source, since the segment is not part of the
// scene. Their Depth is the fraction of the segment that lies beyond the
// contact point, as described by [shape3d.Contact].
func (s *Scene[O, T, S]) CollectSegmentTerrainIntersections(segment shape3d.Segment, mask Mask, yield TerrainContactCallback) {
	s.terrainShapeCandidates = s.terrainShapeCandidates[:0]
	s.terrainShapeTree.QuerySegment(segment, func(index int32) bool {
		s.terrainShapeCandidates = append(s.terrainShapeCandidates, index)
		return true
	})
	s.collectSegmentTerrain(segment, mask, yield)
}

// CheckSegmentTerrainIntersection returns the intersection of the segment with
// the terrain shape that it enters first, if any.
func (s *Scene[O, T, S]) CheckSegmentTerrainIntersection(segment shape3d.Segment, mask Mask) (TerrainContact, bool) {
	var collection DeepestTerrainContact
	s.CollectSegmentTerrainIntersections(segment, mask, collection.AddContact)
	return collection.Contact()
}

// CollectSphereObjectIntersections collects all intersections of the sphere
// with the object shapes in the scene that match the mask.
//
// The reported contacts have no source, since the sphere is not part of the
// scene.
func (s *Scene[O, T, S]) CollectSphereObjectIntersections(sphere shape3d.Sphere, mask Mask, yield ObjectContactCallback) {
	queryAABB := shape3d.AABBFromSphere(sphere)

	s.objectShapeCandidates = s.objectShapeCandidates[:0]
	s.objectShapeTree.QueryAABB(queryAABB, func(index int32) bool {
		s.objectShapeCandidates = append(s.objectShapeCandidates, index)
		return true
	})
	s.collectSphereObject(sphere, mask, yield)
}

// CheckSphereObjectIntersection returns the deepest intersection of the sphere
// with an object shape in the scene, if any.
func (s *Scene[O, T, S]) CheckSphereObjectIntersection(sphere shape3d.Sphere, mask Mask) (ObjectContact, bool) {
	var collection DeepestObjectContact
	s.CollectSphereObjectIntersections(sphere, mask, collection.AddContact)
	return collection.Contact()
}

// CollectSphereTerrainIntersections collects all intersections of the sphere
// with the terrain shapes in the scene that match the mask. At most one
// contact is reported per terrain shape.
//
// The reported contacts have no source, since the sphere is not part of the
// scene.
func (s *Scene[O, T, S]) CollectSphereTerrainIntersections(sphere shape3d.Sphere, mask Mask, yield TerrainContactCallback) {
	queryAABB := shape3d.AABBFromSphere(sphere)

	s.terrainShapeCandidates = s.terrainShapeCandidates[:0]
	s.terrainShapeTree.QueryAABB(queryAABB, func(index int32) bool {
		s.terrainShapeCandidates = append(s.terrainShapeCandidates, index)
		return true
	})
	s.collectSphereTerrain(sphere, mask, yield)
}

// CheckSphereTerrainIntersection returns the deepest intersection of the
// sphere with a terrain shape in the scene, if any.
func (s *Scene[O, T, S]) CheckSphereTerrainIntersection(sphere shape3d.Sphere, mask Mask) (TerrainContact, bool) {
	var collection DeepestTerrainContact
	s.CollectSphereTerrainIntersections(sphere, mask, collection.AddContact)
	return collection.Contact()
}

// CollectBoxObjectIntersections collects all intersections of the box with the
// object shapes in the scene that match the mask.
//
// The reported contacts have no source, since the box is not part of the
// scene.
func (s *Scene[O, T, S]) CollectBoxObjectIntersections(box shape3d.Box, mask Mask, yield ObjectContactCallback) {
	queryAABB := shape3d.AABBFromBox(box)

	s.objectShapeCandidates = s.objectShapeCandidates[:0]
	s.objectShapeTree.QueryAABB(queryAABB, func(index int32) bool {
		s.objectShapeCandidates = append(s.objectShapeCandidates, index)
		return true
	})
	s.collectBoxObject(box, mask, yield)
}

// CheckBoxObjectIntersection returns the deepest intersection of the box with
// an object shape in the scene, if any.
func (s *Scene[O, T, S]) CheckBoxObjectIntersection(box shape3d.Box, mask Mask) (ObjectContact, bool) {
	var collection DeepestObjectContact
	s.CollectBoxObjectIntersections(box, mask, collection.AddContact)
	return collection.Contact()
}

// CollectBoxTerrainIntersections collects all intersections of the box with
// the terrain shapes in the scene that match the mask. At most one contact is
// reported per terrain shape.
//
// The reported contacts have no source, since the box is not part of the
// scene.
func (s *Scene[O, T, S]) CollectBoxTerrainIntersections(box shape3d.Box, mask Mask, yield TerrainContactCallback) {
	queryAABB := shape3d.AABBFromBox(box)

	s.terrainShapeCandidates = s.terrainShapeCandidates[:0]
	s.terrainShapeTree.QueryAABB(queryAABB, func(index int32) bool {
		s.terrainShapeCandidates = append(s.terrainShapeCandidates, index)
		return true
	})
	s.collectBoxTerrain(box, mask, yield)
}

// CheckBoxTerrainIntersection returns the deepest intersection of the box with
// a terrain shape in the scene, if any.
func (s *Scene[O, T, S]) CheckBoxTerrainIntersection(box shape3d.Box, mask Mask) (TerrainContact, bool) {
	var collection DeepestTerrainContact
	s.CollectBoxTerrainIntersections(box, mask, collection.AddContact)
	return collection.Contact()
}

// CollectObjectIntersections yields the intersections between the object
// shapes in this scene.
//
// Each intersecting pair is reported exactly once, and shapes that belong to
// the same object are never tested against one another. Both the source and
// the target of the reported contacts are object shapes.
func (s *Scene[O, T, S]) CollectObjectIntersections(yield ObjectContactCallback) {
	for i := range s.objectShapes {
		srcIndex := int32(i)
		srcShape := &s.objectShapes[srcIndex]
		if srcShape.spatialID == query3d.InvalidTreeItemID {
			continue
		}

		queryAABB := shape3d.AABBFromSphere(srcShape.wsBSphere)

		s.objectShapeCandidates = s.objectShapeCandidates[:0]
		s.objectShapeTree.QueryAABB(queryAABB, func(tgtIndex int32) bool {
			s.objectShapeCandidates = append(s.objectShapeCandidates, tgtIndex)
			return true
		})
		s.collectObjectObject(srcIndex, srcShape, yield)
	}
}

// CollectTerrainIntersections yields the intersections between the object
// shapes and the terrain shapes in this scene. At most one contact is reported
// per object shape and terrain shape pair.
//
// Terrain shapes are never tested against one another, since terrains cannot
// move. The source of the reported contacts is therefore always an object
// shape.
func (s *Scene[O, T, S]) CollectTerrainIntersections(yield TerrainContactCallback) {
	for i := range s.objectShapes {
		srcIndex := int32(i)
		srcShape := &s.objectShapes[srcIndex]
		if srcShape.spatialID == query3d.InvalidTreeItemID {
			continue
		}

		queryAABB := shape3d.AABBFromSphere(srcShape.wsBSphere)

		s.terrainShapeCandidates = s.terrainShapeCandidates[:0]
		s.terrainShapeTree.QueryAABB(queryAABB, func(tgtIndex int32) bool {
			s.terrainShapeCandidates = append(s.terrainShapeCandidates, tgtIndex)
			return true
		})
		s.collectObjectTerrain(srcIndex, srcShape, yield)
	}
}

const nilIndex = -1

func (s *Scene[O, T, S]) allocateObject() int32 {
	if s.freeObjectIndices.IsEmpty() {
		index := len(s.objects)
		s.objects = append(s.objects, objectState[O]{})
		return int32(index)
	} else {
		return s.freeObjectIndices.Pop()
	}
}

func (s *Scene[O, T, S]) releaseObject(index int32) {
	s.freeObjectIndices.Push(index)
}

func (s *Scene[O, T, S]) eachObjectShape(object *objectState[O], cb func(int32, *objectShapeState[S])) {
	index := object.firstShapeIndex
	for index >= 0 {
		shape := &s.objectShapes[index]
		nextIndex := shape.nextShapeIndex
		cb(index, shape)
		index = nextIndex
	}
}

func (s *Scene[O, T, S]) allocateObjectShape() int32 {
	if s.freeObjectShapeIndices.IsEmpty() {
		index := len(s.objectShapes)
		s.objectShapes = append(s.objectShapes, objectShapeState[S]{})
		return int32(index)
	} else {
		return s.freeObjectShapeIndices.Pop()
	}
}

func (s *Scene[O, T, S]) releaseObjectShape(index int32) {
	s.freeObjectShapeIndices.Push(index)
}

func (s *Scene[O, T, S]) attachObjectShape(
	objectIndex int32,
	filterInfo FilterInfo,
	representation objectShapeRepresentation,
	userData S,
) ObjectShapeID {

	object := &s.objects[objectIndex]
	index := s.allocateObjectShape()

	representation.update(object.transform)
	area := shape3d.AABBFromSphere(representation.wsBSphere)

	s.objectShapes[index] = objectShapeState[S]{
		objectIndex:               objectIndex,
		nextShapeIndex:            nilIndex,
		prevShapeIndex:            object.lastShapeIndex,
		spatialID:                 s.objectShapeTree.Insert(area, index),
		filterRepresentation:      newFilterRepresentation(filterInfo),
		objectShapeRepresentation: representation,
		userData:                  userData,
	}
	if object.firstShapeIndex == nilIndex {
		object.firstShapeIndex = index
	} else {
		s.objectShapes[object.lastShapeIndex].nextShapeIndex = index
	}
	object.lastShapeIndex = index

	return ObjectShapeID(index)
}

func (s *Scene[O, T, S]) detachObjectShape(index int32) {
	shape := &s.objectShapes[index]

	s.objectShapeTree.Remove(shape.spatialID)
	shape.spatialID = query3d.InvalidTreeItemID

	object := &s.objects[shape.objectIndex]
	if object.firstShapeIndex == index {
		object.firstShapeIndex = shape.nextShapeIndex
	}
	if object.lastShapeIndex == index {
		object.lastShapeIndex = shape.prevShapeIndex
	}
	if shape.prevShapeIndex != nilIndex {
		prevShape := &s.objectShapes[shape.prevShapeIndex]
		prevShape.nextShapeIndex = shape.nextShapeIndex
	}
	if shape.nextShapeIndex != nilIndex {
		nextShape := &s.objectShapes[shape.nextShapeIndex]
		nextShape.prevShapeIndex = shape.prevShapeIndex
	}
	shape.objectIndex = -1
	shape.userData = gog.Zero[S]() // in case of pointer

	s.releaseObjectShape(index)
}

func (s *Scene[O, T, S]) allocateTerrain() int32 {
	if s.freeTerrainIndices.IsEmpty() {
		index := len(s.terrains)
		s.terrains = append(s.terrains, terrainState[T]{})
		return int32(index)
	} else {
		return s.freeTerrainIndices.Pop()
	}
}

func (s *Scene[O, T, S]) releaseTerrain(index int32) {
	s.freeTerrainIndices.Push(index)
}

func (s *Scene[O, T, S]) eachTerrainShape(terrain *terrainState[T], cb func(int32, *terrainShapeState[S])) {
	index := terrain.firstShapeIndex
	for index >= 0 {
		shape := &s.terrainShapes[index]
		nextIndex := shape.nextShapeIndex
		cb(index, shape)
		index = nextIndex
	}
}

func (s *Scene[O, T, S]) allocateTerrainShape() int32 {
	if s.freeTerrainShapeIndices.IsEmpty() {
		index := len(s.terrainShapes)
		s.terrainShapes = append(s.terrainShapes, terrainShapeState[S]{})
		return int32(index)
	} else {
		return s.freeTerrainShapeIndices.Pop()
	}
}

func (s *Scene[O, T, S]) releaseTerrainShape(index int32) {
	s.freeTerrainShapeIndices.Push(index)
}

func (s *Scene[O, T, S]) attachTerrainShape(
	terrainIndex int32,
	filterInfo FilterInfo,
	representation terrainShapeRepresentation,
	userData S,
) TerrainShapeID {

	terrain := &s.terrains[terrainIndex]
	index := s.allocateTerrainShape()

	area := representation.wsAABB

	s.terrainShapes[index] = terrainShapeState[S]{
		terrainIndex:               terrainIndex,
		nextShapeIndex:             nilIndex,
		prevShapeIndex:             terrain.lastShapeIndex,
		spatialID:                  s.terrainShapeTree.Insert(area, index),
		filterRepresentation:       newFilterRepresentation(filterInfo),
		terrainShapeRepresentation: representation,
		userData:                   userData,
	}
	if terrain.firstShapeIndex == nilIndex {
		terrain.firstShapeIndex = index
	} else {
		s.terrainShapes[terrain.lastShapeIndex].nextShapeIndex = index
	}
	terrain.lastShapeIndex = index

	return TerrainShapeID(index)
}

func (s *Scene[O, T, S]) detachTerrainShape(index int32) {
	shape := &s.terrainShapes[index]

	s.terrainShapeTree.Remove(shape.spatialID)
	shape.spatialID = query3d.InvalidTreeItemID

	terrain := &s.terrains[shape.terrainIndex]
	if terrain.firstShapeIndex == index {
		terrain.firstShapeIndex = shape.nextShapeIndex
	}
	if terrain.lastShapeIndex == index {
		terrain.lastShapeIndex = shape.prevShapeIndex
	}
	if shape.prevShapeIndex != nilIndex {
		prevShape := &s.terrainShapes[shape.prevShapeIndex]
		prevShape.nextShapeIndex = shape.nextShapeIndex
	}
	if shape.nextShapeIndex != nilIndex {
		nextShape := &s.terrainShapes[shape.nextShapeIndex]
		nextShape.prevShapeIndex = shape.prevShapeIndex
	}
	shape.terrainIndex = -1
	shape.userData = gog.Zero[S]() // in case of pointer

	s.releaseTerrainShape(index)
}

func (s *Scene[O, T, S]) collectSegmentObject(segment shape3d.Segment, mask Mask, yield ObjectContactCallback) {
	for index, shape := range s.iterCandidateObjectShapes(mask) {
		if !isec3d.CheckSegmentSphereOverlap(segment, shape.wsBSphere) {
			continue
		}
		onContact := func(contact shape3d.Contact) {
			yield(ObjectContact{
				SourceObjectID: NilObjectID,
				SourceShapeID:  NilObjectShapeID,
				TargetObjectID: ObjectID(shape.objectIndex),
				TargetShapeID:  ObjectShapeID(index),
				Contact:        contact,
			})
		}
		switch shape.kind {
		case objectShapeKindSphere:
			sphere := shape.toSphere()
			isec3d.ResolveSegmentSphere(segment, sphere, onContact)
		case objectShapeKindBox:
			box := shape.toBox()
			isec3d.ResolveSegmentBox(segment, box, onContact)
		}
	}
}

func (s *Scene[O, T, S]) collectSegmentTerrain(segment shape3d.Segment, mask Mask, yield TerrainContactCallback) {
	for index, shape := range s.iterCandidateTerrainShapes(mask) {
		if !isec3d.CheckSegmentSphereOverlap(segment, shape.wsBSphere) {
			continue
		}
		var deepestContact shape3d.DeepestContact
		for _, triangle := range shape.wsTriangles {
			isec3d.ResolveSegmentTriangle(segment, triangle, deepestContact.AddContact)
		}
		if contact, ok := deepestContact.Contact(); ok {
			yield(TerrainContact{
				SourceObjectID:  NilObjectID,
				SourceShapeID:   NilObjectShapeID,
				TargetTerrainID: TerrainID(shape.terrainIndex),
				TargetShapeID:   TerrainShapeID(index),
				Contact:         contact,
			})
		}
	}
}

func (s *Scene[O, T, S]) collectSphereObject(sphere shape3d.Sphere, mask Mask, yield ObjectContactCallback) {
	initGJKShapeForSphere(sphere, &s.tempGJKSource)
	for index, shape := range s.iterCandidateObjectShapes(mask) {
		if !isec3d.CheckSphereSphere(sphere, shape.wsBSphere) {
			continue
		}
		if contact, ok := s.solver.Resolve(s.tempGJKSource, shape.gjkShape()); ok {
			yield(ObjectContact{
				SourceObjectID: NilObjectID,
				SourceShapeID:  NilObjectShapeID,
				TargetObjectID: ObjectID(shape.objectIndex),
				TargetShapeID:  ObjectShapeID(index),
				Contact:        contact,
			})
		}
	}
}

func (s *Scene[O, T, S]) collectSphereTerrain(sphere shape3d.Sphere, mask Mask, yield TerrainContactCallback) {
	initGJKShapeForSphere(sphere, &s.tempGJKSource)
	for index, shape := range s.iterCandidateTerrainShapes(mask) {
		s.resolveTerrainShape(s.tempGJKSource, sphere, shape, func(contact shape3d.Contact) {
			yield(TerrainContact{
				SourceObjectID:  NilObjectID,
				SourceShapeID:   NilObjectShapeID,
				TargetTerrainID: TerrainID(shape.terrainIndex),
				TargetShapeID:   TerrainShapeID(index),
				Contact:         contact,
			})
		})
	}
}

func (s *Scene[O, T, S]) collectBoxObject(box shape3d.Box, mask Mask, yield ObjectContactCallback) {
	initGJKShapeForBox(box, &s.tempGJKSource)
	for index, shape := range s.iterCandidateObjectShapes(mask) {
		if !isec3d.CheckSphereSphere(box.BoundingSphere(), shape.wsBSphere) {
			continue
		}
		if contact, ok := s.solver.Resolve(s.tempGJKSource, shape.gjkShape()); ok {
			yield(ObjectContact{
				SourceObjectID: NilObjectID,
				SourceShapeID:  NilObjectShapeID,
				TargetObjectID: ObjectID(shape.objectIndex),
				TargetShapeID:  ObjectShapeID(index),
				Contact:        contact,
			})
		}
	}
}

func (s *Scene[O, T, S]) collectBoxTerrain(box shape3d.Box, mask Mask, yield TerrainContactCallback) {
	initGJKShapeForBox(box, &s.tempGJKSource)
	for index, shape := range s.iterCandidateTerrainShapes(mask) {
		s.resolveTerrainShape(s.tempGJKSource, box.BoundingSphere(), shape, func(contact shape3d.Contact) {
			yield(TerrainContact{
				SourceObjectID:  NilObjectID,
				SourceShapeID:   NilObjectShapeID,
				TargetTerrainID: TerrainID(shape.terrainIndex),
				TargetShapeID:   TerrainShapeID(index),
				Contact:         contact,
			})
		})
	}
}

func (s *Scene[O, T, S]) collectObjectObject(srcIndex int32, srcShape *objectShapeState[S], yield ObjectContactCallback) {
	srcGJKShape := srcShape.gjkShape()
	for _, tgtIndex := range s.objectShapeCandidates {
		tgtShape := &s.objectShapes[tgtIndex]
		if !objectShapesCanIntersect(srcShape, tgtShape) {
			continue
		}
		if !isec3d.CheckSphereSphere(srcShape.wsBSphere, tgtShape.wsBSphere) {
			continue
		}
		if contact, ok := s.solver.Resolve(srcGJKShape, tgtShape.gjkShape()); ok {
			yield(ObjectContact{
				SourceObjectID: ObjectID(srcShape.objectIndex),
				SourceShapeID:  ObjectShapeID(srcIndex),
				TargetObjectID: ObjectID(tgtShape.objectIndex),
				TargetShapeID:  ObjectShapeID(tgtIndex),
				Contact:        contact,
			})
		}
	}
}

func (s *Scene[O, T, S]) collectObjectTerrain(srcIndex int32, srcShape *objectShapeState[S], yield TerrainContactCallback) {
	srcGJKShape := srcShape.gjkShape()
	for _, tgtIndex := range s.terrainShapeCandidates {
		tgtShape := &s.terrainShapes[tgtIndex]
		if !objectTerrainShapesCanIntersect(srcShape, tgtShape) {
			continue
		}
		s.resolveTerrainShape(srcGJKShape, srcShape.wsBSphere, tgtShape, func(contact shape3d.Contact) {
			yield(TerrainContact{
				SourceObjectID:  ObjectID(srcShape.objectIndex),
				SourceShapeID:   ObjectShapeID(srcIndex),
				TargetTerrainID: TerrainID(tgtShape.terrainIndex),
				TargetShapeID:   TerrainShapeID(tgtIndex),
				Contact:         contact,
			})
		})
	}
}

// resolveTerrainShape resolves the intersection of the specified convex source
// shape with a terrain shape, by testing it against each of the triangles that
// make up the terrain shape.
//
// Only the deepest contact is yielded, and at most once, so that a source
// shape that overlaps many triangles of the same terrain shape does not
// produce a pile of nearly identical contacts.
func (s *Scene[O, T, S]) resolveTerrainShape(srcGJK gjk3d.Shape, srcBS shape3d.Sphere, tgtShape *terrainShapeState[S], yield shape3d.ContactCallback) {
	if !isec3d.CheckSphereSphere(srcBS, tgtShape.wsBSphere) {
		return
	}
	points := initGJKShapeForMesh(&s.tempGJKTarget)
	var deepestContact shape3d.DeepestContact
	for _, triangle := range tgtShape.wsTriangles {
		tgtBSphere := triangle.BoundingSphere()
		if !isec3d.CheckSphereSphere(srcBS, tgtBSphere) {
			continue
		}
		points[0] = triangle.A
		points[1] = triangle.B
		points[2] = triangle.C
		if contact, ok := s.solver.Resolve(srcGJK, s.tempGJKTarget); ok {
			// Prevent contacts that try to push the source shape into the triangle.
			if dprec.Vec3Dot(contact.TargetNormal, triangle.Normal()) > 0 {
				deepestContact.AddContact(contact)
			}
		}
	}
	if contact, ok := deepestContact.Contact(); ok {
		yield(contact)
	}
}

func (s *Scene[O, T, S]) eachCandidateObjectShape(mask Mask, cb func(int32, *objectShapeState[S]) bool) {
	for _, index := range s.objectShapeCandidates {
		shape := &s.objectShapes[index]
		if !shape.satisfiesMask(mask) {
			continue
		}
		if !cb(index, shape) {
			return
		}
	}
}

func (s *Scene[O, T, S]) iterCandidateObjectShapes(mask Mask) iter.Seq2[int32, *objectShapeState[S]] {
	return func(yield func(int32, *objectShapeState[S]) bool) {
		s.eachCandidateObjectShape(mask, yield)
	}
}

func (s *Scene[O, T, S]) eachCandidateTerrainShape(mask Mask, cb func(int32, *terrainShapeState[S]) bool) {
	for _, index := range s.terrainShapeCandidates {
		shape := &s.terrainShapes[index]
		if !shape.satisfiesMask(mask) {
			continue
		}
		if !cb(index, shape) {
			return
		}
	}
}

func (s *Scene[O, T, S]) iterCandidateTerrainShapes(mask Mask) iter.Seq2[int32, *terrainShapeState[S]] {
	return func(yield func(int32, *terrainShapeState[S]) bool) {
		s.eachCandidateTerrainShape(mask, yield)
	}
}

func initGJKShapeForSphere(sphere shape3d.Sphere, out *gjk3d.Shape) {
	out.Position = sphere.Center
	out.Rotation = shape3d.IdentityRotation()
	out.Points = out.Points[:1]
	out.Points[0] = dprec.ZeroVec3()
	out.SkinRadius = sphere.Radius
}

func initGJKShapeForBox(box shape3d.Box, out *gjk3d.Shape) {
	out.Position = box.Center
	out.Rotation = box.Rotation
	out.Points = out.Points[:8]
	halfWidth := box.HalfWidth
	halfHeight := box.HalfHeight
	halfLength := box.HalfLength
	out.Points[0] = dprec.NewVec3(-halfWidth, -halfHeight, -halfLength)
	out.Points[1] = dprec.NewVec3(halfWidth, -halfHeight, -halfLength)
	out.Points[2] = dprec.NewVec3(halfWidth, halfHeight, -halfLength)
	out.Points[3] = dprec.NewVec3(-halfWidth, halfHeight, -halfLength)
	out.Points[4] = dprec.NewVec3(-halfWidth, -halfHeight, halfLength)
	out.Points[5] = dprec.NewVec3(halfWidth, -halfHeight, halfLength)
	out.Points[6] = dprec.NewVec3(halfWidth, halfHeight, halfLength)
	out.Points[7] = dprec.NewVec3(-halfWidth, halfHeight, halfLength)
	out.SkinRadius = 0.0
}

func initGJKShapeForMesh(out *gjk3d.Shape) []dprec.Vec3 {
	out.Position = dprec.ZeroVec3()
	out.Rotation = shape3d.IdentityRotation()
	out.Points = out.Points[:3]
	out.SkinRadius = 0.0
	return out.Points
}
