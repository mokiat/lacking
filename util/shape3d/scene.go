package shape3d

import (
	"slices"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// SceneInfo contains information needed to create an optimal scene.
type SceneInfo struct {

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
func NewScene[O, S any](info SceneInfo) *Scene[O, S] {
	cubeOctreeSettings := CompactTreeSettings(info)

	return &Scene[O, S]{
		freeObjectIndices: ds.NewStack[uint32](256), // ~ 1 KiB
		freeSphereIndices: ds.NewStack[uint32](256), // ~ 1 KiB
		freeBoxIndices:    ds.NewStack[uint32](256), // ~ 1 KiB
		freeMeshIndices:   ds.NewStack[uint32](256), // ~ 1 KiB

		objects: make([]sceneObject[O], 0, 128),
		spheres: make([]sceneSphereShape[S], 0, 128),
		boxes:   make([]sceneBoxShape[S], 0, 128),
		meshes:  make([]sceneMeshShape[S], 0, 128),

		sphereTree: NewCompactTree[uint32](cubeOctreeSettings),
		boxTree:    NewCompactTree[uint32](cubeOctreeSettings),
		meshTree:   NewCompactTree[uint32](cubeOctreeSettings),

		checks: make([]indexPair, 0, 1024),
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

	sphereTree *CompactTree[uint32]
	boxTree    *CompactTree[uint32]
	meshTree   *CompactTree[uint32]

	checks []indexPair
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

// DeleteObject deletes an object. If the object was inserted into the
// scene, it is first removed.
func (s *Scene[O, S]) DeleteObject(objID ObjectID) {
	object := &s.objects[objID]
	s.deleteObjectShapes(object)
	object.firstShape = invalidShapeRef
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
		s.sphereTree.Update(sphere.spatialID, NewCompactCubeFromSphere(bs.Position, bs.Radius))
	})
	s.eachObjectShape(object, shapeKindBox, func(index uint32) {
		box := &s.boxes[index]
		box.update(transform)
		bs := box.boundingSphere()
		s.boxTree.Update(box.spatialID, NewCompactCubeFromSphere(bs.Position, bs.Radius))
	})
	s.eachObjectShape(object, shapeKindMesh, func(index uint32) {
		mesh := &s.meshes[index]
		mesh.update(transform)
		bs := mesh.boundingSphere()
		s.meshTree.Update(mesh.spatialID, NewCompactCubeFromSphere(bs.Position, bs.Radius))
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
	spatialID := s.sphereTree.Insert(NewCompactCubeFromSphere(bs.Position, bs.Radius), index)

	sphereShape := &s.spheres[index]
	*sphereShape = sceneSphereShape[S]{
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
	spatialID := s.boxTree.Insert(NewCompactCubeFromSphere(bs.Position, bs.Radius), index)

	boxShape := &s.boxes[index]
	*boxShape = sceneBoxShape[S]{
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
	spatialID := s.meshTree.Insert(NewCompactCubeFromSphere(bs.Position, bs.Radius), index)

	meshShape := &s.meshes[index]
	*meshShape = sceneMeshShape[S]{
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

// FindIntersections returns an iterator over all of the intersections
// in this scene.
func (s *Scene[O, S]) CollectIntersections(collection ObjectIntersectionCollection) {
	// Sphere vs Sphere intersections.
	s.checks = s.checks[:0]
	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *sceneSphereShape[S]) {
		area := createArea(srcSphere.boundingSphere())
		s.sphereTree.QueryAABB(area, func(tgtIndex uint32) bool {
			tgtSphere := &s.spheres[tgtIndex]
			if (srcIndex < tgtIndex) && shapesCanIntersect(&srcSphere.sceneShape, &tgtSphere.sceneShape) {
				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
			}
			return true
		})
	})
	s.collectSphereSphereIntersections(s.checks, collection)

	// Sphere vs Box intersections.
	s.checks = s.checks[:0]
	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *sceneSphereShape[S]) {
		area := createArea(srcSphere.boundingSphere())
		s.boxTree.QueryAABB(area, func(tgtIndex uint32) bool {
			tgtBox := &s.boxes[tgtIndex]
			if shapesCanIntersect(&srcSphere.sceneShape, &tgtBox.sceneShape) {
				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
			}
			return true
		})
	})
	s.eachDynamicBox(func(srcIndex uint32, srcBox *sceneBoxShape[S]) {
		area := createArea(srcBox.boundingSphere())
		s.sphereTree.QueryAABB(area, func(tgtIndex uint32) bool {
			tgtSphere := &s.spheres[tgtIndex]
			if shapesCanIntersect(&tgtSphere.sceneShape, &srcBox.sceneShape) {
				s.checks = append(s.checks, newIndexPair(tgtIndex, srcIndex)) // flipped
			}
			return true
		})
	})
	s.collectSphereBoxIntersections(s.checks, collection)

	// Sphere vs Mesh intersections.
	s.checks = s.checks[:0]
	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *sceneSphereShape[S]) {
		area := createArea(srcSphere.boundingSphere())
		s.meshTree.QueryAABB(area, func(tgtIndex uint32) bool {
			tgtMesh := &s.meshes[tgtIndex]
			if shapesCanIntersect(&srcSphere.sceneShape, &tgtMesh.sceneShape) {
				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
			}
			return true
		})
	})
	s.eachDynamicMesh(func(srcIndex uint32, srcMesh *sceneMeshShape[S]) {
		area := createArea(srcMesh.boundingSphere())
		s.sphereTree.QueryAABB(area, func(tgtIndex uint32) bool {
			tgtSphere := &s.spheres[tgtIndex]
			if shapesCanIntersect(&tgtSphere.sceneShape, &srcMesh.sceneShape) {
				s.checks = append(s.checks, newIndexPair(tgtIndex, srcIndex)) // flipped
			}
			return true
		})
	})
	s.collectSphereMeshIntersections(s.checks, collection)

	// Box vs Mesh intersections.
	s.checks = s.checks[:0]
	s.eachDynamicBox(func(srcIndex uint32, srcBox *sceneBoxShape[S]) {
		area := createArea(srcBox.boundingSphere())
		s.meshTree.QueryAABB(area, func(tgtIndex uint32) bool {
			tgtMesh := &s.meshes[tgtIndex]
			if shapesCanIntersect(&srcBox.sceneShape, &tgtMesh.sceneShape) {
				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
			}
			return true
		})
	})
	s.eachDynamicMesh(func(srcIndex uint32, srcMesh *sceneMeshShape[S]) {
		area := createArea(srcMesh.boundingSphere())
		s.boxTree.QueryAABB(area, func(tgtIndex uint32) bool {
			tgtBox := &s.boxes[tgtIndex]
			if shapesCanIntersect(&tgtBox.sceneShape, &srcMesh.sceneShape) {
				s.checks = append(s.checks, newIndexPair(tgtIndex, srcIndex)) // flipped
			}
			return true
		})
	})
	s.collectBoxMeshIntersections(s.checks, collection)
}

// CheckSegmentIntersection returns the first intersection of the segment
// with the scene.
func (s *Scene[O, S]) CheckSegmentIntersection(segment Segment, mask uint32) (ObjectIntersection, bool) {
	querySegment := NewCompactQuerySegment(segment.A, segment.B)
	srcShape := sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
	}

	var collection NearestObjectIntersection

	// Segment vs Sphere
	s.checks = s.checks[:0]
	s.sphereTree.QuerySegment(querySegment, func(tgtIndex uint32) bool {
		tgtSphere := &s.spheres[tgtIndex]
		if shapesCanIntersect(&srcShape, &tgtSphere.sceneShape) {
			s.checks = append(s.checks, newIndexPair(0, tgtIndex))
		}
		return true
	})
	slices.Sort(s.checks)
	for _, checkPair := range s.checks {
		sphereIndex := checkPair.tgtIndex()
		sphere := &s.spheres[sphereIndex]
		if intersection, ok := CheckSegmentSphereIntersection(segment, sphere.wsSphere); ok {
			collection.AddIntersection(ObjectIntersection{
				SourceObjectID: InvalidObjectID,
				SourceShapeID:  InvalidShapeID,
				TargetObjectID: ObjectID(sphere.objectIndex),
				TargetShapeID:  ShapeID(newShapeRef(shapeKindSphere, sphereIndex)),
				Intersection:   intersection,
			})
		}
	}

	// Segment vs Box
	s.checks = s.checks[:0]
	s.boxTree.QuerySegment(querySegment, func(tgtIndex uint32) bool {
		tgtBox := &s.boxes[tgtIndex]
		if shapesCanIntersect(&srcShape, &tgtBox.sceneShape) {
			s.checks = append(s.checks, newIndexPair(0, tgtIndex))
		}
		return true
	})
	slices.Sort(s.checks)
	for _, checkPair := range s.checks {
		boxIndex := checkPair.tgtIndex()
		box := &s.boxes[boxIndex]
		if intersection, ok := CheckSegmentBoxIntersection(segment, box.wsBox); ok {
			collection.AddIntersection(ObjectIntersection{
				SourceObjectID: InvalidObjectID,
				SourceShapeID:  InvalidShapeID,
				TargetObjectID: ObjectID(box.objectIndex),
				TargetShapeID:  ShapeID(newShapeRef(shapeKindBox, boxIndex)),
				Intersection:   intersection,
			})
		}
	}

	// Segment vs Mesh
	s.checks = s.checks[:0]
	s.meshTree.QuerySegment(querySegment, func(tgtIndex uint32) bool {
		tgtMesh := &s.meshes[tgtIndex]
		if shapesCanIntersect(&srcShape, &tgtMesh.sceneShape) {
			s.checks = append(s.checks, newIndexPair(0, tgtIndex))
		}
		return true
	})
	slices.Sort(s.checks)
	for _, checkPair := range s.checks {
		meshIndex := checkPair.tgtIndex()
		mesh := &s.meshes[meshIndex]
		if intersection, ok := CheckSegmentMeshIntersection(segment, mesh.wsMesh); ok {
			collection.AddIntersection(ObjectIntersection{
				SourceObjectID: InvalidObjectID,
				SourceShapeID:  InvalidShapeID,
				TargetObjectID: ObjectID(mesh.objectIndex),
				TargetShapeID:  ShapeID(newShapeRef(shapeKindMesh, meshIndex)),
				Intersection:   intersection,
			})
		}
	}

	return collection.Intersection()
}

// CheckSphereIntersection returns the first intersection of the sphere
// with the scene.
func (s *Scene[O, S]) CheckSphereIntersection(srcSphere Sphere, mask uint32) (ObjectIntersection, bool) {
	srcShape := sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
	}
	area := createArea(srcSphere)

	var collection NearestObjectIntersection

	// Sphere vs Sphere
	s.checks = s.checks[:0]
	s.sphereTree.QueryAABB(area, func(tgtIndex uint32) bool {
		tgtSphere := &s.spheres[tgtIndex]
		if shapesCanIntersect(&srcShape, &tgtSphere.sceneShape) {
			s.checks = append(s.checks, newIndexPair(0, tgtIndex))
		}
		return true
	})
	slices.Sort(s.checks)
	for _, checkPair := range s.checks {
		tgtSphereIndex := checkPair.tgtIndex()
		tgtSphere := &s.spheres[tgtSphereIndex]
		if intersection, ok := CheckSphereSphereIntersection(srcSphere, tgtSphere.wsSphere); ok {
			collection.AddIntersection(ObjectIntersection{
				SourceObjectID: InvalidObjectID,
				SourceShapeID:  InvalidShapeID,
				TargetObjectID: ObjectID(tgtSphere.objectIndex),
				TargetShapeID:  ShapeID(newShapeRef(shapeKindSphere, tgtSphereIndex)),
				Intersection:   intersection,
			})
		}
	}

	// Sphere vs Box
	s.checks = s.checks[:0]
	s.boxTree.QueryAABB(area, func(tgtIndex uint32) bool {
		tgtBox := &s.boxes[tgtIndex]
		if shapesCanIntersect(&srcShape, &tgtBox.sceneShape) {
			s.checks = append(s.checks, newIndexPair(0, tgtIndex))
		}
		return true
	})
	slices.Sort(s.checks)
	for _, checkPair := range s.checks {
		tgtBoxIndex := checkPair.tgtIndex()
		tgtBox := &s.boxes[tgtBoxIndex]
		if intersection, ok := CheckSphereBoxIntersection(srcSphere, tgtBox.wsBox); ok {
			collection.AddIntersection(ObjectIntersection{
				SourceObjectID: InvalidObjectID,
				SourceShapeID:  InvalidShapeID,
				TargetObjectID: ObjectID(tgtBox.objectIndex),
				TargetShapeID:  ShapeID(newShapeRef(shapeKindBox, tgtBoxIndex)),
				Intersection:   intersection,
			})
		}
	}

	// Sphere vs Mesh
	s.checks = s.checks[:0]
	s.meshTree.QueryAABB(area, func(tgtIndex uint32) bool {
		tgtMesh := &s.meshes[tgtIndex]
		if shapesCanIntersect(&srcShape, &tgtMesh.sceneShape) {
			s.checks = append(s.checks, newIndexPair(0, tgtIndex))
		}
		return true
	})
	slices.Sort(s.checks)
	for _, checkPair := range s.checks {
		tgtMeshIndex := checkPair.tgtIndex()
		tgtMesh := &s.meshes[tgtMeshIndex]
		if intersection, ok := CheckSphereMeshIntersection(srcSphere, tgtMesh.wsMesh); ok {
			collection.AddIntersection(ObjectIntersection{
				SourceObjectID: InvalidObjectID,
				SourceShapeID:  InvalidShapeID,
				TargetObjectID: ObjectID(tgtMesh.objectIndex),
				TargetShapeID:  ShapeID(newShapeRef(shapeKindMesh, tgtMeshIndex)),
				Intersection:   intersection,
			})
		}
	}

	return collection.Intersection()
}

// GC cleans up internal data and allows for memory reuse. This should be
// called once per frame.
func (s *Scene[O, S]) GC() {
	s.sphereTree.GC()
	s.boxTree.GC()
	s.meshTree.GC()
}

func (s *Scene[O, S]) getShape(ref shapeRef) *sceneShape[S] {
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
		s.sphereTree.Remove(sphere.spatialID) // TODO: Should the ID be set to invalid?
		s.freeSphereIndices.Push(index)
	case shapeKindBox:
		box := &s.boxes[index]
		s.boxTree.Remove(box.spatialID) // TODO: Should the ID be set to invalid?
		s.freeBoxIndices.Push(index)
	case shapeKindMesh:
		mesh := &s.meshes[index]
		s.meshTree.Remove(mesh.spatialID) // TODO: Should the ID be set to invalid?
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

func (s *Scene[O, S]) collectSphereSphereIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
	lastPair := invalidIndexPair
	slices.Sort(pairs)
	for _, pair := range pairs {
		if pair != lastPair {
			srcSphereIndex := pair.srcIndex()
			srcSphere := &s.spheres[srcSphereIndex]
			tgtSphereIndex := pair.tgtIndex()
			tgtSphere := &s.spheres[tgtSphereIndex]
			if intersection, ok := s.checkSphereSphereIntersection(&srcSphere.sphereSolver, &tgtSphere.sphereSolver); ok {
				collection.AddIntersection(ObjectIntersection{
					SourceObjectID: ObjectID(srcSphere.objectIndex),
					SourceShapeID:  ShapeID(newShapeRef(shapeKindSphere, srcSphereIndex)),
					TargetObjectID: ObjectID(tgtSphere.objectIndex),
					TargetShapeID:  ShapeID(newShapeRef(shapeKindSphere, tgtSphereIndex)),
					Intersection:   intersection,
				})
			}
		}
		lastPair = pair
	}
}

func (s *Scene[O, S]) collectSphereBoxIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
	lastPair := invalidIndexPair
	slices.Sort(pairs)
	for _, pair := range pairs {
		if pair != lastPair {
			srcSphereIndex := pair.srcIndex()
			srcSphere := &s.spheres[srcSphereIndex]
			tgtBoxIndex := pair.tgtIndex()
			tgtBox := &s.boxes[tgtBoxIndex]
			if intersection, ok := s.checkSphereBoxIntersection(&srcSphere.sphereSolver, &tgtBox.boxSolver); ok {
				collection.AddIntersection(ObjectIntersection{
					SourceObjectID: ObjectID(srcSphere.objectIndex),
					SourceShapeID:  ShapeID(newShapeRef(shapeKindSphere, srcSphereIndex)),
					TargetObjectID: ObjectID(tgtBox.objectIndex),
					TargetShapeID:  ShapeID(newShapeRef(shapeKindBox, tgtBoxIndex)),
					Intersection:   intersection,
				})
			}
		}
		lastPair = pair
	}
}

func (s *Scene[O, S]) collectSphereMeshIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
	lastPair := invalidIndexPair
	slices.Sort(pairs)
	for _, pair := range pairs {
		if pair != lastPair {
			srcSphereIndex := pair.srcIndex()
			srcSphere := &s.spheres[srcSphereIndex]
			tgtMeshIndex := pair.tgtIndex()
			tgtMesh := &s.meshes[tgtMeshIndex]
			if intersection, ok := s.checkSphereMeshIntersection(&srcSphere.sphereSolver, &tgtMesh.meshSolver); ok {
				collection.AddIntersection(ObjectIntersection{
					SourceObjectID: ObjectID(srcSphere.objectIndex),
					SourceShapeID:  ShapeID(newShapeRef(shapeKindSphere, srcSphereIndex)),
					TargetObjectID: ObjectID(tgtMesh.objectIndex),
					TargetShapeID:  ShapeID(newShapeRef(shapeKindMesh, tgtMeshIndex)),
					Intersection:   intersection,
				})
			}
		}
		lastPair = pair
	}
}

func (s *Scene[O, S]) collectBoxMeshIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
	lastPair := invalidIndexPair
	slices.Sort(pairs)
	for _, pair := range pairs {
		if pair != lastPair {
			srcBoxIndex := pair.srcIndex()
			srcBox := &s.boxes[srcBoxIndex]
			tgtMeshIndex := pair.tgtIndex()
			tgtMesh := &s.meshes[tgtMeshIndex]
			if intersection, ok := s.checkBoxMeshIntersection(&srcBox.boxSolver, &tgtMesh.meshSolver); ok {
				collection.AddIntersection(ObjectIntersection{
					SourceObjectID: ObjectID(srcBox.objectIndex),
					SourceShapeID:  ShapeID(newShapeRef(shapeKindBox, srcBoxIndex)),
					TargetObjectID: ObjectID(tgtMesh.objectIndex),
					TargetShapeID:  ShapeID(newShapeRef(shapeKindMesh, tgtMeshIndex)),
					Intersection:   intersection,
				})
			}
		}
		lastPair = pair
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

func createArea(bs Sphere) CompactQueryAABB {
	return NewCompactQueryAABBFromSphere(bs.Position, bs.Radius)
}

func newIndexPair(source, target uint32) indexPair {
	return indexPair((uint64(source) << 32) | uint64(target))
}

const invalidIndexPair = indexPair(0xFFFFFFFFFFFFFFFF)

type indexPair uint64

func (p indexPair) srcIndex() uint32 {
	return uint32(p >> 32)
}

func (p indexPair) tgtIndex() uint32 {
	return uint32(p & 0xFFFFFFFF)
}
