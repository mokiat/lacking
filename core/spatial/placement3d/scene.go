package placement3d

import (
	"iter"
	"slices"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
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
type Scene[T, S any] struct {
	freeObjectIndices *ds.Stack[uint32]
	freeSphereIndices *ds.Stack[uint32]
	freeBoxIndices    *ds.Stack[uint32]
	freeMeshIndices   *ds.Stack[uint32]

	objects []sceneObject[T]
	spheres []sceneSphereShape[S]
	boxes   []sceneBoxShape[S]
	meshes  []sceneMeshShape[S]

	staticTree  *query3d.Octree[shapeRef]
	dynamicTree *query3d.Octree[shapeRef]

	tempShape   sceneShape[S]
	tempSegment shape3d.Segment
	tempSphere  sphereSolver
	tempBox     boxSolver
	tempMesh    meshSolver

	checks []shapeRefPair
}

// NewScene creates a new scene.
func NewScene[O, S any](settings SceneSettings) *Scene[O, S] {
	treeSettings := query3d.OctreeSettings(settings)

	return &Scene[O, S]{
		freeObjectIndices: ds.EmptyStack[uint32](),
		freeSphereIndices: ds.EmptyStack[uint32](),
		freeBoxIndices:    ds.EmptyStack[uint32](),
		freeMeshIndices:   ds.EmptyStack[uint32](),

		objects: make([]sceneObject[O], 0),
		spheres: make([]sceneSphereShape[S], 0),
		boxes:   make([]sceneBoxShape[S], 0),
		meshes:  make([]sceneMeshShape[S], 0),

		staticTree:  query3d.NewOctree[shapeRef](treeSettings),
		dynamicTree: query3d.NewOctree[shapeRef](treeSettings),

		checks: make([]shapeRefPair, 0),
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

	flags := objectFlagsNone
	if info.Static {
		flags |= objectFlagsStatic
	}

	if s.freeObjectIndices.IsEmpty() {
		index := len(s.objects)
		s.objects = append(s.objects, sceneObject[O]{
			transform:  transform,
			firstShape: invalidShapeRef,
			flags:      flags,
			userData:   info.UserData,
		})
		return ObjectID(index)
	} else {
		index := s.freeObjectIndices.Pop()
		s.objects[index] = sceneObject[O]{
			transform:  transform,
			firstShape: invalidShapeRef,
			flags:      flags,
			userData:   info.UserData,
		}
		return ObjectID(index)
	}
}

// DeleteObject deletes an object.
func (s *Scene[O, S]) DeleteObject(objID ObjectID) {
	object := &s.objects[objID]
	s.deleteObjectShapes(object)
	object.userData = gog.Zero[O]() // in case of pointer
	s.freeObjectIndices.Push(uint32(objID))
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
	s.eachObjectShape(object, shapeKindSphere, func(index uint32) {
		sphere := &s.spheres[index]
		sphere.update(transform)
		bs := sphere.boundingSphere()
		if sphere.static {
			s.staticTree.Update(sphere.spatialID, query3d.AreaFromSphere(bs))
		} else {
			s.dynamicTree.Update(sphere.spatialID, query3d.AreaFromSphere(bs))
		}
	})
	s.eachObjectShape(object, shapeKindBox, func(index uint32) {
		box := &s.boxes[index]
		box.update(transform)
		bs := box.boundingSphere()
		if box.static {
			s.staticTree.Update(box.spatialID, query3d.AreaFromSphere(bs))
		} else {
			s.dynamicTree.Update(box.spatialID, query3d.AreaFromSphere(bs))
		}
	})
	s.eachObjectShape(object, shapeKindMesh, func(index uint32) {
		mesh := &s.meshes[index]
		mesh.update(transform)
		bs := mesh.boundingSphere()
		if mesh.static {
			s.staticTree.Update(mesh.spatialID, query3d.AreaFromSphere(bs))
		} else {
			s.dynamicTree.Update(mesh.spatialID, query3d.AreaFromSphere(bs))
		}
	})
}

// AttachSphere creates a sphere shape and attaches it to the object to be
// used for intersection tests.
func (s *Scene[O, S]) AttachSphere(objID ObjectID, info SphereInfo[S]) ShapeID {
	var index uint32
	if s.freeSphereIndices.IsEmpty() {
		index = uint32(len(s.spheres))
		s.spheres = append(s.spheres, sceneSphereShape[S]{})
	} else {
		index = s.freeSphereIndices.Pop()
	}
	ref := newShapeRef(shapeKindSphere, index)

	object := &s.objects[objID]

	solver := newSphereSolver(info.Sphere)
	solver.update(object.transform)

	bs := solver.boundingSphere()
	var spatialID query3d.TreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(query3d.AreaFromSphere(bs), ref)
	} else {
		spatialID = s.dynamicTree.Insert(query3d.AreaFromSphere(bs), ref)
	}

	s.spheres[index] = sceneSphereShape[S]{
		sceneShape: sceneShape[S]{
			objectIndex: uint32(objID),
			nextShape:   object.firstShape,
			spatialID:   spatialID,
			static:      object.isStatic(),
			rejectGroup: info.RejectGroup,
			sourceMask:  info.SourceMask.ValueOrDefault(0b1),
			targetMask:  info.TargetMask.ValueOrDefault(0b1),
			userData:    info.UserData,
		},
		sphereSolver: solver,
	}
	object.firstShape = ref

	return ShapeID(ref)
}

// AttachBox creates a box shape and attaches it to the object to be used for
// intersection tests.
func (s *Scene[O, S]) AttachBox(objID ObjectID, info BoxInfo[S]) ShapeID {
	var index uint32
	if s.freeBoxIndices.IsEmpty() {
		index = uint32(len(s.boxes))
		s.boxes = append(s.boxes, sceneBoxShape[S]{})
	} else {
		index = s.freeBoxIndices.Pop()
	}
	ref := newShapeRef(shapeKindBox, index)

	object := &s.objects[objID]

	solver := newBoxSolver(info.Box)
	solver.update(object.transform)

	bs := solver.boundingSphere()
	var spatialID query3d.TreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(query3d.AreaFromSphere(bs), ref)
	} else {
		spatialID = s.dynamicTree.Insert(query3d.AreaFromSphere(bs), ref)
	}

	s.boxes[index] = sceneBoxShape[S]{
		sceneShape: sceneShape[S]{
			objectIndex: uint32(objID),
			nextShape:   object.firstShape,
			spatialID:   spatialID,
			static:      object.isStatic(),
			rejectGroup: info.RejectGroup,
			sourceMask:  info.SourceMask.ValueOrDefault(0b1),
			targetMask:  info.TargetMask.ValueOrDefault(0b1),
			userData:    info.UserData,
		},
		boxSolver: solver,
	}
	object.firstShape = ref

	return ShapeID(ref)
}

// AttachMesh creates a mesh shape and attaches it to the object to be used for
// intersection tests.
func (s *Scene[O, S]) AttachMesh(objID ObjectID, info MeshInfo[S]) ShapeID {
	var index uint32
	if s.freeMeshIndices.IsEmpty() {
		index = uint32(len(s.meshes))
		s.meshes = append(s.meshes, sceneMeshShape[S]{})
	} else {
		index = s.freeMeshIndices.Pop()
	}
	ref := newShapeRef(shapeKindMesh, index)

	object := &s.objects[objID]

	solver := newMeshSolver(info.Mesh)
	solver.update(object.transform)

	bs := solver.boundingSphere()
	var spatialID query3d.TreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(query3d.AreaFromSphere(bs), ref)
	} else {
		spatialID = s.dynamicTree.Insert(query3d.AreaFromSphere(bs), ref)
	}

	s.meshes[index] = sceneMeshShape[S]{
		sceneShape: sceneShape[S]{
			objectIndex: uint32(objID),
			nextShape:   object.firstShape,
			spatialID:   spatialID,
			static:      object.isStatic(),
			rejectGroup: info.RejectGroup,
			sourceMask:  info.SourceMask.ValueOrDefault(0b1),
			targetMask:  info.TargetMask.ValueOrDefault(0b1),
			userData:    info.UserData,
		},
		meshSolver: solver,
	}
	object.firstShape = ref

	return ShapeID(ref)
}

// DeleteShape deletes a shape from an object. The object is not
// deleted and continues to exist in the scene.
func (s *Scene[O, S]) DeleteShape(shapeID ShapeID) {
	ref := shapeRef(shapeID)
	shape := s.getShape(ref)

	object := &s.objects[shape.objectIndex]
	if object.firstShape == ref {
		object.firstShape = shape.nextShape
	} else {
		objShapeRef := object.firstShape
		for objShapeRef != invalidShapeRef {
			objShape := s.getShape(objShapeRef)
			if objShape.nextShape == ref {
				objShape.nextShape = shape.nextShape
				break
			}
			objShapeRef = objShape.nextShape
		}
	}

	s.freeShape(ref)
}

// GetShapeUserData returns the user data associated with the given shape.
func (s *Scene[O, S]) GetShapeUserData(shapeID ShapeID) S {
	shape := s.getShape(shapeRef(shapeID))
	return shape.userData
}

// SetShapeUserData assigns the specified user data to the shape.
func (s *Scene[O, S]) SetShapeUserData(shapeID ShapeID, userData S) {
	shape := s.getShape(shapeRef(shapeID))
	shape.userData = userData
}

// EachSphere iterates over all sphere shapes in the scene that match the
// filter and yields them to the provided callback.
func (s *Scene[O, S]) EachSphere(filter Filter, yield func(shape3d.Sphere) bool) {
	for index := range uint32(len(s.spheres)) {
		shape := &s.spheres[index]
		if shape.spatialID == query3d.InvalidTreeItemID {
			continue
		}
		if !shape.matchesFilter(filter) {
			continue
		}
		if !yield(shape.sphereSolver.wsSphere) {
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
	for index := range uint32(len(s.boxes)) {
		shape := &s.boxes[index]
		if shape.spatialID == query3d.InvalidTreeItemID {
			continue
		}
		if !shape.matchesFilter(filter) {
			continue
		}
		if !yield(shape.boxSolver.wsBox) {
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
	for index := range uint32(len(s.meshes)) {
		shape := &s.meshes[index]
		if shape.spatialID == query3d.InvalidTreeItemID {
			continue
		}
		if !shape.matchesFilter(filter) {
			continue
		}
		if !yield(shape.meshSolver.wsMesh) {
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

// CollectIntersections yields intersections found in this scene.
func (s *Scene[O, S]) CollectIntersections(yield ContactCallback) {
	s.checks = s.checks[:0]

	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *sceneSphereShape[S]) {
		srcRef := newShapeRef(shapeKindSphere, srcIndex)
		queryAABB := query3d.AABBFromSphere(srcSphere.boundingSphere())
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	})

	s.eachDynamicBox(func(srcIndex uint32, srcBox *sceneBoxShape[S]) {
		srcRef := newShapeRef(shapeKindBox, srcIndex)
		queryAABB := query3d.AABBFromSphere(srcBox.boundingSphere())
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	})

	s.eachDynamicMesh(func(srcIndex uint32, srcMesh *sceneMeshShape[S]) {
		srcRef := newShapeRef(shapeKindMesh, srcIndex)
		queryAABB := query3d.AABBFromSphere(srcMesh.boundingSphere())
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	})

	s.collectIntersections(yield)
}

// CheckSegmentIntersection returns the first intersection of the segment
// with the scene.
func (s *Scene[O, S]) CheckSegmentIntersection(segment shape3d.Segment, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectSegmentIntersections(segment, filter, collection.AddContact)
	return collection.Contact()
}

// CollectSegmentIntersections collects all intersections of the segment
// with objects in the scene.
func (s *Scene[O, S]) CollectSegmentIntersections(segment shape3d.Segment, filter Filter, yield ContactCallback) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempSegment = segment
	srcRef := newTempShapeRef(shapeKindSegment)

	s.checks = s.checks[:0]
	querySegment := query3d.NewSegment(segment.A, segment.B)
	if !filter.SkipDynamic {
		s.dynamicTree.QuerySegment(querySegment, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	}
	if !filter.SkipStatic {
		s.staticTree.QuerySegment(querySegment, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	}
	s.collectIntersections(yield)
}

// CheckSphereIntersection returns the first intersection of the sphere
// with the scene.
func (s *Scene[O, S]) CheckSphereIntersection(sphere shape3d.Sphere, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectSphereIntersections(sphere, filter, collection.AddContact)
	return collection.Contact()
}

// CheckSphereIntersection returns the first intersection of the sphere
// with the scene.
func (s *Scene[O, S]) CollectSphereIntersections(sphere shape3d.Sphere, filter Filter, yield ContactCallback) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempSphere = newSphereSolver(sphere)
	srcRef := newTempShapeRef(shapeKindSphere)

	s.checks = s.checks[:0]
	queryAABB := query3d.AABBFromSphere(sphere)
	if !filter.SkipDynamic {
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	}
	if !filter.SkipStatic {
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	}
	s.collectIntersections(yield)
}

// CheckBoxIntersection returns the first intersection of the box
// with the scene.
func (s *Scene[O, S]) CheckBoxIntersection(box shape3d.Box, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectBoxIntersections(box, filter, collection.AddContact)
	return collection.Contact()
}

// CollectBoxIntersections collects all intersections of the box
// with objects in the scene.
func (s *Scene[O, S]) CollectBoxIntersections(box shape3d.Box, filter Filter, yield ContactCallback) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempBox = newBoxSolver(box)
	srcRef := newTempShapeRef(shapeKindBox)

	s.checks = s.checks[:0]
	queryAABB := query3d.AABBFromSphere(box.BoundingSphere())
	if !filter.SkipDynamic {
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	}
	if !filter.SkipStatic {
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	}
	s.collectIntersections(yield)
}

// CheckMeshIntersection returns the first intersection of the mesh
// with the scene.
func (s *Scene[O, S]) CheckMeshIntersection(mesh shape3d.Mesh, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectMeshIntersections(mesh, filter, collection.AddContact)
	return collection.Contact()
}

// CollectMeshIntersections collects all intersections of the mesh
// with objects in the scene.
func (s *Scene[O, S]) CollectMeshIntersections(mesh shape3d.Mesh, filter Filter, yield ContactCallback) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempMesh = newMeshSolver(mesh)
	srcRef := newTempShapeRef(shapeKindMesh)

	s.checks = s.checks[:0]
	queryAABB := query3d.AABBFromSphere(mesh.BoundingSphere())
	if !filter.SkipDynamic {
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	}
	if !filter.SkipStatic {
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	}
	s.collectIntersections(yield)
}

func (s *Scene[O, S]) getShape(ref shapeRef) *sceneShape[S] {
	if ref.isTemporary() {
		return &s.tempShape
	}
	switch ref.kind() {
	case shapeKindSphere:
		sphere := &s.spheres[ref.index()]
		return &sphere.sceneShape
	case shapeKindBox:
		box := &s.boxes[ref.index()]
		return &box.sceneShape
	case shapeKindMesh:
		mesh := &s.meshes[ref.index()]
		return &mesh.sceneShape
	default:
		panic("unknown shape reference")
	}
}

func (s *Scene[O, S]) freeShape(ref shapeRef) {
	index := ref.index()
	switch ref.kind() {
	case shapeKindSphere:
		sphere := &s.spheres[index]
		if sphere.static {
			s.staticTree.Remove(sphere.spatialID)
		} else {
			s.dynamicTree.Remove(sphere.spatialID)
		}
		sphere.spatialID = query3d.InvalidTreeItemID
		sphere.userData = gog.Zero[S]() // in case of pointer
		sphere.nextShape = invalidShapeRef
		sphere.sphereSolver = newSphereSolver(shape3d.Sphere{})
		s.freeSphereIndices.Push(index)
	case shapeKindBox:
		box := &s.boxes[index]
		if box.static {
			s.staticTree.Remove(box.spatialID)
		} else {
			s.dynamicTree.Remove(box.spatialID)
		}
		box.spatialID = query3d.InvalidTreeItemID
		box.userData = gog.Zero[S]() // in case of pointer
		box.nextShape = invalidShapeRef
		box.boxSolver = newBoxSolver(shape3d.Box{})
		s.freeBoxIndices.Push(index)
	case shapeKindMesh:
		mesh := &s.meshes[index]
		if mesh.static {
			s.staticTree.Remove(mesh.spatialID)
		} else {
			s.dynamicTree.Remove(mesh.spatialID)
		}
		mesh.spatialID = query3d.InvalidTreeItemID
		mesh.userData = gog.Zero[S]() // in case of pointer
		mesh.nextShape = invalidShapeRef
		mesh.meshSolver = newMeshSolver(shape3d.Mesh{})
		s.freeMeshIndices.Push(index)
	default:
		panic("unknown shape reference")
	}
}

func (s *Scene[O, S]) eachDynamicSphere(cb func(uint32, *sceneSphereShape[S])) {
	for index := range uint32(len(s.spheres)) {
		shape := &s.spheres[index]
		if shape.static || (shape.spatialID == query3d.InvalidTreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[O, S]) eachDynamicBox(cb func(uint32, *sceneBoxShape[S])) {
	for index := range uint32(len(s.boxes)) {
		shape := &s.boxes[index]
		if shape.static || (shape.spatialID == query3d.InvalidTreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[O, S]) eachDynamicMesh(cb func(uint32, *sceneMeshShape[S])) {
	for index := range uint32(len(s.meshes)) {
		shape := &s.meshes[index]
		if shape.static || (shape.spatialID == query3d.InvalidTreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[O, S]) deleteObjectShapes(object *sceneObject[O]) {
	ref := object.firstShape
	for ref != invalidShapeRef {
		shape := s.getShape(ref)
		nextRef := shape.nextShape
		s.freeShape(ref)
		ref = nextRef
	}
	object.firstShape = invalidShapeRef
}

func (s *Scene[O, S]) eachObjectShape(object *sceneObject[O], kind shapeKind, cb func(uint32)) {
	ref := object.firstShape
	for ref != invalidShapeRef {
		shape := s.getShape(ref)
		nextRef := shape.nextShape
		if ref.kind() == kind {
			cb(ref.index())
		}
		ref = nextRef
	}
}

func (s *Scene[O, S]) collectIntersections(yield ContactCallback) {
	slices.Sort(s.checks)

	index := 0
	for index < len(s.checks) {
		refPair := s.checks[index]

		srcRef := refPair.source()
		srcKind := srcRef.kind()

		tgtRef := refPair.target()
		tgtKind := tgtRef.kind()

		switch {
		case srcKind == shapeKindSegment && tgtKind == shapeKindSphere:
			index = s.collectSegmentSphereIntersections(index, false, yield)
		case srcKind == shapeKindSphere && tgtKind == shapeKindSegment:
			index = s.collectSegmentSphereIntersections(index, true, yield)
		case srcKind == shapeKindSegment && tgtKind == shapeKindBox:
			index = s.collectSegmentBoxIntersections(index, false, yield)
		case srcKind == shapeKindBox && tgtKind == shapeKindSegment:
			index = s.collectSegmentBoxIntersections(index, true, yield)
		case srcKind == shapeKindSegment && tgtKind == shapeKindMesh:
			index = s.collectSegmentMeshIntersections(index, false, yield)
		case srcKind == shapeKindMesh && tgtKind == shapeKindSegment:
			index = s.collectSegmentMeshIntersections(index, true, yield)
		case srcKind == shapeKindSphere && tgtKind == shapeKindSphere:
			index = s.collectSphereSphereIntersections(index, yield)
		case srcKind == shapeKindSphere && tgtKind == shapeKindBox:
			index = s.collectSphereBoxIntersections(index, false, yield)
		case srcKind == shapeKindBox && tgtKind == shapeKindSphere:
			index = s.collectSphereBoxIntersections(index, true, yield)
		case srcKind == shapeKindSphere && tgtKind == shapeKindMesh:
			index = s.collectSphereMeshIntersections(index, false, yield)
		case srcKind == shapeKindMesh && tgtKind == shapeKindSphere:
			index = s.collectSphereMeshIntersections(index, true, yield)
		case srcKind == shapeKindBox && tgtKind == shapeKindMesh:
			index = s.collectBoxMeshIntersections(index, false, yield)
		case srcKind == shapeKindMesh && tgtKind == shapeKindBox:
			index = s.collectBoxMeshIntersections(index, true, yield)
		default:
			index++
		}
	}
}

func (s *Scene[O, S]) collectSegmentSphereIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getSphereSolver(tgtRef)

		s.checkSegmentSphereIntersection(srcSegment, tgtSolver, func(contact shape3d.Contact) {
			objContact := Contact{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Contact:        contact,
			}
			if flipped {
				yield(objContact.Flipped())
			} else {
				yield(objContact)
			}
		})
	})
}

func (s *Scene[O, S]) collectSegmentBoxIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getBoxSolver(tgtRef)

		s.checkSegmentBoxIntersection(srcSegment, tgtSolver, func(contact shape3d.Contact) {
			objContact := Contact{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Contact:        contact,
			}
			if flipped {
				yield(objContact.Flipped())
			} else {
				yield(objContact)
			}
		})
	})
}

func (s *Scene[O, S]) collectSegmentMeshIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getMeshSolver(tgtRef)

		s.checkSegmentMeshIntersection(srcSegment, tgtSolver, func(contact shape3d.Contact) {
			objContact := Contact{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Contact:        contact,
			}
			if flipped {
				yield(objContact.Flipped())
			} else {
				yield(objContact)
			}
		})
	})
}

func (s *Scene[O, S]) collectSphereSphereIntersections(index int, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, false, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getSphereSolver(srcRef)
		tgtShape, tgtSolver := s.getSphereSolver(tgtRef)

		s.checkSphereSphereIntersection(srcSolver, tgtSolver, func(contact shape3d.Contact) {
			yield(Contact{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Contact:        contact,
			})
		})
	})
}

func (s *Scene[O, S]) collectSphereBoxIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getSphereSolver(srcRef)
		tgtShape, tgtSolver := s.getBoxSolver(tgtRef)

		s.checkSphereBoxIntersection(srcSolver, tgtSolver, func(contact shape3d.Contact) {
			objContact := Contact{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Contact:        contact,
			}
			if flipped {
				yield(objContact.Flipped())
			} else {
				yield(objContact)
			}
		})
	})
}

func (s *Scene[O, S]) collectSphereMeshIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getSphereSolver(srcRef)
		tgtShape, tgtSolver := s.getMeshSolver(tgtRef)

		s.checkSphereMeshIntersection(srcSolver, tgtSolver, func(contact shape3d.Contact) {
			objContact := Contact{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Contact:        contact,
			}
			if flipped {
				yield(objContact.Flipped())
			} else {
				yield(objContact)
			}
		})
	})
}

func (s *Scene[O, S]) collectBoxMeshIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getBoxSolver(srcRef)
		tgtShape, tgtSolver := s.getMeshSolver(tgtRef)

		s.checkBoxMeshIntersection(srcSolver, tgtSolver, func(contact shape3d.Contact) {
			objContact := Contact{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Contact:        contact,
			}
			if flipped {
				yield(objContact.Flipped())
			} else {
				yield(objContact)
			}
		})
	})
}

func (s *Scene[O, S]) getSegmentSolver(ref shapeRef) (*sceneShape[S], shape3d.Segment) {
	if !ref.isTemporary() {
		panic("expected temporary shape reference")
	}
	return &s.tempShape, s.tempSegment
}

func (s *Scene[O, S]) getSphereSolver(ref shapeRef) (*sceneShape[S], *sphereSolver) {
	if ref.isTemporary() {
		return &s.tempShape, &s.tempSphere
	}
	sphere := &s.spheres[ref.index()]
	return &sphere.sceneShape, &sphere.sphereSolver
}

func (s *Scene[O, S]) getBoxSolver(ref shapeRef) (*sceneShape[S], *boxSolver) {
	if ref.isTemporary() {
		return &s.tempShape, &s.tempBox
	}
	box := &s.boxes[ref.index()]
	return &box.sceneShape, &box.boxSolver
}

func (s *Scene[O, S]) getMeshSolver(ref shapeRef) (*sceneShape[S], *meshSolver) {
	if ref.isTemporary() {
		return &s.tempShape, &s.tempMesh
	}
	mesh := &s.meshes[ref.index()]
	return &mesh.sceneShape, &mesh.meshSolver
}

func (s *Scene[O, S]) consumeSameKindRefPairs(index int, flipped bool, cb func(refPair shapeRefPair)) int {
	refPair := s.checks[index]
	previousSrcKind := refPair.source().kind()
	previousTgtKind := refPair.target().kind()
	for index < len(s.checks) {
		refPair := s.checks[index]
		srcKind := refPair.source().kind()
		tgtKind := refPair.target().kind()
		if srcKind != previousSrcKind || tgtKind != previousTgtKind {
			break
		}
		srcShape := s.getShape(refPair.source())
		tgtShape := s.getShape(refPair.target())
		if !shapesCanIntersect(srcShape, tgtShape) {
			index++
			continue
		}
		if flipped {
			cb(refPair.flipped())
		} else {
			cb(refPair)
		}
		index++
	}
	return index
}

func (s *Scene[O, S]) checkSegmentSphereIntersection(source shape3d.Segment, target *sphereSolver, yield shape3d.ContactCallback) {
	isec3d.ResolveSegmentSphere(source, target.wsSphere, yield)
}

func (s *Scene[O, S]) checkSegmentBoxIntersection(source shape3d.Segment, target *boxSolver, yield shape3d.ContactCallback) {
	if !isec3d.CheckSegmentSphereOverlap(source, target.wsBoundingSphere) {
		return
	}
	isec3d.ResolveSegmentBox(source, target.wsBox, yield)
}

func (s *Scene[O, S]) checkSegmentMeshIntersection(source shape3d.Segment, target *meshSolver, yield shape3d.ContactCallback) {
	if !isec3d.CheckSegmentSphereOverlap(source, target.wsBoundingSphere) {
		return
	}
	isec3d.ResolveSegmentMesh(source, target.wsMesh, yield)
}

func (s *Scene[O, S]) checkSphereSphereIntersection(source, target *sphereSolver, yield shape3d.ContactCallback) {
	isec3d.ResolveSphereSphere(source.wsSphere, target.wsSphere, yield)
}

func (s *Scene[O, S]) checkSphereBoxIntersection(source *sphereSolver, target *boxSolver, yield shape3d.ContactCallback) {
	if !isec3d.CheckSphereSphere(source.wsSphere, target.wsBoundingSphere) {
		return
	}
	isec3d.ResolveSphereBox(source.wsSphere, target.wsBox, yield)
}

func (s *Scene[O, S]) checkSphereMeshIntersection(source *sphereSolver, target *meshSolver, yield shape3d.ContactCallback) {
	if !isec3d.CheckSphereSphere(source.wsSphere, target.wsBoundingSphere) {
		return
	}
	isec3d.ResolveSphereMesh(source.wsSphere, target.wsMesh, yield)
}

func (s *Scene[O, S]) checkBoxMeshIntersection(source *boxSolver, target *meshSolver, yield shape3d.ContactCallback) {
	if !isec3d.CheckSphereSphere(source.wsBoundingSphere, target.wsBoundingSphere) {
		return
	}
	isec3d.ResolveBoxMesh(source.wsBox, target.wsMesh, yield)
}

func wrapObjectID[S any](shape *sceneShape[S]) ObjectID {
	return ObjectID(shape.objectIndex)
}

func wrapShapeID[S any](ref shapeRef) ShapeID {
	if ref.isTemporary() {
		return InvalidShapeID
	}
	return ShapeID(ref)
}
