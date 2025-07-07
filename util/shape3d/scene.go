package shape3d

import (
	"fmt"
	"slices"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/filter"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/spatial"
)

func NewScene[T any]() *Scene[T] {
	dynamicOctreeSettings := spatial.DynamicOctreeSettings{
		Size:                opt.V(16384.0),   // TODO: Configurable?
		MaxDepth:            opt.V(int32(15)), // TODO: Configurable?
		BiasRatio:           opt.V(1.0),
		InitialNodeCapacity: opt.V(int32(1024)),
		InitialItemCapacity: opt.V(int32(1024)),
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

		// TODO: Use a custom implementation that employs areas and not
		// HexahedronRegion, which is slower to evaluate and compute.
		sphereTree: spatial.NewDynamicOctree[uint32](dynamicOctreeSettings),
		boxTree:    spatial.NewDynamicOctree[uint32](dynamicOctreeSettings),
		meshTree:   spatial.NewDynamicOctree[uint32](dynamicOctreeSettings),
	}
}

type Scene[T any] struct {
	freeObjectIndices *ds.Stack[uint32]
	freeSphereIndices *ds.Stack[uint32]
	freeBoxIndices    *ds.Stack[uint32]
	freeMeshIndices   *ds.Stack[uint32]

	objects []Object[T]
	spheres []SphereShape
	boxes   []BoxShape
	meshes  []MeshShape

	sphereTree *spatial.DynamicOctree[uint32]
	boxTree    *spatial.DynamicOctree[uint32]
	meshTree   *spatial.DynamicOctree[uint32]
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
	s.eachObjectShape(object, filter.Equal(shapeKindSphere), func(index uint32) {
		sphere := &s.spheres[index]
		sphere.Update(transform)
		bs := sphere.BoundingSphere()
		s.sphereTree.Update(sphere.spatialID, bs.Position, bs.Radius)
	})
	s.eachObjectShape(object, filter.Equal(shapeKindBox), func(index uint32) {
		box := &s.boxes[index]
		box.Update(transform)
		bs := box.BoundingSphere()
		s.boxTree.Update(box.spatialID, bs.Position, bs.Radius)
	})
	s.eachObjectShape(object, filter.Equal(shapeKindMesh), func(index uint32) {
		mesh := &s.meshes[index]
		mesh.Update(transform)
		bs := mesh.BoundingSphere()
		s.meshTree.Update(mesh.spatialID, bs.Position, bs.Radius)
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
	spatialID := s.sphereTree.Insert(bs.Position, bs.Radius, index)

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
	spatialID := s.boxTree.Insert(bs.Position, bs.Radius, index)

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
	spatialID := s.meshTree.Insert(bs.Position, bs.Radius, index)

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
	panic("TODO")
	// shape := s.shapes.Get(shapeID.internalID)
	// if shape == nil {
	// 	panic("shape id is invalid")
	// }
	// object := s.objects.Get(shape.objectID)

	// s.detachObjectShape(object, shape)
	// s.updateObjectBoundary(object)
	// s.freeObjectShape(shape)
}

func createRegion(bs Sphere) spatial.HexahedronRegion {
	return spatial.CuboidRegion(
		bs.Position,
		dprec.NewVec3(bs.Radius*2.0, bs.Radius*2.0, bs.Radius*2.0),
	)
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

var checks = make([]indexPair, 0, 1024)

func shapesCanIntersect(a, b *Shape) bool {
	if a.objectIndex == b.objectIndex {
		return false
	}
	if a.rejectGroup != 0 && (a.rejectGroup == b.rejectGroup) {
		return false
	}
	if ((a.sourceMask & b.targetMask) == 0) && ((a.targetMask & b.sourceMask) == 0) {
		return false
	}
	return true
}

// FindIntersections returns an iterator over all of the intersections
// in this scene.
func (s *Scene[T]) CollectIntersections(collection ObjectIntersectionCollection) {
	sphereCount := 0
	boxCount := 0
	meshCount := 0

	// Sphere vs Sphere intersections.
	checks = checks[:0]
	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *SphereShape) {
		region := createRegion(srcSphere.BoundingSphere())
		s.sphereTree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			sphereCount++
			tgtSphere := &s.spheres[tgtIndex]
			if (srcIndex < tgtIndex) && shapesCanIntersect(&srcSphere.Shape, &tgtSphere.Shape) {
				checks = append(checks, newIndexPair(srcIndex, tgtIndex))
			}
		}))
	})
	s.collectSphereSphereIntersections(checks, collection)

	// Sphere vs Box intersections.
	checks = checks[:0]
	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *SphereShape) {
		region := createRegion(srcSphere.BoundingSphere())
		s.boxTree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			boxCount++
			tgtBox := &s.boxes[tgtIndex]
			if shapesCanIntersect(&srcSphere.Shape, &tgtBox.Shape) {
				checks = append(checks, newIndexPair(srcIndex, tgtIndex))
			}
		}))
	})
	s.eachDynamicBox(func(srcIndex uint32, srcBox *BoxShape) {
		region := createRegion(srcBox.BoundingSphere())
		s.sphereTree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			sphereCount++
			tgtSphere := &s.spheres[tgtIndex]
			if shapesCanIntersect(&tgtSphere.Shape, &srcBox.Shape) {
				checks = append(checks, newIndexPair(tgtIndex, srcIndex)) // flipped
			}
		}))
	})
	s.collectSphereBoxIntersections(checks, collection)

	// Sphere vs Mesh intersections.
	checks = checks[:0]
	s.eachDynamicSphere(func(srcIndex uint32, srcSphere *SphereShape) {
		region := createRegion(srcSphere.BoundingSphere())
		s.meshTree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			meshCount++
			tgtMesh := &s.meshes[tgtIndex]
			if shapesCanIntersect(&srcSphere.Shape, &tgtMesh.Shape) {
				checks = append(checks, newIndexPair(srcIndex, tgtIndex))
			}
		}))
	})
	s.eachDynamicMesh(func(srcIndex uint32, srcMesh *MeshShape) {
		region := createRegion(srcMesh.BoundingSphere())
		s.sphereTree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			sphereCount++
			tgtSphere := &s.spheres[tgtIndex]
			if shapesCanIntersect(&tgtSphere.Shape, &srcMesh.Shape) {
				checks = append(checks, newIndexPair(tgtIndex, srcIndex)) // flipped
			}
		}))
	})
	s.collectSphereMeshIntersections(checks, collection)

	// Box vs Mesh intersections.
	checks = checks[:0]
	s.eachDynamicBox(func(srcIndex uint32, srcBox *BoxShape) {
		region := createRegion(srcBox.BoundingSphere())
		s.meshTree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			meshCount++
			tgtMesh := &s.meshes[tgtIndex]
			if shapesCanIntersect(&srcBox.Shape, &tgtMesh.Shape) {
				checks = append(checks, newIndexPair(srcIndex, tgtIndex))
			}
		}))
	})
	s.eachDynamicMesh(func(srcIndex uint32, srcMesh *MeshShape) {
		region := createRegion(srcMesh.BoundingSphere())
		s.boxTree.VisitHexahedronRegion(&region, spatial.VisitorFunc[uint32](func(tgtIndex uint32) {
			boxCount++
			tgtBox := &s.boxes[tgtIndex]
			if shapesCanIntersect(&tgtBox.Shape, &srcMesh.Shape) {
				checks = append(checks, newIndexPair(tgtIndex, srcIndex)) // flipped
			}
		}))
	})
	s.collectBoxMeshIntersections(checks, collection)

	fmt.Printf("S: %d; B: %d; M: %d\n", sphereCount, boxCount, meshCount)
}

func (s *Scene[T]) eachDynamicSphere(cb func(uint32, *SphereShape)) {
	for index := range uint32(len(s.spheres)) {
		shape := &s.spheres[index]
		if shape.static || (shape.spatialID == spatial.InvalidDynamicOctreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[T]) eachDynamicBox(cb func(uint32, *BoxShape)) {
	for index := range uint32(len(s.boxes)) {
		shape := &s.boxes[index]
		if shape.static || (shape.spatialID == spatial.InvalidDynamicOctreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[T]) eachDynamicMesh(cb func(uint32, *MeshShape)) {
	for index := range uint32(len(s.meshes)) {
		shape := &s.meshes[index]
		if shape.static || (shape.spatialID == spatial.InvalidDynamicOctreeItemID) {
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
			if intersection, ok := s.checkSphereBoxIntersection(&srcSphere.sphereSolver, &tgtBox.boxSolver, false); ok {
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
			if intersection, ok := s.checkSphereMeshIntersection(&srcSphere.sphereSolver, &tgtMesh.meshSolver, false); ok {
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
			if intersection, ok := s.checkBoxMeshIntersection(&srcBox.boxSolver, &tgtMesh.meshSolver, false); ok {
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

func (s *Scene[T]) GC() {
	s.sphereTree.GC()
	s.boxTree.GC()
	s.meshTree.GC()
}

// var tgtObjects []mem.SparseID

// func (s *Scene[T]) eachNearbyObject(object *Object[T], cb func(*Object[T])) {
// 	bs := object.TransformedBoundingSphere()
// 	region := spatial.CuboidRegion(
// 		bs.Position,
// 		dprec.NewVec3(bs.Radius*2.0, bs.Radius*2.0, bs.Radius*2.0),
// 	)
// 	tgtObjects = tgtObjects[:0]

// 	s.objectTree.VisitHexahedronRegion(&region, spatial.VisitorFunc[mem.SparseID](func(tgtID mem.SparseID) {
// 		tgtObjects = append(tgtObjects, tgtID)
// 	}))

// 	for _, tgtID := range tgtObjects {
// 		cb(s.objects.Get(tgtID))
// 	}
// }

// func (s *Scene[T]) CheckSphereIntersection(sphere Sphere) (ObjectIntersection, bool) {
// 	checkRegion := spatial.CuboidRegion(
// 		sphere.Position,
// 		dprec.NewVec3(sphere.Radius*2.0, sphere.Radius*2.0, sphere.Radius*2.0),
// 	)

// 	var result WorstObjectIntersection
// 	s.objectTree.VisitHexahedronRegion(&checkRegion, spatial.VisitorFunc[mem.SparseID](func(objID mem.SparseID) {
// 		object := s.objects.Get(objID)
// 		if intersection, ok := s.checkIntersectionSphereWithObject(sphere, object).Unwrap(); ok {
// 			result.AddIntersection(intersection)
// 		}
// 	}))
// 	return result.Intersection()
// }

// func (s *Scene[T]) checkIntersectionSphereWithObject(sphere Sphere, object *Object[T]) opt.T[ObjectIntersection] {
// 	var result BestIntersection
// 	s.eachObjectSphere(object, func(_ mem.SparseID, shape *SphereShape) {
// 		// TODO: Track this inside the Shape object
// 		testSphere := Sphere{
// 			Position: object.transform.Apply(shape.template.Position),
// 			Radius:   shape.template.Radius,
// 		}
// 		if intersection, ok := CheckSphereSphereIntersection(sphere, testSphere); ok {
// 			result.AddIntersection(intersection)
// 		}
// 	})
// 	// TODO: Do this for boxes as well.
// 	// TODO: Do this for meshes as well.

// 	intersection, ok := result.Intersection().Unwrap()
// 	if !ok {
// 		return opt.Unspecified[ObjectIntersection]()
// 	}
// 	return opt.V(ObjectIntersection{
// 		FirstObjectID: NilObjectID(),
// 		SecondObjectID: ObjectID{
// 			internalID: object.id,
// 		},
// 		Intersection: intersection,
// 	})
// }

// func (s *Scene[T]) attachObjectShape(object *Object[T], actualID mem.SparseID, kind shapeKind) mem.SparseID {
// 	shapeID, shape := s.shapes.New()
// 	*shape = objectShape{
// 		id:          shapeID,
// 		objectID:    object.id,
// 		nextShapeID: object.firstShapeID,
// 		actualID:    actualID,
// 		kind:        kind,
// 	}
// 	object.firstShapeID = shapeID
// 	return shapeID
// }

// func (s *Scene[T]) detachObjectShape(object *Object[T], shape *objectShape) {
// 	shapeID := shape.id
// 	if object.firstShapeID == shapeID {
// 		object.firstShapeID = shape.nextShapeID
// 	} else {
// 		childID := object.firstShapeID
// 		for !childID.IsNil() {
// 			child := s.shapes.Get(childID)
// 			if child.nextShapeID == shapeID {
// 				child.nextShapeID = shape.nextShapeID
// 				break
// 			}
// 			childID = child.nextShapeID
// 		}
// 	}
// }

func (s *Scene[T]) deleteObjectShapes(object *Object[T]) {
	panic("TODO")
}

func (s *Scene[T]) eachObjectShape(object *Object[T], fltr filter.Func[shapeKind], cb func(uint32)) {
	ref := object.firstShape
	for ref != invalidShapeRef {
		nextRef := invalidShapeRef
		switch ref.Kind() {
		case shapeKindSphere:
			sphere := &s.spheres[ref.Index()]
			nextRef = sphere.nextShape
		case shapeKindBox:
			box := &s.boxes[ref.Index()]
			nextRef = box.nextShape
		case shapeKindMesh:
			mesh := &s.meshes[ref.Index()]
			nextRef = mesh.nextShape
		default:
			panic("reached an unknown shape reference")
		}
		if fltr(ref.Kind()) {
			cb(ref.Index())
		}
		ref = nextRef
	}
}

// func (s *Scene[T]) eachObjectShapeBoundingSphere(object *Object[T], cb func(Sphere)) {
// 	s.eachObjectShape(object, func(shape *objectShape) {
// 		switch shape.kind {
// 		case shapeKindSphere:
// 			sphere := s.spheres.Get(shape.actualID)
// 			cb(sphere.BoundingSphere())
// 		case shapeKindBox:
// 			box := s.boxes.Get(shape.actualID)
// 			cb(box.BoundingSphere())
// 		case shapeKindMesh:
// 			mesh := s.meshes.Get(shape.actualID)
// 			cb(mesh.BoundingSphere())
// 		}
// 	})
// }

// func (s *Scene[T]) freeObjectShape(shape *objectShape) {
// 	switch shape.kind {
// 	case shapeKindSphere:
// 		s.spheres.Delete(shape.actualID)
// 	case shapeKindBox:
// 		s.boxes.Delete(shape.actualID)
// 	case shapeKindMesh:
// 		s.meshes.Delete(shape.actualID)
// 	}
// 	s.shapes.Delete(shape.id)
// }

// func (s *Scene[T]) updateObjectBoundary(object *Object[T]) {
// 	// Calculate the centroid of all shapes.
// 	center := dprec.ZeroVec3()
// 	count := 0
// 	s.eachObjectShapeBoundingSphere(object, func(bs Sphere) {
// 		center = dprec.Vec3Sum(center, bs.Position)
// 		count++
// 	})
// 	if count > 0.0 {
// 		center = dprec.Vec3Quot(center, float64(count))
// 	}

// 	// Calculate the radius of the sphere that encompases all shapes.
// 	radius := 0.0
// 	s.eachObjectShapeBoundingSphere(object, func(bs Sphere) {
// 		distance := dprec.Vec3Diff(bs.Position, center).Length()
// 		radius = max(radius, distance+bs.Radius)
// 	})

// 	object.boundingSphere = Sphere{
// 		Position: center,
// 		Radius:   radius,
// 	}

// 	// Ensure the object's spatial placement is updated.
// 	s.invalidateObjectSpatialPlacement(object)
// }

// func (s *Scene[T]) invalidateObjectSpatialPlacement(object *Object[T]) {
// 	if object.spatialID != spatial.InvalidDynamicOctreeItemID {
// 		bs := object.TransformedBoundingSphere()
// 		s.objectTree.Update(object.spatialID, bs.Position, bs.Radius)
// 	}
// }

// func (s *Scene[T]) findObjectIntersections(srcObject *Object[T]) iter.Seq[ObjectIntersection] {
// 	return func(yield func(ObjectIntersection) bool) {
// 		if srcObject.static {
// 			return
// 		}

// 		bs := srcObject.TransformedBoundingSphere()
// 		region := spatial.CuboidRegion(
// 			bs.Position,
// 			dprec.NewVec3(bs.Radius*2.0, bs.Radius*2.0, bs.Radius*2.0),
// 		)

// 		s.objectTree.VisitHexahedronRegion(&region, spatial.VisitorFunc[mem.SparseID](func(tgtID mem.SparseID) {
// 			tgtObject := s.objects.Get(tgtID)
// 			if !tgtObject.static && (tgtID.IsBefore(srcObject.id)) {
// 				return
// 			}
// 			if srcObject.sourceGroup == tgtObject.sourceGroup {
// 				return
// 			}
// 			// TODO: Add layer checks

// 			if intersection, ok := s.checkObjectObjectIntersection(srcObject, tgtObject); ok {
// 				if !yield(intersection) {
// 					return
// 				}
// 			}
// 		}))
// 	}
// }

// func (s *Scene[T]) checkObjectObjectIntersection(source, target *Object[T]) (Intersection, bool) {
// 	var worstIntersection WorstIntersection

// 	srcBS := source.TransformedBoundingSphere()
// 	tgtBS := target.TransformedBoundingSphere()
// 	if !IsSphereSphereIntersection(srcBS, tgtBS) {
// 		return Intersection{}, false
// 	}

// 	s.eachObjectShape(source, func(srcShape *objectShape) {
// 		s.eachObjectShape(target, func(tgtShape *objectShape) {
// 			if intersection, ok := s.checkShapeShapeIntersection(srcShape, tgtShape); ok {
// 				worstIntersection.AddIntersection(intersection)
// 			}
// 		})
// 	})

// 	return worstIntersection.Intersection()
// }

// func (s *Scene[T]) checkShapeShapeIntersection(source, target *objectShape) (Intersection, bool) {
// 	switch [2]shapeKind{source.kind, target.kind} {
// 	case [2]shapeKind{shapeKindSphere, shapeKindSphere}:
// 		srcSphere := s.spheres.Get(source.actualID)
// 		tgtSphere := s.spheres.Get(target.actualID)
// 		return s.checkSphereSphereIntersection(&srcSphere.solver, &tgtSphere.solver)
// 	case [2]shapeKind{shapeKindSphere, shapeKindMesh}:
// 		srcSphere := s.spheres.Get(source.actualID)
// 		tgtMesh := s.meshes.Get(target.actualID)
// 		return s.checkSphereMeshIntersection(&srcSphere.solver, &tgtMesh.solver, false)
// 	case [2]shapeKind{shapeKindMesh, shapeKindSphere}:
// 		srcMesh := s.meshes.Get(source.actualID)
// 		tgtSphere := s.spheres.Get(target.actualID)
// 		return s.checkSphereMeshIntersection(&tgtSphere.solver, &srcMesh.solver, true)
// 	default:
// 		return Intersection{}, false
// 	}
// }

func (s *Scene[T]) checkSphereSphereIntersection(source, target *sphereSolver) (Intersection, bool) {
	return CheckSphereSphereIntersection(source.wsSphere, target.wsSphere)
}

func (s *Scene[T]) checkSphereBoxIntersection(source *sphereSolver, target *boxSolver, flip bool) (Intersection, bool) {
	if !IsSphereSphereIntersection(source.wsSphere, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	var lastIntersection LastIntersection
	if intersection, ok := CheckSphereBoxIntersection(source.wsSphere, target.wsBox); ok {
		addIntersection(&lastIntersection, flip, intersection)
	}
	return lastIntersection.Intersection()
}

func (s *Scene[T]) checkSphereMeshIntersection(source *sphereSolver, target *meshSolver, flip bool) (Intersection, bool) {
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
			addIntersection(&worstIntersection, flip, intersection)
		}
	}
	return worstIntersection.Intersection()
}

func (s *Scene[T]) checkBoxMeshIntersection(source *boxSolver, target *meshSolver, flip bool) (Intersection, bool) {
	if !IsSphereSphereIntersection(source.wsBoundingSphere, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	var lastIntersection LastIntersection
	if intersection, ok := CheckBoxMeshIntersection(source.wsBox, target.wsMesh); ok {
		addIntersection(&lastIntersection, flip, intersection)
	}
	return lastIntersection.Intersection()
}
