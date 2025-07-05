package shape3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/mem"
	"github.com/mokiat/lacking/util/spatial"
)

func NewScene[T any]() *Scene[T] {
	return &Scene[T]{
		objects: mem.NewSparseList[Object[T]](64), // TODO: Configurable.
		objectTree: spatial.NewDynamicOctree[mem.SparseID](spatial.DynamicOctreeSettings{
			Size:                opt.V(32000.0),
			MaxDepth:            opt.V(int32(15)),
			BiasRatio:           opt.V(2.0),
			InitialNodeCapacity: opt.V(int32(4 * 1024)),
			InitialItemCapacity: opt.V(int32(8 * 1024)),
		}),

		spheres: mem.NewSparseList[SphereShape](64), // TODO: Configurable.
		boxes:   mem.NewSparseList[BoxShape](64),    // TODO: Configurable.
		meshes:  mem.NewSparseList[MeshShape](64),   // TODO: Configurable.
	}
}

type Scene[T any] struct {
	objects    *mem.SparseList[Object[T]]
	objectTree *spatial.DynamicOctree[mem.SparseID]

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

		sourceGroup: info.SourceGroup,
		sourceMask:  info.SourceMask.ValueOrDefault(0b1),
		targetMask:  info.TargetMask.ValueOrDefault(0b1),

		firstSphereID: mem.NilSparseID(),
		firstBoxID:    mem.NilSparseID(),
		firstMeshID:   mem.NilSparseID(),

		spatialID:      opt.Unspecified[spatial.DynamicOctreeItemID](),
		boundingSphere: Sphere{},

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

	if object.spatialID.Specified {
		s.RemoveObject(objID)
	}

	s.eachObjectSphere(object, func(sphereID mem.SparseID, _ *SphereShape) {
		s.spheres.Delete(sphereID)
	})
	// s.eachObjectCylinder(object, func(cylinderID mem.SparseID, _ *CylinderShape) {
	// 	s.cylinders.Delete(cylinderID)
	// })
	s.eachObjectBox(object, func(boxID mem.SparseID, _ *BoxShape) {
		s.boxes.Delete(boxID)
	})
	s.eachObjectMesh(object, func(meshID mem.SparseID, _ *MeshShape) {
		s.meshes.Delete(meshID)
	})
}

// InsertObject adds the object to the scene, making it visible to other
// objects.
func (s *Scene[T]) InsertObject(objID ObjectID) {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	if object.spatialID.Specified {
		panic("object already inserted")
	}

	bs := object.TransformedBoundingSphere()
	spatialID := s.objectTree.Insert(bs.Position, bs.Radius, objID.internalID)
	object.spatialID = opt.V(spatialID)
}

// RemoveObject removes the object from the scene, making it invisible to
// other objects.
func (s *Scene[T]) RemoveObject(objID ObjectID) {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	spatialID, ok := object.spatialID.Unwrap()
	if !ok {
		panic("object already removed")
	}

	s.objectTree.Remove(spatialID)
	object.spatialID = opt.Unspecified[spatial.DynamicOctreeItemID]()
}

func (s *Scene[T]) GetUserData(objID ObjectID) T {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}
	return object.userData
}

// IsValidObject returns whether an object with the specified id exists.
func (s *Scene[T]) IsValidObject(objID ObjectID) bool {
	return s.objects.Has(objID.internalID)
}

// AttachSphere creates a sphere shape and attaches it to the object
// to be used for intersection tests. The position of the sphere is
// relative to the object's origin.
func (s *Scene[T]) AttachSphere(objID ObjectID, template Sphere) SphereID {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	id, sphereShape := s.spheres.New()
	sphereShape.Init(id, template)

	s.attachObjectSphere(object, sphereShape)
	s.updateObjectBoundary(object)

	return SphereID{
		internalID: id,
	}
}

// DeleteSphere deletes a sphere shape from an object. The object is not
// deleted and continues to exist in the scene.
func (s *Scene[T]) DeleteSphere(sphereID SphereID) {
	sphere := s.spheres.Get(sphereID.internalID)
	if sphere == nil {
		panic("sphere id is invalid")
	}
	object := s.objects.Get(sphere.objectID)

	s.detachObjectSphere(object, sphere)
	s.updateObjectBoundary(object)

	s.spheres.Delete(sphereID.internalID)
}

// IsValidSphere returns whether a sphere with the specified id exists.
func (s *Scene[T]) IsValidSphere(sphereID SphereID) bool {
	return s.spheres.Has(sphereID.internalID)
}

// AttachBox creates a box shape and attaches it to the object
// to be used for intersection tests. The position and rotation of the box is
// relative to the object's transform.
func (s *Scene[T]) AttachBox(objID ObjectID, template Box) BoxID {
	object := s.objects.Get(objID.internalID)
	if object == nil {
		panic("object id is invalid")
	}

	id, boxShape := s.boxes.New()
	boxShape.Init(id, template)

	s.attachObjectBox(object, boxShape)
	s.updateObjectBoundary(object)

	return BoxID{
		internalID: id,
	}
}

// DeleteBox deletes a box shape from an object. The object is not
// deleted and continues to exist in the scene.
func (s *Scene[T]) DeleteBox(boxID BoxID) {
	box := s.boxes.Get(boxID.internalID)
	if box == nil {
		panic("box id is invalid")
	}
	object := s.objects.Get(box.objectID)

	s.detachObjectBox(object, box)
	s.updateObjectBoundary(object)

	s.boxes.Delete(boxID.internalID)
}

// IsValidBox returns whether a box with the specified id exists.
func (s *Scene[T]) IsValidBox(boxID BoxID) bool {
	return s.boxes.Has(boxID.internalID)
}

func (s *Scene[T]) CheckSphereIntersection(sphere Sphere) opt.T[ObjectIntersection] {
	checkRegion := spatial.CuboidRegion(
		sphere.Position,
		dprec.NewVec3(sphere.Radius*2.0, sphere.Radius*2.0, sphere.Radius*2.0),
	)

	var result WorstObjectIntersection
	s.objectTree.VisitHexahedronRegion(&checkRegion, spatial.VisitorFunc[mem.SparseID](func(objID mem.SparseID) {
		object := s.objects.Get(objID)
		if intersection, ok := s.checkIntersectionSphereWithObject(sphere, object).Unwrap(); ok {
			result.AddIntersection(intersection)
		}
	}))
	return result.Intersection()
}

func (s *Scene[T]) checkIntersectionSphereWithObject(sphere Sphere, object *Object[T]) opt.T[ObjectIntersection] {
	var result BestIntersection
	s.eachObjectSphere(object, func(_ mem.SparseID, shape *SphereShape) {
		// TODO: Track this inside the Shape object
		testSphere := Sphere{
			Position: object.transform.Apply(shape.template.Position),
			Radius:   shape.template.Radius,
		}
		if intersection, ok := CheckSphereSphereIntersection(sphere, testSphere); ok {
			result.AddIntersection(intersection)
		}
	})
	// TODO: Do this for boxes as well.
	// TODO: Do this for meshes as well.

	intersection, ok := result.Intersection().Unwrap()
	if !ok {
		return opt.Unspecified[ObjectIntersection]()
	}
	return opt.V(ObjectIntersection{
		FirstObjectID: NilObjectID(),
		SecondObjectID: ObjectID{
			internalID: object.id,
		},
		Intersection: intersection,
	})
}

func (s *Scene[T]) attachObjectSphere(object *Object[T], shape *SphereShape) {
	shape.objectID = object.id
	shape.nextSphereID = object.firstSphereID
	object.firstSphereID = shape.id
}

func (s *Scene[T]) detachObjectSphere(object *Object[T], shape *SphereShape) {
	shapeID := shape.id
	if object.firstSphereID == shapeID {
		object.firstSphereID = shape.nextSphereID
	} else {
		childID := object.firstSphereID
		for !childID.IsNil() {
			child := s.spheres.Get(childID)
			if child.nextSphereID == shapeID {
				child.nextSphereID = shape.nextSphereID
				break
			}
			childID = child.nextSphereID
		}
	}
}

func (s *Scene[T]) attachObjectBox(object *Object[T], shape *BoxShape) {
	shape.objectID = object.id
	shape.nextBoxID = object.firstBoxID
	object.firstBoxID = shape.id
}

func (s *Scene[T]) detachObjectBox(object *Object[T], shape *BoxShape) {
	shapeID := shape.id
	if object.firstBoxID == shapeID {
		object.firstBoxID = shape.nextBoxID
	} else {
		childID := object.firstBoxID
		for !childID.IsNil() {
			child := s.boxes.Get(childID)
			if child.nextBoxID == shapeID {
				child.nextBoxID = shape.nextBoxID
				break
			}
			childID = child.nextBoxID
		}
	}
}

func (s *Scene[T]) eachObjectSphere(object *Object[T], cb func(mem.SparseID, *SphereShape)) {
	sphereID := object.firstSphereID
	for !sphereID.IsNil() {
		shape := s.spheres.Get(sphereID)
		nextSphereID := shape.nextSphereID // track to allow deletion
		cb(sphereID, shape)
		sphereID = nextSphereID
	}
}

func (s *Scene[T]) eachObjectBox(object *Object[T], cb func(mem.SparseID, *BoxShape)) {
	boxID := object.firstBoxID
	for !boxID.IsNil() {
		box := s.boxes.Get(boxID)
		nextBoxID := box.nextBoxID // track to allow deletion
		cb(boxID, box)
		boxID = nextBoxID
	}
}

func (s *Scene[T]) eachObjectMesh(object *Object[T], cb func(mem.SparseID, *MeshShape)) {
	meshID := object.firstMeshID
	for !meshID.IsNil() {
		mesh := s.meshes.Get(meshID)
		nextMeshID := mesh.nextMeshID // track to allow deletion
		cb(meshID, mesh)
		meshID = nextMeshID
	}
}

func (s *Scene[T]) eachObjectShapeBoundingSphere(object *Object[T], cb func(Sphere)) {
	s.eachObjectSphere(object, func(_ mem.SparseID, shape *SphereShape) {
		cb(shape.BoundingSphere())
	})
	s.eachObjectBox(object, func(_ mem.SparseID, shape *BoxShape) {
		cb(shape.BoundingSphere())
	})
	s.eachObjectMesh(object, func(_ mem.SparseID, shape *MeshShape) {
		cb(shape.BoundingSphere())
	})
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
	if spatialID, ok := object.spatialID.Unwrap(); ok {
		bs := object.TransformedBoundingSphere()
		s.objectTree.Update(spatialID, bs.Position, bs.Radius)
	}
}
