package shape3d

import (
	"slices"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
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

// NewScene creates a new scene.
func NewScene[O, S any](settings SceneSettings) *Scene[O, S] {
	cubeOctreeSettings := CompactTreeSettings(settings)

	return &Scene[O, S]{
		freeObjectIndices: ds.NewStack[uint32](256), // ~ 1 KiB
		freeSphereIndices: ds.NewStack[uint32](256), // ~ 1 KiB
		freeBoxIndices:    ds.NewStack[uint32](256), // ~ 1 KiB
		freeMeshIndices:   ds.NewStack[uint32](256), // ~ 1 KiB

		objects: make([]sceneObject[O], 0, 128),
		spheres: make([]sceneSphereShape[S], 0, 128),
		boxes:   make([]sceneBoxShape[S], 0, 128),
		meshes:  make([]sceneMeshShape[S], 0, 128),

		staticTree:  NewCompactTree[shapeRef](cubeOctreeSettings),
		dynamicTree: NewCompactTree[shapeRef](cubeOctreeSettings),

		checks: make([]shapeRefPair, 0, 1024),
	}
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

	staticTree  *CompactTree[shapeRef]
	dynamicTree *CompactTree[shapeRef]

	tempShape   sceneShape[S]
	tempSegment Segment
	tempSphere  sphereSolver
	tempBox     boxSolver
	tempMesh    meshSolver

	checks []shapeRefPair
}

// CreateObject creates a new object.
func (s *Scene[O, S]) CreateObject(info ObjectInfo[O]) ObjectID {
	transform := Transform{
		Translation: info.Position.ValueOrDefault(dprec.ZeroVec3()),
		Rotation:    info.Rotation.ValueOrDefault(dprec.IdentityQuat()),
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
func (s *Scene[O, S]) GetObjectTransform(objID ObjectID) Transform {
	object := &s.objects[objID]
	return object.transform
}

// SetObjectTransform relocates the given object.
func (s *Scene[O, S]) SetObjectTransform(objID ObjectID, transform Transform) {
	object := &s.objects[objID]
	object.transform = transform
	s.eachObjectShape(object, shapeKindSphere, func(index uint32) {
		sphere := &s.spheres[index]
		sphere.update(transform)
		bs := sphere.boundingSphere()
		if sphere.static {
			s.staticTree.Update(sphere.spatialID, NewCompactCubeFromSphere(bs))
		} else {
			s.dynamicTree.Update(sphere.spatialID, NewCompactCubeFromSphere(bs))
		}
	})
	s.eachObjectShape(object, shapeKindBox, func(index uint32) {
		box := &s.boxes[index]
		box.update(transform)
		bs := box.boundingSphere()
		if box.static {
			s.staticTree.Update(box.spatialID, NewCompactCubeFromSphere(bs))
		} else {
			s.dynamicTree.Update(box.spatialID, NewCompactCubeFromSphere(bs))
		}
	})
	s.eachObjectShape(object, shapeKindMesh, func(index uint32) {
		mesh := &s.meshes[index]
		mesh.update(transform)
		bs := mesh.boundingSphere()
		if mesh.static {
			s.staticTree.Update(mesh.spatialID, NewCompactCubeFromSphere(bs))
		} else {
			s.dynamicTree.Update(mesh.spatialID, NewCompactCubeFromSphere(bs))
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
	var spatialID CompactTreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(NewCompactCubeFromSphere(bs), ref)
	} else {
		spatialID = s.dynamicTree.Insert(NewCompactCubeFromSphere(bs), ref)
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
	var spatialID CompactTreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(NewCompactCubeFromSphere(bs), ref)
	} else {
		spatialID = s.dynamicTree.Insert(NewCompactCubeFromSphere(bs), ref)
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
	var spatialID CompactTreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(NewCompactCubeFromSphere(bs), ref)
	} else {
		spatialID = s.dynamicTree.Insert(NewCompactCubeFromSphere(bs), ref)
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

// CollectIntersections yields intersections found in this scene.
func (s *Scene[O, S]) CollectIntersections(collection ObjectIntersectionCollection) {
	s.checks = s.checks[:0]

	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *sceneSphereShape[S]) {
		srcRef := newShapeRef(shapeKindSphere, srcIndex)
		queryAABB := NewCompactQueryAABBFromSphere(srcSphere.boundingSphere())
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
		queryAABB := NewCompactQueryAABBFromSphere(srcBox.boundingSphere())
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
		queryAABB := NewCompactQueryAABBFromSphere(srcMesh.boundingSphere())
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	})

	s.collectIntersections(collection)
}

// CheckSegmentIntersection returns the first intersection of the segment
// with the scene.
func (s *Scene[O, S]) CheckSegmentIntersection(segment Segment, mask uint32) (ObjectIntersection, bool) {
	var collection LargestObjectIntersection
	s.CollectSegmentIntersections(segment, mask, &collection)
	return collection.Intersection()
}

// CollectSegmentIntersections collects all intersections of the segment
// with objects in the scene.
func (s *Scene[O, S]) CollectSegmentIntersections(segment Segment, mask uint32, collection ObjectIntersectionCollection) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempSegment = segment
	srcRef := newTempShapeRef(shapeKindSegment)

	s.checks = s.checks[:0]
	querySegment := NewCompactQuerySegment(segment.A, segment.B)
	s.dynamicTree.QuerySegment(querySegment, func(tgtRef shapeRef) bool {
		s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
		return true
	})
	s.staticTree.QuerySegment(querySegment, func(tgtRef shapeRef) bool {
		s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
		return true
	})
	s.collectIntersections(collection)
}

// CheckSphereIntersection returns the first intersection of the sphere
// with the scene.
func (s *Scene[O, S]) CheckSphereIntersection(sphere Sphere, mask uint32) (ObjectIntersection, bool) {
	var collection LargestObjectIntersection
	s.CollectSphereIntersections(sphere, mask, &collection)
	return collection.Intersection()
}

// CheckSphereIntersection returns the first intersection of the sphere
// with the scene.
func (s *Scene[O, S]) CollectSphereIntersections(sphere Sphere, mask uint32, collection ObjectIntersectionCollection) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempSphere = newSphereSolver(sphere)
	srcRef := newTempShapeRef(shapeKindSphere)

	s.checks = s.checks[:0]
	queryAABB := NewCompactQueryAABBFromSphere(sphere)
	s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
		s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
		return true
	})
	s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
		s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
		return true
	})
	s.collectIntersections(collection)
}

// CheckBoxIntersection returns the first intersection of the box
// with the scene.
func (s *Scene[O, S]) CheckBoxIntersection(box Box, mask uint32) (ObjectIntersection, bool) {
	var collection LargestObjectIntersection
	s.CollectBoxIntersections(box, mask, &collection)
	return collection.Intersection()
}

// CollectBoxIntersections collects all intersections of the box
// with objects in the scene.
func (s *Scene[O, S]) CollectBoxIntersections(box Box, mask uint32, collection ObjectIntersectionCollection) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempBox = newBoxSolver(box)
	srcRef := newTempShapeRef(shapeKindBox)

	s.checks = s.checks[:0]
	queryAABB := NewCompactQueryAABBFromSphere(box.BoundingSphere())
	s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
		s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
		return true
	})
	s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
		s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
		return true
	})
	s.collectIntersections(collection)
}

// CheckMeshIntersection returns the first intersection of the mesh
// with the scene.
func (s *Scene[O, S]) CheckMeshIntersection(mesh Mesh, mask uint32) (ObjectIntersection, bool) {
	var collection LargestObjectIntersection
	s.CollectMeshIntersections(mesh, mask, &collection)
	return collection.Intersection()
}

// CollectMeshIntersections collects all intersections of the mesh
// with objects in the scene.
func (s *Scene[O, S]) CollectMeshIntersections(mesh Mesh, mask uint32, collection ObjectIntersectionCollection) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempMesh = newMeshSolver(mesh)
	srcRef := newTempShapeRef(shapeKindMesh)

	s.checks = s.checks[:0]
	queryAABB := NewCompactQueryAABBFromSphere(mesh.BoundingSphere())
	s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
		s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
		return true
	})
	s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
		s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
		return true
	})
	s.collectIntersections(collection)
}

// GC cleans up internal data and allows for memory reuse. This should be
// called once per frame.
func (s *Scene[O, S]) GC() {
	s.staticTree.GC()
	s.dynamicTree.GC()
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
		sphere.spatialID = InvalidCompactTreeItemID
		sphere.userData = gog.Zero[S]() // in case of pointer
		sphere.nextShape = invalidShapeRef
		sphere.sphereSolver = newSphereSolver(Sphere{})
		s.freeSphereIndices.Push(index)
	case shapeKindBox:
		box := &s.boxes[index]
		if box.static {
			s.staticTree.Remove(box.spatialID)
		} else {
			s.dynamicTree.Remove(box.spatialID)
		}
		box.spatialID = InvalidCompactTreeItemID
		box.userData = gog.Zero[S]() // in case of pointer
		box.nextShape = invalidShapeRef
		box.boxSolver = newBoxSolver(Box{})
		s.freeBoxIndices.Push(index)
	case shapeKindMesh:
		mesh := &s.meshes[index]
		if mesh.static {
			s.staticTree.Remove(mesh.spatialID)
		} else {
			s.dynamicTree.Remove(mesh.spatialID)
		}
		mesh.spatialID = InvalidCompactTreeItemID
		mesh.userData = gog.Zero[S]() // in case of pointer
		mesh.nextShape = invalidShapeRef
		mesh.meshSolver = newMeshSolver(Mesh{})
		s.freeMeshIndices.Push(index)
	default:
		panic("unknown shape reference")
	}
}

func (s *Scene[O, S]) eachDynamicSphere(cb func(uint32, *sceneSphereShape[S])) {
	for index := range uint32(len(s.spheres)) {
		shape := &s.spheres[index]
		if shape.static || (shape.spatialID == InvalidCompactTreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[O, S]) eachDynamicBox(cb func(uint32, *sceneBoxShape[S])) {
	for index := range uint32(len(s.boxes)) {
		shape := &s.boxes[index]
		if shape.static || (shape.spatialID == InvalidCompactTreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[O, S]) eachDynamicMesh(cb func(uint32, *sceneMeshShape[S])) {
	for index := range uint32(len(s.meshes)) {
		shape := &s.meshes[index]
		if shape.static || (shape.spatialID == InvalidCompactTreeItemID) {
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

func (s *Scene[O, S]) collectIntersections(collection ObjectIntersectionCollection) {
	slices.Sort(s.checks)

	index := 0
	for index < len(s.checks) {
		refPair := s.checks[index]

		srcRef := refPair.source()
		srcKind := srcRef.kind()

		tgtRef := refPair.target()
		tgtKind := tgtRef.kind()

		srcShape := s.getShape(srcRef)
		tgtShape := s.getShape(tgtRef)
		if !shapesCanIntersect(srcShape, tgtShape) {
			index++
			continue
		}

		switch {
		case srcKind == shapeKindSegment && tgtKind == shapeKindSphere:
			index = s.collectSegmentSphereIntersections(index, false, collection)
		case srcKind == shapeKindSphere && tgtKind == shapeKindSegment:
			index = s.collectSegmentSphereIntersections(index, true, collection)
		case srcKind == shapeKindSegment && tgtKind == shapeKindBox:
			index = s.collectSegmentBoxIntersections(index, false, collection)
		case srcKind == shapeKindBox && tgtKind == shapeKindSegment:
			index = s.collectSegmentBoxIntersections(index, true, collection)
		case srcKind == shapeKindSegment && tgtKind == shapeKindMesh:
			index = s.collectSegmentMeshIntersections(index, false, collection)
		case srcKind == shapeKindMesh && tgtKind == shapeKindSegment:
			index = s.collectSegmentMeshIntersections(index, true, collection)
		case srcKind == shapeKindSphere && tgtKind == shapeKindSphere:
			index = s.collectSphereSphereIntersections(index, collection)
		case srcKind == shapeKindSphere && tgtKind == shapeKindBox:
			index = s.collectSphereBoxIntersections(index, false, collection)
		case srcKind == shapeKindBox && tgtKind == shapeKindSphere:
			index = s.collectSphereBoxIntersections(index, true, collection)
		case srcKind == shapeKindSphere && tgtKind == shapeKindMesh:
			index = s.collectSphereMeshIntersections(index, false, collection)
		case srcKind == shapeKindMesh && tgtKind == shapeKindSphere:
			index = s.collectSphereMeshIntersections(index, true, collection)
		case srcKind == shapeKindBox && tgtKind == shapeKindMesh:
			index = s.collectBoxMeshIntersections(index, false, collection)
		case srcKind == shapeKindMesh && tgtKind == shapeKindBox:
			index = s.collectBoxMeshIntersections(index, true, collection)
		default:
			index++
		}
	}
}

func (s *Scene[O, S]) collectSegmentSphereIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getSphereSolver(tgtRef)

		if intersection, ok := s.checkSegmentSphereIntersection(srcSegment, tgtSolver); ok {
			intersection := ObjectIntersection{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Intersection:   intersection,
			}
			if flipped {
				collection.AddIntersection(intersection.Flipped())
			} else {
				collection.AddIntersection(intersection)
			}
		}
	})
}

func (s *Scene[O, S]) collectSegmentBoxIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getBoxSolver(tgtRef)

		if intersection, ok := s.checkSegmentBoxIntersection(srcSegment, tgtSolver); ok {
			intersection := ObjectIntersection{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Intersection:   intersection,
			}
			if flipped {
				collection.AddIntersection(intersection.Flipped())
			} else {
				collection.AddIntersection(intersection)
			}
		}
	})
}

func (s *Scene[O, S]) collectSegmentMeshIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getMeshSolver(tgtRef)

		if intersection, ok := s.checkSegmentMeshIntersection(srcSegment, tgtSolver); ok {
			intersection := ObjectIntersection{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Intersection:   intersection,
			}
			if flipped {
				collection.AddIntersection(intersection.Flipped())
			} else {
				collection.AddIntersection(intersection)
			}
		}
	})
}

func (s *Scene[O, S]) collectSphereSphereIntersections(index int, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, false, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getSphereSolver(srcRef)
		tgtShape, tgtSolver := s.getSphereSolver(tgtRef)

		if intersection, ok := s.checkSphereSphereIntersection(srcSolver, tgtSolver); ok {
			collection.AddIntersection(ObjectIntersection{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Intersection:   intersection,
			})
		}
	})
}

func (s *Scene[O, S]) collectSphereBoxIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getSphereSolver(srcRef)
		tgtShape, tgtSolver := s.getBoxSolver(tgtRef)

		if intersection, ok := s.checkSphereBoxIntersection(srcSolver, tgtSolver); ok {
			intersection := ObjectIntersection{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Intersection:   intersection,
			}
			if flipped {
				collection.AddIntersection(intersection.Flipped())
			} else {
				collection.AddIntersection(intersection)
			}
		}
	})
}

func (s *Scene[O, S]) collectSphereMeshIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getSphereSolver(srcRef)
		tgtShape, tgtSolver := s.getMeshSolver(tgtRef)

		if intersection, ok := s.checkSphereMeshIntersection(srcSolver, tgtSolver); ok {
			intersection := ObjectIntersection{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Intersection:   intersection,
			}
			if flipped {
				collection.AddIntersection(intersection.Flipped())
			} else {
				collection.AddIntersection(intersection)
			}
		}
	})
}

func (s *Scene[O, S]) collectBoxMeshIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getBoxSolver(srcRef)
		tgtShape, tgtSolver := s.getMeshSolver(tgtRef)

		if intersection, ok := s.checkBoxMeshIntersection(srcSolver, tgtSolver); ok {
			intersection := ObjectIntersection{
				SourceObjectID: wrapObjectID(srcShape),
				SourceShapeID:  wrapShapeID[S](srcRef),
				TargetObjectID: wrapObjectID(tgtShape),
				TargetShapeID:  wrapShapeID[S](tgtRef),
				Intersection:   intersection,
			}
			if flipped {
				collection.AddIntersection(intersection.Flipped())
			} else {
				collection.AddIntersection(intersection)
			}
		}
	})
}

func (s *Scene[O, S]) getSegmentSolver(ref shapeRef) (*sceneShape[S], Segment) {
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
		if flipped {
			cb(refPair.flipped())
		} else {
			cb(refPair)
		}
		index++
	}
	return index
}

func (s *Scene[O, S]) checkSegmentSphereIntersection(source Segment, target *sphereSolver) (Intersection, bool) {
	return CheckSegmentSphereIntersection(source, target.wsSphere)
}

func (s *Scene[O, S]) checkSegmentBoxIntersection(source Segment, target *boxSolver) (Intersection, bool) {
	if !IsSegmentSphereOverlap(source, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	return CheckSegmentBoxIntersection(source, target.wsBox)
}

func (s *Scene[O, S]) checkSegmentMeshIntersection(source Segment, target *meshSolver) (Intersection, bool) {
	if !IsSegmentSphereOverlap(source, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	return CheckSegmentMeshIntersection(source, target.wsMesh)
}

func (s *Scene[O, S]) checkSphereSphereIntersection(source, target *sphereSolver) (Intersection, bool) {
	return CheckSphereSphereIntersection(source.wsSphere, target.wsSphere)
}

func (s *Scene[O, S]) checkSphereBoxIntersection(source *sphereSolver, target *boxSolver) (Intersection, bool) {
	if !IsSphereSphereIntersection(source.wsSphere, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	return CheckSphereBoxIntersection(source.wsSphere, target.wsBox)
}

func (s *Scene[O, S]) checkSphereMeshIntersection(source *sphereSolver, target *meshSolver) (Intersection, bool) {
	if !IsSphereSphereIntersection(source.wsSphere, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	return CheckSphereMeshIntersection(source.wsSphere, target.wsMesh)
}

func (s *Scene[O, S]) checkBoxMeshIntersection(source *boxSolver, target *meshSolver) (Intersection, bool) {
	if !IsSphereSphereIntersection(source.wsBoundingSphere, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	return CheckBoxMeshIntersection(source.wsBox, target.wsMesh)
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
