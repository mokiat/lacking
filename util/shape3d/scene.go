package shape3d

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/mem"
	"github.com/mokiat/lacking/util/spatial"
)

func NewScene[T any]() *Scene[T] {
	return &Scene[T]{
		objects: mem.NewSparseList[Object[T]](1024 * 8), // TODO: Configurable.
		objectTree: spatial.NewDynamicOctree[mem.SparseID](spatial.DynamicOctreeSettings{
			Size:                opt.V(32000.0),
			MaxDepth:            opt.V(int32(15)),
			BiasRatio:           opt.V(2.0),
			InitialNodeCapacity: opt.V(int32(4 * 1024)),
			InitialItemCapacity: opt.V(int32(8 * 1024)),
		}),

		shapes:  mem.NewSparseList[objectShape](1024 * 8), // TODO: Configurable.
		spheres: mem.NewSparseList[SphereShape](1024 * 8), // TODO: Configurable.
		boxes:   mem.NewSparseList[BoxShape](1024 * 8),    // TODO: Configurable.
		meshes:  mem.NewSparseList[MeshShape](1024 * 8),   // TODO: Configurable.
	}
}

type Scene[T any] struct {
	objects    *mem.SparseList[Object[T]]
	objectTree *spatial.DynamicOctree[mem.SparseID]

	shapes  *mem.SparseList[objectShape]
	spheres *mem.SparseList[SphereShape]
	boxes   *mem.SparseList[BoxShape]
	meshes  *mem.SparseList[MeshShape]
}

// CreateObject creates a new object.
func (s *Scene[T]) CreateObject(info ObjectInfo[T]) ObjectID {
	id, object := s.objects.New()

	*object = Object[T]{
		id: id,

		transform: Transform{
			Translation: info.Position.ValueOrDefault(dprec.ZeroVec3()),
			Rotation:    info.Rotation.ValueOrDefault(dprec.IdentityQuat()),
		},

		sourceMask:  info.SourceMask.ValueOrDefault(0b1),
		targetMask:  info.TargetMask.ValueOrDefault(0b1),
		rejectGroup: info.RejectGroup,

		spatialID:      spatial.InvalidDynamicOctreeItemID,
		firstShapeID:   mem.NilSparseID(),
		boundingSphere: Sphere{},

		static: info.Static,

		userData: info.UserData,
	}

	result := ObjectID{
		internalID: id,
	}

	if info.Insert.ValueOrDefault(true) {
		s.InsertObject(result)
	}

	return result
}

// DeleteObject deletes an object. If the object was inserted into the
// scene, it is first removed.
func (s *Scene[T]) DeleteObject(objID ObjectID) {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	if object.spatialID != spatial.InvalidDynamicOctreeItemID {
		s.RemoveObject(objID)
	}

	s.eachObjectShape(object, func(shape *objectShape) {
		s.freeObjectShape(shape)
	})
}

// IsValidObject returns whether an object with the specified id exists.
func (s *Scene[T]) IsValidObject(objID ObjectID) bool {
	return s.objects.Has(objID.internalID)
}

// InsertObject adds the object to the scene, making it visible to other
// objects.
func (s *Scene[T]) InsertObject(objID ObjectID) {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	if object.spatialID != spatial.InvalidDynamicOctreeItemID {
		panic("object already inserted")
	}

	bs := object.TransformedBoundingSphere()
	spatialID := s.objectTree.Insert(bs.Position, bs.Radius, objID.internalID)
	object.spatialID = spatialID
}

// RemoveObject removes the object from the scene, making it invisible to
// other objects.
func (s *Scene[T]) RemoveObject(objID ObjectID) {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	spatialID := object.spatialID
	if spatialID == spatial.InvalidDynamicOctreeItemID {
		panic("object already removed")
	}

	s.objectTree.Remove(spatialID)
	object.spatialID = spatial.InvalidDynamicOctreeItemID
}

// SetUserData assigns the specified user data to the object.
func (s *Scene[T]) SetUserData(objID ObjectID, value T) {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}
	object.userData = value
}

// GetUserData returns the user data associated with the given object.
func (s *Scene[T]) GetUserData(objID ObjectID) T {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}
	return object.userData
}

// SetTransform relocates the given object.
func (s *Scene[T]) SetTransform(objID ObjectID, transform Transform) {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	object.transform = transform
	s.eachObjectShape(object, func(shape *objectShape) {
		switch shape.kind {
		case shapeKindSphere:
			sphere := s.spheres.Get(shape.actualID)
			sphere.SetTransform(transform)
		case shapeKindBox:
			box := s.boxes.Get(shape.actualID)
			box.SetTransform(transform)
		case shapeKindMesh:
			mesh := s.meshes.Get(shape.actualID)
			mesh.SetTransform(transform)
		}
	})

	s.invalidateObjectSpatialPlacement(object)
}

// GetTransform returns the given object's transform.
func (s *Scene[T]) GetTransform(objID ObjectID) Transform {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}
	return object.transform
}

// AttachSphere creates a sphere shape and attaches it to the object to be
// used for intersection tests. The position of the sphere is relative to the
// object's origin.
func (s *Scene[T]) AttachSphere(objID ObjectID, template Sphere) ShapeID {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	sphereID, sphereShape := s.spheres.New()
	sphereShape.Init(template, object.transform)

	shapeID := s.attachObjectShape(object, sphereID, shapeKindSphere)
	s.updateObjectBoundary(object)

	return ShapeID{
		internalID: shapeID,
	}
}

// AttachBox creates a box shape and attaches it to the object to be used for
// intersection tests. The position and rotation of the box is relative to the
// object's transform.
func (s *Scene[T]) AttachBox(objID ObjectID, template Box) BoxID {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	boxID, boxShape := s.boxes.New()
	boxShape.Init(template, object.transform)

	shapeID := s.attachObjectShape(object, boxID, shapeKindBox)
	s.updateObjectBoundary(object)

	return BoxID{
		internalID: shapeID,
	}
}

// AttachMesh creates a mesh shape and attaches it to the object to be used for
// intersection tests. The position and rotation of the mesh is relative to the
// object's transform.
func (s *Scene[T]) AttachMesh(objID ObjectID, template Mesh) MeshID {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	meshID, meshShape := s.meshes.New()
	meshShape.Init(template, object.transform)

	shapeID := s.attachObjectShape(object, meshID, shapeKindMesh)
	s.updateObjectBoundary(object)

	return MeshID{
		internalID: shapeID,
	}
}

// DeleteShape deletes a shape from an object. The object is not
// deleted and continues to exist in the scene.
func (s *Scene[T]) DeleteShape(shapeID ShapeID) {
	shape := s.shapes.Get(shapeID.internalID)
	if shape == nil {
		panic("shape id is invalid")
	}
	object := s.objects.Get(shape.objectID)

	s.detachObjectShape(object, shape)
	s.updateObjectBoundary(object)
	s.freeObjectShape(shape)
}

// IsValidShape returns whether a shape with the specified id exists.
func (s *Scene[T]) IsValidShape(shapeID ShapeID) bool {
	return s.shapes.Has(shapeID.internalID)
}

// FindIntersections returns an iterator over all of the intersections
// in this scene.
func (s *Scene[T]) CollectIntersections(collection ObjectIntersectionCollection) {
	var objectCount int
	s.objects.Each(func(mem.SparseID, *Object[T]) {
		objectCount++
	})
	var sphereCount int
	s.spheres.Each(func(mem.SparseID, *SphereShape) {
		sphereCount++
	})
	var meshCount int
	s.meshes.Each(func(mem.SparseID, *MeshShape) {
		meshCount++
	})
	fmt.Println("Objects:", objectCount, "Spheres", sphereCount, "Meshes", meshCount)

	s.objects.Each(func(_ mem.SparseID, srcObject *Object[T]) {
		if srcObject.static {
			return
		}

		s.eachNearbyObject(srcObject, func(tgtObject *Object[T]) {
			if srcObject.id == tgtObject.id {
				return
			}
			if !tgtObject.static && (tgtObject.id.IsBefore(srcObject.id)) {
				return
			}
			if (tgtObject.rejectGroup != 0) && (tgtObject.rejectGroup == srcObject.rejectGroup) {
				return
			}
			if ((srcObject.targetMask & tgtObject.sourceMask) == 0) && ((srcObject.sourceMask & tgtObject.targetMask) == 0) {
				return
			}

			if intersection, ok := s.checkObjectObjectIntersection(srcObject, tgtObject); ok {
				collection.AddIntersection(ObjectIntersection{
					FirstObjectID:  srcObject.ObjectID(),
					SecondObjectID: tgtObject.ObjectID(),
					Intersection:   intersection,
				})
			}
		})
	})
}

func (s *Scene[T]) GC() {
	s.objectTree.GC()
}

func (s *Scene[T]) eachNearbyObject(object *Object[T], cb func(*Object[T])) {
	bs := object.TransformedBoundingSphere()
	region := spatial.CuboidRegion(
		bs.Position,
		dprec.NewVec3(bs.Radius*2.0, bs.Radius*2.0, bs.Radius*2.0),
	)
	s.objectTree.VisitHexahedronRegion(&region, spatial.VisitorFunc[mem.SparseID](func(tgtID mem.SparseID) {
		cb(s.objects.Get(tgtID))
	}))
}

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

func (s *Scene[T]) attachObjectShape(object *Object[T], actualID mem.SparseID, kind shapeKind) mem.SparseID {
	shapeID, shape := s.shapes.New()
	*shape = objectShape{
		id:          shapeID,
		objectID:    object.id,
		nextShapeID: object.firstShapeID,
		actualID:    actualID,
		kind:        kind,
	}
	object.firstShapeID = shapeID
	return shapeID
}

func (s *Scene[T]) detachObjectShape(object *Object[T], shape *objectShape) {
	shapeID := shape.id
	if object.firstShapeID == shapeID {
		object.firstShapeID = shape.nextShapeID
	} else {
		childID := object.firstShapeID
		for !childID.IsNil() {
			child := s.shapes.Get(childID)
			if child.nextShapeID == shapeID {
				child.nextShapeID = shape.nextShapeID
				break
			}
			childID = child.nextShapeID
		}
	}
}

func (s *Scene[T]) eachObjectShape(object *Object[T], cb func(*objectShape)) {
	shapeID := object.firstShapeID
	for !shapeID.IsNil() {
		shape := s.shapes.Get(shapeID)
		shapeID = shape.nextShapeID // set before callback to allow deletion
		cb(shape)
	}
}

func (s *Scene[T]) eachObjectShapeBoundingSphere(object *Object[T], cb func(Sphere)) {
	s.eachObjectShape(object, func(shape *objectShape) {
		switch shape.kind {
		case shapeKindSphere:
			sphere := s.spheres.Get(shape.actualID)
			cb(sphere.BoundingSphere())
		case shapeKindBox:
			box := s.boxes.Get(shape.actualID)
			cb(box.BoundingSphere())
		case shapeKindMesh:
			mesh := s.meshes.Get(shape.actualID)
			cb(mesh.BoundingSphere())
		}
	})
}

func (s *Scene[T]) freeObjectShape(shape *objectShape) {
	switch shape.kind {
	case shapeKindSphere:
		s.spheres.Delete(shape.actualID)
	case shapeKindBox:
		s.boxes.Delete(shape.actualID)
	case shapeKindMesh:
		s.meshes.Delete(shape.actualID)
	}
	s.shapes.Delete(shape.id)
}

func (s *Scene[T]) updateObjectBoundary(object *Object[T]) {
	// Calculate the centroid of all shapes.
	center := dprec.ZeroVec3()
	count := 0
	s.eachObjectShapeBoundingSphere(object, func(bs Sphere) {
		center = dprec.Vec3Sum(center, bs.Position)
		count++
	})
	if count > 0.0 {
		center = dprec.Vec3Quot(center, float64(count))
	}

	// Calculate the radius of the sphere that encompases all shapes.
	radius := 0.0
	s.eachObjectShapeBoundingSphere(object, func(bs Sphere) {
		distance := dprec.Vec3Diff(bs.Position, center).Length()
		radius = max(radius, distance+bs.Radius)
	})

	object.boundingSphere = Sphere{
		Position: center,
		Radius:   radius,
	}

	// Ensure the object's spatial placement is updated.
	s.invalidateObjectSpatialPlacement(object)
}

func (s *Scene[T]) invalidateObjectSpatialPlacement(object *Object[T]) {
	if object.spatialID != spatial.InvalidDynamicOctreeItemID {
		bs := object.TransformedBoundingSphere()
		s.objectTree.Update(object.spatialID, bs.Position, bs.Radius)
	}
}

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

func (s *Scene[T]) checkObjectObjectIntersection(source, target *Object[T]) (Intersection, bool) {
	var worstIntersection WorstIntersection

	srcBS := source.TransformedBoundingSphere()
	tgtBS := target.TransformedBoundingSphere()
	if !IsSphereSphereIntersection(srcBS, tgtBS) {
		return Intersection{}, false
	}

	s.eachObjectShape(source, func(srcShape *objectShape) {
		s.eachObjectShape(target, func(tgtShape *objectShape) {
			if intersection, ok := s.checkShapeShapeIntersection(srcShape, tgtShape); ok {
				worstIntersection.AddIntersection(intersection)
			}
		})
	})

	return worstIntersection.Intersection()
}

func (s *Scene[T]) checkShapeShapeIntersection(source, target *objectShape) (Intersection, bool) {
	switch [2]shapeKind{source.kind, target.kind} {
	case [2]shapeKind{shapeKindSphere, shapeKindSphere}:
		srcSphere := s.spheres.Get(source.actualID)
		tgtSphere := s.spheres.Get(target.actualID)
		return s.checkSphereSphereIntersection(&srcSphere.solver, &tgtSphere.solver)
	case [2]shapeKind{shapeKindSphere, shapeKindMesh}:
		srcSphere := s.spheres.Get(source.actualID)
		tgtMesh := s.meshes.Get(target.actualID)
		return s.checkSphereMeshIntersection(&srcSphere.solver, &tgtMesh.solver, false)
	case [2]shapeKind{shapeKindMesh, shapeKindSphere}:
		srcMesh := s.meshes.Get(source.actualID)
		tgtSphere := s.spheres.Get(target.actualID)
		return s.checkSphereMeshIntersection(&tgtSphere.solver, &srcMesh.solver, true)
	default:
		return Intersection{}, false
	}
}

func (s *Scene[T]) checkSphereSphereIntersection(source, target *sphereSolver) (Intersection, bool) {
	return CheckSphereSphereIntersection(source.wsSphere, target.wsSphere)
}

func (s *Scene[T]) checkSphereMeshIntersection(source *sphereSolver, target *meshSolver, flip bool) (Intersection, bool) {
	if !IsSphereSphereIntersection(source.wsSphere, target.wsBoundingSphere) {
		return Intersection{}, false
	}
	var worstIntersection WorstIntersection
	wsSphere := source.wsSphere
	for _, wsTriangle := range target.wsTriangles {
		if !IsSphereSphereIntersection(wsSphere, wsTriangle.BoundingSphere()) {
			continue
		}
		if intersection, ok := CheckSphereTriangleIntersection(wsSphere, wsTriangle); ok {
			addIntersection(&worstIntersection, flip, intersection)
		}
	}
	return worstIntersection.Intersection()
}
