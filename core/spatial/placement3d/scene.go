package placement3d

import (
	"iter"
	"slices"

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

// Scene represents a 3D scene where objects made of shapes can be added.
type Scene[O, S any] struct {
	staticTree  *query3d.Octree[shapeRef]
	dynamicTree *query3d.Octree[shapeRef]

	solver *gjk3d.Solver

	freeObjectIndices      *ds.Stack[int32]
	freeConvexShapeIndices *ds.Stack[int32]
	freeMeshShapeIndices   *ds.Stack[int32]

	objects      []sceneObject[O]
	convexShapes []convexShape[S]
	meshShapes   []meshShape[S]

	convexCandidates []int32
	meshCandidates   []int32
	pairCandidates   []shapeRefPair
}

// NewScene creates a new scene.
func NewScene[O, S any](settings SceneSettings) *Scene[O, S] {
	treeSettings := query3d.OctreeSettings(settings)

	return &Scene[O, S]{
		staticTree:  query3d.NewOctree[shapeRef](treeSettings),
		dynamicTree: query3d.NewOctree[shapeRef](treeSettings),

		solver: gjk3d.NewSolver(),

		freeObjectIndices:      ds.EmptyStack[int32](),
		freeConvexShapeIndices: ds.EmptyStack[int32](),
		freeMeshShapeIndices:   ds.EmptyStack[int32](),

		objects:      make([]sceneObject[O], 0),
		convexShapes: make([]convexShape[S], 0),
		meshShapes:   make([]meshShape[S], 0),

		convexCandidates: make([]int32, 0),
		meshCandidates:   make([]int32, 0),
		pairCandidates:   make([]shapeRefPair, 0),
	}
}

// CreateObject creates a new object.
func (s *Scene[O, S]) CreateObject(info ObjectInfo[O]) ObjectID {
	transform := shape3d.Transform{
		Translation: info.Position.ValueOrDefault(dprec.ZeroVec3()),
		Rotation: shape3d.RotationFromQuat(
			info.Rotation.ValueOrDefault(dprec.IdentityQuat()),
		),
	}

	index := s.allocateObject()
	s.objects[index] = sceneObject[O]{
		transform:        transform,
		firstConvexShape: invalidShapeRef,
		firstMeshShape:   invalidShapeRef,
		static:           info.Static,
		userData:         info.UserData,
	}
	return ObjectID(index)
}

// DeleteObject deletes an object.
func (s *Scene[O, S]) DeleteObject(objID ObjectID) {
	index := int32(objID)
	object := &s.objects[index]
	object.userData = gog.Zero[O]() // in case of pointer
	s.deleteObjectConvexShapes(object)
	s.deleteObjectMeshShapes(object)
	s.releaseObject(index)
}

// GetObjectUserData returns the user data associated with the given object.
func (s *Scene[O, S]) GetObjectUserData(objID ObjectID) O {
	object := &s.objects[objID]
	return object.userData
}

// SetObjectUserData assigns the specified user data to the object.
func (s *Scene[O, S]) SetObjectUserData(objID ObjectID, userData O) {
	object := &s.objects[objID]
	object.userData = userData
}

// GetObjectTransform returns the given object's transform.
func (s *Scene[O, S]) GetObjectTransform(objID ObjectID) shape3d.Transform {
	object := &s.objects[objID]
	return object.transform
}

// SetObjectTransform relocates the given object.
func (s *Scene[O, S]) SetObjectTransform(objID ObjectID, transform shape3d.Transform) {
	object := &s.objects[objID]
	object.transform = transform

	tree := gog.Ternary(object.static, s.staticTree, s.dynamicTree)
	s.eachObjectConvexShape(object, func(shape *convexShape[S]) {
		shape.update(transform)
		bs := shape.boundingSphere()
		tree.Update(shape.spatialID, query3d.AreaFromSphere(bs))
	})
	s.eachObjectMeshShape(object, func(shape *meshShape[S]) {
		shape.update(transform)
		bs := shape.boundingSphere()
		tree.Update(shape.spatialID, query3d.AreaFromSphere(bs))
	})
}

// AttachSphere creates a sphere shape and attaches it to the object to be
// used for intersection tests.
func (s *Scene[O, S]) AttachSphere(objID ObjectID, info SphereInfo[S]) ShapeID {
	sphere := info.Sphere
	transform := shape3d.Transform{
		Translation: sphere.Center,
		Rotation:    shape3d.IdentityRotation(),
	}

	return s.attachConvexShape(int32(objID), info.ShapeInfo, convexShapeRepresentation{
		lsBSphere:   sphere,
		wsBSphere:   sphere,
		lsTransform: transform,
		wsTransform: transform,
		kind:        convexKindSphere,
		points: []dprec.Vec3{ // TODO: Consider reusing from a buffer.
			dprec.ZeroVec3(),
		},
		skinRadius: sphere.Radius,
	})
}

// AttachBox creates a box shape and attaches it to the object to be used for
// intersection tests.
func (s *Scene[O, S]) AttachBox(objID ObjectID, info BoxInfo[S]) ShapeID {
	box := info.Box
	transform := shape3d.Transform{
		Translation: info.Box.Center,
		Rotation:    info.Box.Rotation,
	}
	bSphere := box.BoundingSphere()
	halfWidth := box.HalfWidth
	halfHeight := box.HalfHeight
	halfLength := box.HalfLength

	return s.attachConvexShape(int32(objID), info.ShapeInfo, convexShapeRepresentation{
		lsBSphere:   bSphere,
		wsBSphere:   bSphere,
		lsTransform: transform,
		wsTransform: transform,
		kind:        convexKindBox,
		points: []dprec.Vec3{
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
	})
}

// AttachMesh creates a mesh shape and attaches it to the object to be used for
// intersection tests.
func (s *Scene[O, S]) AttachMesh(objID ObjectID, info MeshInfo[S]) ShapeID {
	bSphere := info.Mesh.BoundingSphere()
	triangles := info.Mesh.Triangles

	return s.attachMeshShape(int32(objID), info.ShapeInfo, meshShapeRepresentation{
		lsBSphere:   bSphere,
		wsBSphere:   bSphere,
		lsTriangles: triangles,
		wsTriangles: slices.Clone(triangles),
		points:      [3]dprec.Vec3{}, // just used for GJK to avoid allocations
	})
}

// DeleteShape deletes a shape from an object. The object is not
// deleted and continues to exist in the scene.
func (s *Scene[O, S]) DeleteShape(shapeID ShapeID) {
	ref := shapeRef(shapeID)

	if ref.isMesh() {
		s.detachMeshShape(ref.index())
	} else {
		s.detachConvexShape(ref.index())
	}
}

// GetShapeUserData returns the user data associated with the given shape.
func (s *Scene[O, S]) GetShapeUserData(shapeID ShapeID) S {
	ref := shapeRef(shapeID)
	if ref.isMesh() {
		shape := &s.meshShapes[ref.index()]
		return shape.userData
	} else {
		shape := &s.convexShapes[ref.index()]
		return shape.userData
	}
}

// SetShapeUserData assigns the specified user data to the shape.
func (s *Scene[O, S]) SetShapeUserData(shapeID ShapeID, userData S) {
	ref := shapeRef(shapeID)
	if ref.isMesh() {
		shape := &s.meshShapes[ref.index()]
		shape.userData = userData
	} else {
		shape := &s.convexShapes[ref.index()]
		shape.userData = userData
	}
}

// EachSphere iterates over all sphere shapes in the scene that match the
// filter and yields them to the provided callback.
func (s *Scene[O, S]) EachSphere(filter Filter, yield func(shape3d.Sphere) bool) {
	for index := range s.convexShapes {
		shape := &s.convexShapes[index]
		if shape.spatialID == query3d.InvalidTreeItemID {
			continue
		}
		if shape.kind != convexKindSphere {
			continue
		}
		if !shape.matchesFilter(filter) {
			continue
		}
		if !yield(shape.toSphere()) {
			return
		}
	}
}

// SphereIter returns an iterator over all sphere shapes in the scene that match
// the filter.
func (s *Scene[O, S]) SphereIter(filter Filter) iter.Seq[shape3d.Sphere] {
	return func(yield func(shape3d.Sphere) bool) {
		s.EachSphere(filter, yield)
	}
}

// EachBox iterates over all box shapes in the scene that match the
// filter and yields them to the provided callback.
func (s *Scene[O, S]) EachBox(filter Filter, yield func(shape3d.Box) bool) {
	for index := range s.convexShapes {
		shape := &s.convexShapes[index]
		if shape.spatialID == query3d.InvalidTreeItemID {
			continue
		}
		if shape.kind != convexKindBox {
			continue
		}
		if !shape.matchesFilter(filter) {
			continue
		}
		if !yield(shape.toBox()) {
			return
		}
	}
}

// BoxIter returns an iterator over all box shapes in the scene that match
// the filter.
func (s *Scene[O, S]) BoxIter(filter Filter) iter.Seq[shape3d.Box] {
	return func(yield func(shape3d.Box) bool) {
		s.EachBox(filter, yield)
	}
}

// EachMesh iterates over all mesh shapes in the scene that match the
// filter and yields them to the provided callback.
func (s *Scene[O, S]) EachMesh(filter Filter, yield func(shape3d.Mesh) bool) {
	for index := range s.meshShapes {
		shape := &s.meshShapes[index]
		if shape.spatialID == query3d.InvalidTreeItemID {
			continue
		}
		if !shape.matchesFilter(filter) {
			continue
		}
		if !yield(shape3d.NewMesh(shape.wsTriangles)) {
			return
		}
	}
}

// MeshIter returns an iterator over all mesh shapes in the scene that match
// the filter.
func (s *Scene[O, S]) MeshIter(filter Filter) iter.Seq[shape3d.Mesh] {
	return func(yield func(shape3d.Mesh) bool) {
		s.EachMesh(filter, yield)
	}
}

// CheckSegmentIntersection returns the deepest intersection of the segment
// with the scene.
func (s *Scene[O, S]) CheckSegmentIntersection(segment shape3d.Segment, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectSegmentIntersections(segment, filter, collection.AddContact)
	return collection.Contact()
}

// CollectSegmentIntersections collects all intersections of the segment
// with objects in the scene.
func (s *Scene[O, S]) CollectSegmentIntersections(segment shape3d.Segment, filter Filter, yield ContactCallback) {
	querySegment := query3d.NewSegment(segment.A, segment.B)

	s.convexCandidates = s.convexCandidates[:0]
	s.meshCandidates = s.meshCandidates[:0]

	if !filter.SkipDynamic {
		s.dynamicTree.QuerySegment(querySegment, func(tgtRef shapeRef) bool {
			if tgtRef.isMesh() {
				s.meshCandidates = append(s.meshCandidates, tgtRef.index())
			} else {
				s.convexCandidates = append(s.convexCandidates, tgtRef.index())
			}
			return true
		})
	}

	if !filter.SkipStatic {
		s.staticTree.QuerySegment(querySegment, func(tgtRef shapeRef) bool {
			if tgtRef.isMesh() {
				s.meshCandidates = append(s.meshCandidates, tgtRef.index())
			} else {
				s.convexCandidates = append(s.convexCandidates, tgtRef.index())
			}
			return true
		})
	}

	s.resolveSegmentConvex(segment, filter, yield)
	s.resolveSegmentMesh(segment, filter, yield)
}

// CollectIntersections yields intersections found in this scene.
func (s *Scene[O, S]) CollectIntersections(yield ContactCallback) {
	s.pairCandidates = s.pairCandidates[:0]

	s.eachDynamicConvexShape(func(srcIndex int32, shape *convexShape[S]) {
		srcRef := newShapeRef(srcIndex, false)
		queryAABB := query3d.AABBFromSphere(shape.boundingSphere())
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.pairCandidates = append(s.pairCandidates, newShapeRefPair(srcRef, tgtRef))
			return true
		})
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.pairCandidates = append(s.pairCandidates, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	})

	s.eachDynamicMeshShape(func(srcIndex int32, shape *meshShape[S]) {
		srcRef := newShapeRef(srcIndex, true)
		queryAABB := query3d.AABBFromSphere(shape.boundingSphere())
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.pairCandidates = append(s.pairCandidates, newShapeRefPair(srcRef, tgtRef))
			return true
		})
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.pairCandidates = append(s.pairCandidates, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	})

	s.resolvePairs(yield)
}

// // CheckSphereIntersection returns the deepest intersection of the sphere
// // with the scene.
// func (s *Scene[O, S]) CheckSphereIntersection(sphere shape3d.Sphere, filter Filter) (Contact, bool) {
// 	var collection DeepestContact
// 	s.CollectSphereIntersections(sphere, filter, collection.AddContact)
// 	return collection.Contact()
// }

// // CollectSphereIntersections collects all intersections of the sphere
// // with objects in the scene.
// func (s *Scene[O, S]) CollectSphereIntersections(sphere shape3d.Sphere, filter Filter, yield ContactCallback) {
// 	s.tempShape = sceneShape[S]{
// 		objectIndex: invalidObjectIndex,
// 		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
// 		static:      true, // important, otherwise double-check prevention will kick in
// 	}
// 	s.tempSphere = newSphereSolver(sphere)
// 	srcRef := newTempShapeRef(shapeKindSphere)

// 	s.pairCandidates = s.pairCandidates[:0]
// 	queryAABB := query3d.AABBFromSphere(sphere)
// 	if !filter.SkipDynamic {
// 		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
// 			s.pairCandidates = append(s.pairCandidates, newShapeRefPair(srcRef, tgtRef))
// 			return true
// 		})
// 	}
// 	if !filter.SkipStatic {
// 		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
// 			s.pairCandidates = append(s.pairCandidates, newShapeRefPair(srcRef, tgtRef))
// 			return true
// 		})
// 	}
// 	s.collectIntersections(yield)
// }

// // CheckBoxIntersection returns the deepest intersection of the box
// // with the scene.
// func (s *Scene[O, S]) CheckBoxIntersection(box shape3d.Box, filter Filter) (Contact, bool) {
// 	var collection DeepestContact
// 	s.CollectBoxIntersections(box, filter, collection.AddContact)
// 	return collection.Contact()
// }

// // CollectBoxIntersections collects all intersections of the box
// // with objects in the scene.
// func (s *Scene[O, S]) CollectBoxIntersections(box shape3d.Box, filter Filter, yield ContactCallback) {
// 	s.tempShape = sceneShape[S]{
// 		objectIndex: invalidObjectIndex,
// 		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
// 		static:      true, // important, otherwise double-check prevention will kick in
// 	}
// 	s.tempBox = newBoxSolver(box)
// 	srcRef := newTempShapeRef(shapeKindBox)

// 	s.pairCandidates = s.pairCandidates[:0]
// 	queryAABB := query3d.AABBFromSphere(box.BoundingSphere())
// 	if !filter.SkipDynamic {
// 		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
// 			s.pairCandidates = append(s.pairCandidates, newShapeRefPair(srcRef, tgtRef))
// 			return true
// 		})
// 	}
// 	if !filter.SkipStatic {
// 		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
// 			s.pairCandidates = append(s.pairCandidates, newShapeRefPair(srcRef, tgtRef))
// 			return true
// 		})
// 	}
// 	s.collectIntersections(yield)
// }

// // CheckMeshIntersection returns the deepest intersection of the mesh
// // with the scene.
// func (s *Scene[O, S]) CheckMeshIntersection(mesh shape3d.Mesh, filter Filter) (Contact, bool) {
// 	var collection DeepestContact
// 	s.CollectMeshIntersections(mesh, filter, collection.AddContact)
// 	return collection.Contact()
// }

// // CollectMeshIntersections collects all intersections of the mesh
// // with objects in the scene.
// func (s *Scene[O, S]) CollectMeshIntersections(mesh shape3d.Mesh, filter Filter, yield ContactCallback) {
// 	s.tempShape = sceneShape[S]{
// 		objectIndex: invalidObjectIndex,
// 		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
// 		static:      true, // important, otherwise double-check prevention will kick in
// 	}
// 	s.tempMesh = newMeshSolver(mesh)
// 	srcRef := newTempShapeRef(shapeKindMesh)

// 	s.pairCandidates = s.pairCandidates[:0]
// 	queryAABB := query3d.AABBFromSphere(mesh.BoundingSphere())
// 	if !filter.SkipDynamic {
// 		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
// 			s.pairCandidates = append(s.pairCandidates, newShapeRefPair(srcRef, tgtRef))
// 			return true
// 		})
// 	}
// 	if !filter.SkipStatic {
// 		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
// 			s.pairCandidates = append(s.pairCandidates, newShapeRefPair(srcRef, tgtRef))
// 			return true
// 		})
// 	}
// 	s.collectIntersections(yield)
// }

func (s *Scene[O, S]) allocateObject() int32 {
	if s.freeObjectIndices.IsEmpty() {
		index := len(s.objects)
		s.objects = append(s.objects, sceneObject[O]{})
		return int32(index)
	} else {
		return s.freeObjectIndices.Pop()
	}
}

func (s *Scene[O, S]) releaseObject(index int32) {
	s.freeObjectIndices.Push(index)
}

func (s *Scene[O, S]) eachObjectConvexShape(object *sceneObject[O], cb func(*convexShape[S])) {
	index := object.firstConvexShape
	for index >= 0 {
		shape := &s.convexShapes[index]
		nextIndex := shape.nextShape
		cb(shape)
		index = nextIndex
	}
}

func (s *Scene[O, S]) eachObjectMeshShape(object *sceneObject[O], cb func(*meshShape[S])) {
	index := object.firstMeshShape
	for index >= 0 {
		shape := &s.meshShapes[index]
		nextIndex := shape.nextShape
		cb(shape)
		index = nextIndex
	}
}

func (s *Scene[O, S]) deleteObjectConvexShapes(object *sceneObject[O]) {
	index := object.firstConvexShape
	for index >= 0 {
		shape := &s.convexShapes[index]
		nextIndex := shape.nextShape
		s.detachConvexShape(index)
		index = nextIndex
	}
	object.firstConvexShape = -1
}

func (s *Scene[O, S]) deleteObjectMeshShapes(object *sceneObject[O]) {
	index := object.firstMeshShape
	for index >= 0 {
		shape := &s.meshShapes[index]
		nextIndex := shape.nextShape
		s.detachMeshShape(index)
		index = nextIndex
	}
	object.firstMeshShape = -1
}

func (s *Scene[O, S]) allocateConvexShape() int32 {
	if s.freeConvexShapeIndices.IsEmpty() {
		index := len(s.convexShapes)
		s.convexShapes = append(s.convexShapes, convexShape[S]{})
		return int32(index)
	} else {
		return s.freeConvexShapeIndices.Pop()
	}
}

func (s *Scene[O, S]) releaseConvexShape(index int32) {
	s.freeConvexShapeIndices.Push(index)
}

func (s *Scene[O, S]) allocateMeshShape() int32 {
	if s.freeMeshShapeIndices.IsEmpty() {
		index := len(s.meshShapes)
		s.meshShapes = append(s.meshShapes, meshShape[S]{})
		return int32(index)
	} else {
		return s.freeMeshShapeIndices.Pop()
	}
}

func (s *Scene[O, S]) releaseMeshShape(index int32) {
	s.freeMeshShapeIndices.Push(index)
}

func (s *Scene[O, S]) attachConvexShape(objectIndex int32, info ShapeInfo[S], representation convexShapeRepresentation) ShapeID {
	object := &s.objects[objectIndex]

	index := s.allocateConvexShape()
	ref := newShapeRef(index, false)

	representation.update(object.transform)

	tree := gog.Ternary(object.static, s.staticTree, s.dynamicTree)
	area := query3d.AreaFromSphere(representation.boundingSphere())

	s.convexShapes[index] = convexShape[S]{
		baseShape: baseShape[S]{
			objectIndex: objectIndex,
			nextShape:   object.firstConvexShape,

			spatialID: tree.Insert(area, ref),
			static:    object.static,

			rejectGroup: info.RejectGroup,
			sourceMask:  info.SourceMask.ValueOrDefault(0b1),
			targetMask:  info.TargetMask.ValueOrDefault(0b1),

			userData: info.UserData,
		},
		convexShapeRepresentation: representation,
	}
	object.firstConvexShape = index

	return ShapeID(ref)
}

func (s *Scene[O, S]) attachMeshShape(objectIndex int32, info ShapeInfo[S], representation meshShapeRepresentation) ShapeID {
	object := &s.objects[objectIndex]

	index := s.allocateMeshShape()
	ref := newShapeRef(index, true)

	representation.update(object.transform)

	tree := gog.Ternary(object.static, s.staticTree, s.dynamicTree)
	area := query3d.AreaFromSphere(representation.boundingSphere())

	s.meshShapes[index] = meshShape[S]{
		baseShape: baseShape[S]{
			objectIndex: objectIndex,
			nextShape:   object.firstMeshShape,

			spatialID: tree.Insert(area, ref),
			static:    object.static,

			rejectGroup: info.RejectGroup,
			sourceMask:  info.SourceMask.ValueOrDefault(0b1),
			targetMask:  info.TargetMask.ValueOrDefault(0b1),

			userData: info.UserData,
		},
		meshShapeRepresentation: representation,
	}
	object.firstMeshShape = index

	return ShapeID(ref)
}

func (s *Scene[O, S]) detachConvexShape(index int32) {
	shape := &s.convexShapes[index]
	object := &s.objects[shape.objectIndex]
	shape.objectIndex = -1

	tree := gog.Ternary(object.static, s.staticTree, s.dynamicTree)
	tree.Remove(shape.spatialID)
	shape.spatialID = query3d.InvalidTreeItemID

	if object.firstConvexShape == index {
		object.firstConvexShape = shape.nextShape
	} else {
		objShapeIndex := object.firstConvexShape
		for objShapeIndex >= 0 {
			objShape := &s.convexShapes[objShapeIndex]
			if objShape.nextShape == index {
				objShape.nextShape = shape.nextShape
				break
			}
			objShapeIndex = objShape.nextShape
		}
	}

	s.releaseConvexShape(index)
}

func (s *Scene[O, S]) detachMeshShape(index int32) {
	shape := &s.meshShapes[index]
	object := &s.objects[shape.objectIndex]
	shape.objectIndex = -1

	tree := gog.Ternary(object.static, s.staticTree, s.dynamicTree)
	tree.Remove(shape.spatialID)
	shape.spatialID = query3d.InvalidTreeItemID

	if object.firstMeshShape == index {
		object.firstMeshShape = shape.nextShape
	} else {
		objShapeIndex := object.firstMeshShape
		for objShapeIndex >= 0 {
			objShape := &s.meshShapes[objShapeIndex]
			if objShape.nextShape == index {
				objShape.nextShape = shape.nextShape
				break
			}
			objShapeIndex = objShape.nextShape
		}
	}

	s.releaseMeshShape(index)
}

func (s *Scene[O, S]) eachDynamicConvexShape(cb func(int32, *convexShape[S])) {
	for index := range s.convexShapes {
		shape := &s.convexShapes[index]
		if shape.static || (shape.spatialID == query3d.InvalidTreeItemID) {
			continue
		}
		cb(int32(index), shape)
	}
}

func (s *Scene[O, S]) eachDynamicMeshShape(cb func(int32, *meshShape[S])) {
	for index := range s.meshShapes {
		shape := &s.meshShapes[index]
		if shape.static || (shape.spatialID == query3d.InvalidTreeItemID) {
			continue
		}
		cb(int32(index), shape)
	}
}

func (s *Scene[O, S]) resolveSegmentConvex(segment shape3d.Segment, filter Filter, yield ContactCallback) {
	for _, index := range s.convexCandidates {
		shape := &s.convexShapes[index]
		if !shape.matchesFilter(filter) {
			continue
		}
		switch shape.kind {
		case convexKindSphere:
			sphere := shape.toSphere()
			isec3d.ResolveSegmentSphere(segment, sphere, func(contact shape3d.Contact) {
				yield(Contact{
					SourceObjectID: InvalidObjectID,
					SourceShapeID:  InvalidShapeID,
					TargetObjectID: ObjectID(shape.objectIndex),
					TargetShapeID:  ShapeID(newShapeRef(index, false)),
				})
			})
		case convexKindBox:
			box := shape.toBox()
			isec3d.ResolveSegmentBox(segment, box, func(contact shape3d.Contact) {
				yield(Contact{
					SourceObjectID: InvalidObjectID,
					SourceShapeID:  InvalidShapeID,
					TargetObjectID: ObjectID(shape.objectIndex),
					TargetShapeID:  ShapeID(newShapeRef(index, false)),
				})
			})
		}
	}
}

func (s *Scene[O, S]) resolveSegmentMesh(segment shape3d.Segment, filter Filter, yield ContactCallback) {
	for _, index := range s.meshCandidates {
		shape := &s.meshShapes[index]
		if !shape.matchesFilter(filter) {
			continue
		}
		// TODO: Maybe track just worse contact? Check old implementation!
		for _, triangle := range shape.wsTriangles {
			isec3d.ResolveSegmentTriangle(segment, triangle, func(contact shape3d.Contact) {
				yield(Contact{
					SourceObjectID: InvalidObjectID,
					SourceShapeID:  InvalidShapeID,
					TargetObjectID: ObjectID(shape.objectIndex),
					TargetShapeID:  ShapeID(newShapeRef(index, false)),
				})
			})
		}
	}
}

func (s *Scene[O, S]) resolvePairs(yield ContactCallback) {
	for _, candidate := range s.pairCandidates {
		if candidate.source.isMesh() {
			candidate = candidate.flipped()
		}
		if candidate.source.isMesh() {
			continue // we don't support mesh-mesh (different from hull-hull)
		}
		srcIndex := candidate.source.index()
		tgtIndex := candidate.target.index()
		if candidate.target.isMesh() {
			s.resolveConvexMesh(srcIndex, tgtIndex, yield)
		} else {
			s.resolveConvexConvex(srcIndex, tgtIndex, yield)
		}
	}
}

func (s *Scene[O, S]) resolveConvexConvex(srcIndex, tgtIndex int32, yield ContactCallback) {
	srcShape := &s.convexShapes[srcIndex]
	tgtShape := &s.convexShapes[tgtIndex]
	if !shapesCanIntersect(&srcShape.baseShape, &tgtShape.baseShape) {
		return
	}
	if !isec3d.CheckSphereSphere(srcShape.wsBSphere, tgtShape.wsBSphere) {
		return
	}

	srcGJKShape := srcShape.gjkShape()
	tgtGJKShape := tgtShape.gjkShape()
	if contact, ok := s.solver.Resolve(srcGJKShape, tgtGJKShape); ok {
		yield(Contact{
			SourceObjectID: ObjectID(srcShape.objectIndex),
			SourceShapeID:  ShapeID(newShapeRef(srcIndex, false)),
			TargetObjectID: ObjectID(tgtShape.objectIndex),
			TargetShapeID:  ShapeID(newShapeRef(tgtIndex, false)),
			Contact:        contact,
		})
	}
}

func (s *Scene[O, S]) resolveConvexMesh(srcIndex, tgtIndex int32, yield ContactCallback) {
	srcShape := &s.convexShapes[srcIndex]
	tgtShape := &s.meshShapes[tgtIndex]
	if !shapesCanIntersect(&srcShape.baseShape, &tgtShape.baseShape) {
		return
	}
	if !isec3d.CheckSphereSphere(srcShape.wsBSphere, tgtShape.wsBSphere) {
		return
	}

	var deepestContact shape3d.DeepestContact

	srcGJKShape := srcShape.gjkShape()
	for i := range tgtShape.gjkShapeCount() {
		// TODO: Run bounding sphere check first!
		tgtGJKShape := tgtShape.gjkShape(i)
		if contact, ok := s.solver.Resolve(srcGJKShape, tgtGJKShape); ok {
			deepestContact.AddContact(contact)
		}
	}

	if contact, ok := deepestContact.Contact(); ok {
		yield(Contact{
			SourceObjectID: ObjectID(srcShape.objectIndex),
			SourceShapeID:  ShapeID(newShapeRef(srcIndex, false)),
			TargetObjectID: ObjectID(tgtShape.objectIndex),
			TargetShapeID:  ShapeID(newShapeRef(tgtIndex, false)),
			Contact:        contact,
		})
	}
}
