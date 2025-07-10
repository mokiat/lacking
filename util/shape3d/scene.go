package shape3d

import (
	"slices"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/spatial"
)

// SceneInfo contains information needed to create an optimal scene.
type SceneInfo struct {

	// Size specifies the dimension (from side to side) of the scene.
	// Inserting an item outside these bounds has undefined behavior.
	Size opt.T[float64]

	// MaxDepth controls the maximum depth that the underlying octree can reach.
	MaxDepth opt.T[int32]

	// InitialNodeCapacity is a hint as to the number of nodes that will be
	// needed to place all items. Usually one would find that number empirically.
	// This allows the octree to preallocate memory and avoid dynamic allocations.
	InitialNodeCapacity opt.T[int32]

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the octree. This allows the octree to preallocate
	// memory and avoid dynamic allocations during insertion.
	InitialItemCapacity opt.T[int32]
}

// NewScene creates a new scene.
func NewScene[T any](info SceneInfo) *Scene[T] {
	cubeOctreeSettings := spatial.CompactOctreeSettings{
		Size:                info.Size,
		MaxDepth:            info.MaxDepth,
		InitialNodeCapacity: info.InitialNodeCapacity,
		InitialItemCapacity: info.InitialItemCapacity,
	}

	return &Scene[T]{
		freeObjectIndices: ds.NewStack[uint32](256), // ~ 1 KiB
		freeSphereIndices: ds.NewStack[uint32](256), // ~ 1 KiB
		freeBoxIndices:    ds.NewStack[uint32](256), // ~ 1 KiB
		freeMeshIndices:   ds.NewStack[uint32](256), // ~ 1 KiB

		objects: make([]Object[T], 0, 128),
		spheres: make([]SphereShape, 0, 128),
		boxes:   make([]BoxShape, 0, 128),
		meshes:  make([]MeshShape, 0, 128),

		sphereTree: spatial.NewCompactOctree[uint32](cubeOctreeSettings),
		boxTree:    spatial.NewCompactOctree[uint32](cubeOctreeSettings),
		meshTree:   spatial.NewCompactOctree[uint32](cubeOctreeSettings),

		checks: make([]indexPair, 0, 1024),
	}
}

// Scene represents a 3D scene where objects made of shapes can be added.
type Scene[T any] struct {
	freeObjectIndices *ds.Stack[uint32]
	freeSphereIndices *ds.Stack[uint32]
	freeBoxIndices    *ds.Stack[uint32]
	freeMeshIndices   *ds.Stack[uint32]

	objects []Object[T]
	spheres []SphereShape
	boxes   []BoxShape
	meshes  []MeshShape

	sphereTree *spatial.CompactOctree[uint32]
	boxTree    *spatial.CompactOctree[uint32]
	meshTree   *spatial.CompactOctree[uint32]

	checks []indexPair
}

// CreateObject creates a new object.
func (s *Scene[T]) CreateObject(info ObjectInfo[T]) ObjectID {
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
		s.objects = append(s.objects, Object[T]{
			transform:  transform,
			firstShape: invalidShapeRef,
			flags:      flags,
			userData:   info.UserData,
		})
		return ObjectID(index)
	} else {
		index := s.freeObjectIndices.Pop()
		s.objects[index] = Object[T]{
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
func (s *Scene[T]) DeleteObject(objID ObjectID) {
	object := &s.objects[objID]
	s.deleteObjectShapes(object)
	object.firstShape = invalidShapeRef
	object.userData = gog.Zero[T]() // in case of pointer
	s.freeObjectIndices.Push(uint32(objID))
}

// GetObjectUserData returns the user data associated with the given object.
func (s *Scene[T]) GetObjectUserData(objID ObjectID) T {
	object := &s.objects[objID]
	return object.userData
}

// SetObjectUserData assigns the specified user data to the object.
func (s *Scene[T]) SetObjectUserData(objID ObjectID, userData T) {
	object := &s.objects[objID]
	object.userData = userData
}

// GetObjectTransform returns the given object's transform.
func (s *Scene[T]) GetObjectTransform(objID ObjectID) Transform {
	object := &s.objects[objID]
	return object.transform
}

// SetObjectTransform relocates the given object.
func (s *Scene[T]) SetObjectTransform(objID ObjectID, transform Transform) {
	object := &s.objects[objID]
	object.transform = transform
	s.eachObjectShape(object, shapeKindSphere, func(index uint32) {
		sphere := &s.spheres[index]
		sphere.Update(transform)
		bs := sphere.BoundingSphere()
		s.sphereTree.Update(sphere.spatialID, spatial.CubeAreaFromSphere(bs.Position, bs.Radius))
	})
	s.eachObjectShape(object, shapeKindBox, func(index uint32) {
		box := &s.boxes[index]
		box.Update(transform)
		bs := box.BoundingSphere()
		s.boxTree.Update(box.spatialID, spatial.CubeAreaFromSphere(bs.Position, bs.Radius))
	})
	s.eachObjectShape(object, shapeKindMesh, func(index uint32) {
		mesh := &s.meshes[index]
		mesh.Update(transform)
		bs := mesh.BoundingSphere()
		s.meshTree.Update(mesh.spatialID, spatial.CubeAreaFromSphere(bs.Position, bs.Radius))
	})
}

// AttachSphere creates a sphere shape and attaches it to the object to be
// used for intersection tests.
func (s *Scene[T]) AttachSphere(objID ObjectID, info SphereInfo) ShapeID {
	var index uint32
	if s.freeSphereIndices.IsEmpty() {
		index = uint32(len(s.spheres))
		s.spheres = append(s.spheres, SphereShape{})
	} else {
		index = s.freeSphereIndices.Pop()
	}
	ref := newShapeRef(shapeKindSphere, index)

	object := &s.objects[objID]

	solver := newSphereSolver(info.Sphere)
	solver.Update(object.transform)

	bs := solver.BoundingSphere()
	spatialID := s.sphereTree.Insert(spatial.CubeAreaFromSphere(bs.Position, bs.Radius), index)

	sphereShape := &s.spheres[index]
	*sphereShape = SphereShape{
		Shape: Shape{
			objectIndex: uint32(objID),
			nextShape:   object.firstShape,
			spatialID:   spatialID,
			static:      object.IsStatic(),
			rejectGroup: info.RejectGroup,
			sourceMask:  info.SourceMask.ValueOrDefault(0b1),
			targetMask:  info.TargetMask.ValueOrDefault(0b1),
		},
		sphereSolver: solver,
	}

	object.firstShape = ref
	return ShapeID(ref)
}

// AttachBox creates a box shape and attaches it to the object to be used for
// intersection tests.
func (s *Scene[T]) AttachBox(objID ObjectID, info BoxInfo) ShapeID {
	var index uint32
	if s.freeBoxIndices.IsEmpty() {
		index = uint32(len(s.boxes))
		s.boxes = append(s.boxes, BoxShape{})
	} else {
		index = s.freeBoxIndices.Pop()
	}
	ref := newShapeRef(shapeKindBox, index)

	object := &s.objects[objID]

	solver := newBoxSolver(info.Box)
	solver.Update(object.transform)

	bs := solver.BoundingSphere()
	spatialID := s.boxTree.Insert(spatial.CubeAreaFromSphere(bs.Position, bs.Radius), index)

	boxShape := &s.boxes[index]
	*boxShape = BoxShape{
		Shape: Shape{
			objectIndex: uint32(objID),
			nextShape:   object.firstShape,
			spatialID:   spatialID,
			static:      object.IsStatic(),
			rejectGroup: info.RejectGroup,
			sourceMask:  info.SourceMask.ValueOrDefault(0b1),
			targetMask:  info.TargetMask.ValueOrDefault(0b1),
		},
		boxSolver: solver,
	}

	object.firstShape = ref
	return ShapeID(ref)
}

// AttachMesh creates a mesh shape and attaches it to the object to be used for
// intersection tests.
func (s *Scene[T]) AttachMesh(objID ObjectID, info MeshInfo) ShapeID {
	var index uint32
	if s.freeMeshIndices.IsEmpty() {
		index = uint32(len(s.meshes))
		s.meshes = append(s.meshes, MeshShape{})
	} else {
		index = s.freeMeshIndices.Pop()
	}
	ref := newShapeRef(shapeKindMesh, index)

	object := &s.objects[objID]

	solver := newMeshSolver(info.Mesh)
	solver.Update(object.transform)

	bs := solver.BoundingSphere()
	spatialID := s.meshTree.Insert(spatial.CubeAreaFromSphere(bs.Position, bs.Radius), index)

	meshShape := &s.meshes[index]
	*meshShape = MeshShape{
		Shape: Shape{
			objectIndex: uint32(objID),
			nextShape:   object.firstShape,
			spatialID:   spatialID,
			static:      object.IsStatic(),
			rejectGroup: info.RejectGroup,
			sourceMask:  info.SourceMask.ValueOrDefault(0b1),
			targetMask:  info.TargetMask.ValueOrDefault(0b1),
		},
		meshSolver: solver,
	}

	object.firstShape = ref
	return ShapeID(ref)
}

// DeleteShape deletes a shape from an object. The object is not
// deleted and continues to exist in the scene.
func (s *Scene[T]) DeleteShape(shapeID ShapeID) {
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
func (s *Scene[T]) CollectIntersections(collection ObjectIntersectionCollection) {
	// Sphere vs Sphere intersections.
	s.checks = s.checks[:0]
	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *SphereShape) {
		area := createArea(srcSphere.BoundingSphere())
		s.sphereTree.VisitArea(area, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			tgtSphere := &s.spheres[tgtIndex]
			if (srcIndex < tgtIndex) && shapesCanIntersect(&srcSphere.Shape, &tgtSphere.Shape) {
				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
			}
		}))
	})
	s.collectSphereSphereIntersections(s.checks, collection)

	// Sphere vs Box intersections.
	s.checks = s.checks[:0]
	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *SphereShape) {
		area := createArea(srcSphere.BoundingSphere())
		s.boxTree.VisitArea(area, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			tgtBox := &s.boxes[tgtIndex]
			if shapesCanIntersect(&srcSphere.Shape, &tgtBox.Shape) {
				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
			}
		}))
	})
	s.eachDynamicBox(func(srcIndex uint32, srcBox *BoxShape) {
		area := createArea(srcBox.BoundingSphere())
		s.sphereTree.VisitArea(area, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			tgtSphere := &s.spheres[tgtIndex]
			if shapesCanIntersect(&tgtSphere.Shape, &srcBox.Shape) {
				s.checks = append(s.checks, newIndexPair(tgtIndex, srcIndex)) // flipped
			}
		}))
	})
	s.collectSphereBoxIntersections(s.checks, collection)

	// Sphere vs Mesh intersections.
	s.checks = s.checks[:0]
	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *SphereShape) {
		area := createArea(srcSphere.BoundingSphere())
		s.meshTree.VisitArea(area, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			tgtMesh := &s.meshes[tgtIndex]
			if shapesCanIntersect(&srcSphere.Shape, &tgtMesh.Shape) {
				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
			}
		}))
	})
	s.eachDynamicMesh(func(srcIndex uint32, srcMesh *MeshShape) {
		area := createArea(srcMesh.BoundingSphere())
		s.sphereTree.VisitArea(area, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			tgtSphere := &s.spheres[tgtIndex]
			if shapesCanIntersect(&tgtSphere.Shape, &srcMesh.Shape) {
				s.checks = append(s.checks, newIndexPair(tgtIndex, srcIndex)) // flipped
			}
		}))
	})
	s.collectSphereMeshIntersections(s.checks, collection)

	// Box vs Mesh intersections.
	s.checks = s.checks[:0]
	s.eachDynamicBox(func(srcIndex uint32, srcBox *BoxShape) {
		area := createArea(srcBox.BoundingSphere())
		s.meshTree.VisitArea(area, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			tgtMesh := &s.meshes[tgtIndex]
			if shapesCanIntersect(&srcBox.Shape, &tgtMesh.Shape) {
				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
			}
		}))
	})
	s.eachDynamicMesh(func(srcIndex uint32, srcMesh *MeshShape) {
		area := createArea(srcMesh.BoundingSphere())
		s.boxTree.VisitArea(area, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			tgtBox := &s.boxes[tgtIndex]
			if shapesCanIntersect(&tgtBox.Shape, &srcMesh.Shape) {
				s.checks = append(s.checks, newIndexPair(tgtIndex, srcIndex)) // flipped
			}
		}))
	})
	s.collectBoxMeshIntersections(s.checks, collection)
}

// CheckSegmentIntersection returns the first intersection of the segment
// with the scene.
func (s *Scene[T]) CheckSegmentIntersection(segment Segment, mask uint32) (ObjectIntersection, bool) {
	srcShape := Shape{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
	}

	var collection BestObjectIntersection

	// Segment vs Sphere
	s.checks = s.checks[:0]
	s.sphereTree.VisitSegment(segment.A, segment.B, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
		tgtSphere := &s.spheres[tgtIndex]
		if shapesCanIntersect(&srcShape, &tgtSphere.Shape) {
			s.checks = append(s.checks, newIndexPair(0, tgtIndex))
		}
	}))
	slices.Sort(s.checks)
	for _, sphereIndex := range s.checks {
		sphere := &s.spheres[sphereIndex]
		if intersection, ok := CheckSegmentSphereIntersection(segment, sphere.wsSphere); ok {
			collection.AddIntersection(ObjectIntersection{
				FirstObjectID:  InvalidObjectID,
				SecondObjectID: ObjectID(sphere.objectIndex),
				Intersection:   intersection,
			})
		}
	}

	// Segment vs Box
	s.checks = s.checks[:0]
	s.boxTree.VisitSegment(segment.A, segment.B, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
		tgtBox := &s.boxes[tgtIndex]
		if shapesCanIntersect(&srcShape, &tgtBox.Shape) {
			s.checks = append(s.checks, newIndexPair(0, tgtIndex))
		}
	}))
	slices.Sort(s.checks)
	for _, boxIndex := range s.checks {
		box := &s.boxes[boxIndex]
		if intersection, ok := CheckSegmentBoxIntersection(segment, box.wsBox); ok {
			collection.AddIntersection(ObjectIntersection{
				FirstObjectID:  InvalidObjectID,
				SecondObjectID: ObjectID(box.objectIndex),
				Intersection:   intersection,
			})
		}
	}

	// Segment vs Mesh
	s.checks = s.checks[:0]
	s.meshTree.VisitSegment(segment.A, segment.B, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
		tgtMesh := &s.meshes[tgtIndex]
		if shapesCanIntersect(&srcShape, &tgtMesh.Shape) {
			s.checks = append(s.checks, newIndexPair(0, tgtIndex))
		}
	}))
	slices.Sort(s.checks)
	for _, meshIndex := range s.checks {
		mesh := &s.meshes[meshIndex]
		if intersection, ok := CheckSegmentMeshIntersection(segment, mesh.wsMesh); ok {
			collection.AddIntersection(ObjectIntersection{
				FirstObjectID:  InvalidObjectID,
				SecondObjectID: ObjectID(mesh.objectIndex),
				Intersection:   intersection,
			})
		}
	}

	return collection.Intersection()
}

// GC cleans up internal data and allows for memory reuse. This should be
// called once per frame.
func (s *Scene[T]) GC() {
	s.sphereTree.GC()
	s.boxTree.GC()
	s.meshTree.GC()
}

func (s *Scene[T]) getShape(ref shapeRef) *Shape {
	switch ref.Kind() {
	case shapeKindSphere:
		sphere := &s.spheres[ref.Index()]
		return &sphere.Shape
	case shapeKindBox:
		box := &s.boxes[ref.Index()]
		return &box.Shape
	case shapeKindMesh:
		mesh := &s.meshes[ref.Index()]
		return &mesh.Shape
	default:
		panic("unknown shape reference")
	}
}

func (s *Scene[T]) freeShape(ref shapeRef) {
	switch ref.Kind() {
	case shapeKindSphere:
		s.freeSphereIndices.Push(ref.Index())
	case shapeKindBox:
		s.freeBoxIndices.Push(ref.Index())
	case shapeKindMesh:
		s.freeMeshIndices.Push(ref.Index())
	default:
		panic("unknown shape reference")
	}
}

func (s *Scene[T]) eachDynamicSphere(cb func(uint32, *SphereShape)) {
	for index := range uint32(len(s.spheres)) {
		shape := &s.spheres[index]
		if shape.static || (shape.spatialID == spatial.InvalidCompactOctreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[T]) eachDynamicBox(cb func(uint32, *BoxShape)) {
	for index := range uint32(len(s.boxes)) {
		shape := &s.boxes[index]
		if shape.static || (shape.spatialID == spatial.InvalidCompactOctreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[T]) eachDynamicMesh(cb func(uint32, *MeshShape)) {
	for index := range uint32(len(s.meshes)) {
		shape := &s.meshes[index]
		if shape.static || (shape.spatialID == spatial.InvalidCompactOctreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[T]) collectSphereSphereIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
	var lastPair indexPair
	slices.Sort(pairs)
	for _, pair := range pairs {
		if pair != lastPair {
			srcSphere := &s.spheres[pair.srcIndex()]
			tgtSphere := &s.spheres[pair.tgtIndex()]
			if intersection, ok := s.checkSphereSphereIntersection(&srcSphere.sphereSolver, &tgtSphere.sphereSolver); ok {
				collection.AddIntersection(ObjectIntersection{
					FirstObjectID:  ObjectID(srcSphere.objectIndex),
					SecondObjectID: ObjectID(tgtSphere.objectIndex),
					Intersection:   intersection,
				})
			}
		}
		lastPair = pair
	}
}

func (s *Scene[T]) collectSphereBoxIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
	var lastPair indexPair
	slices.Sort(pairs)
	for _, pair := range pairs {
		if pair != lastPair {
			srcSphere := &s.spheres[pair.srcIndex()]
			tgtBox := &s.boxes[pair.tgtIndex()]
			if intersection, ok := s.checkSphereBoxIntersection(&srcSphere.sphereSolver, &tgtBox.boxSolver); ok {
				collection.AddIntersection(ObjectIntersection{
					FirstObjectID:  ObjectID(srcSphere.objectIndex),
					SecondObjectID: ObjectID(tgtBox.objectIndex),
					Intersection:   intersection,
				})
			}
		}
		lastPair = pair
	}
}

func (s *Scene[T]) collectSphereMeshIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
	var lastPair indexPair
	slices.Sort(pairs)
	for _, pair := range pairs {
		if pair != lastPair {
			srcSphere := &s.spheres[pair.srcIndex()]
			tgtMesh := &s.meshes[pair.tgtIndex()]
			if intersection, ok := s.checkSphereMeshIntersection(&srcSphere.sphereSolver, &tgtMesh.meshSolver); ok {
				collection.AddIntersection(ObjectIntersection{
					FirstObjectID:  ObjectID(srcSphere.objectIndex),
					SecondObjectID: ObjectID(tgtMesh.objectIndex),
					Intersection:   intersection,
				})
			}
		}
		lastPair = pair
	}
}

func (s *Scene[T]) collectBoxMeshIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
	var lastPair indexPair
	slices.Sort(pairs)
	for _, pair := range pairs {
		if pair != lastPair {
			srcBox := &s.boxes[pair.srcIndex()]
			tgtMesh := &s.meshes[pair.tgtIndex()]
			if intersection, ok := s.checkBoxMeshIntersection(&srcBox.boxSolver, &tgtMesh.meshSolver); ok {
				collection.AddIntersection(ObjectIntersection{
					FirstObjectID:  ObjectID(srcBox.objectIndex),
					SecondObjectID: ObjectID(tgtMesh.objectIndex),
					Intersection:   intersection,
				})
			}
		}
		lastPair = pair
	}
}

func (s *Scene[T]) deleteObjectShapes(object *Object[T]) {
	ref := object.firstShape
	for ref != invalidShapeRef {
		shape := s.getShape(ref)
		nextRef := shape.nextShape
		s.freeShape(ref)
		ref = nextRef
	}
}

func (s *Scene[T]) eachObjectShape(object *Object[T], kind shapeKind, cb func(uint32)) {
	ref := object.firstShape
	for ref != invalidShapeRef {
		shape := s.getShape(ref)
		nextRef := shape.nextShape
		if ref.Kind() == kind {
			cb(ref.Index())
		}
		ref = nextRef
	}
}

func (s *Scene[T]) checkSphereSphereIntersection(source, target *sphereSolver) (Intersection, bool) {
	return CheckSphereSphereIntersection(source.wsSphere, target.wsSphere)
}

func (s *Scene[T]) checkSphereBoxIntersection(source *sphereSolver, target *boxSolver) (Intersection, bool) {
	if !IsSphereSphereIntersection(source.wsSphere, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	return CheckSphereBoxIntersection(source.wsSphere, target.wsBox)
}

func (s *Scene[T]) checkSphereMeshIntersection(source *sphereSolver, target *meshSolver) (Intersection, bool) {
	if !IsSphereSphereIntersection(source.wsSphere, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	var worstIntersection WorstIntersection
	wsSphere := source.wsSphere
	for _, wsTriangle := range target.wsMesh.Triangles {
		if !IsSphereSphereIntersection(wsSphere, wsTriangle.BoundingSphere()) {
			continue
		}
		if intersection, ok := CheckSphereTriangleIntersection(wsSphere, wsTriangle); ok {
			worstIntersection.AddIntersection(intersection)
		}
	}
	return worstIntersection.Intersection()
}

func (s *Scene[T]) checkBoxMeshIntersection(source *boxSolver, target *meshSolver) (Intersection, bool) {
	if !IsSphereSphereIntersection(source.wsBoundingSphere, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	return CheckBoxMeshIntersection(source.wsBox, target.wsMesh)
}

func createArea(bs Sphere) spatial.CubeArea {
	return spatial.CubeAreaFromSphere(bs.Position, bs.Radius)
}

func newIndexPair(source, target uint32) indexPair {
	return indexPair((uint64(source) << 32) | uint64(target))
}

type indexPair uint64

func (p indexPair) srcIndex() uint32 {
	return uint32(p >> 32)
}

func (p indexPair) tgtIndex() uint32 {
	return uint32(p & 0xFFFFFFFF)
}
