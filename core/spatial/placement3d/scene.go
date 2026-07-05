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

// Scene represents a 3D scene where objects made of shapes can be added.
type Scene[O, S, M any] struct {
	shapeTree *query3d.Octree[int32]
	meshTree  *query3d.Octree[int32]

	solver *gjk3d.Solver

	freeObjectIndices *ds.Stack[int32]
	freeShapeIndices  *ds.Stack[int32]
	freeMeshIndices   *ds.Stack[int32]

	objects []sceneObject[O]
	shapes  []shape[S]
	meshes  []meshShape[M]

	shapeCandidates []int32
	meshCandidates  []int32

	tempGJKSource gjk3d.Shape
	tempGJKTarget gjk3d.Shape
}

// NewScene creates a new scene.
func NewScene[O, S, M any](settings SceneSettings) *Scene[O, S, M] {
	treeSettings := query3d.OctreeSettings(settings)

	return &Scene[O, S, M]{
		shapeTree: query3d.NewOctree[int32](treeSettings),
		meshTree:  query3d.NewOctree[int32](treeSettings),

		solver: gjk3d.NewSolver(),

		freeObjectIndices: ds.EmptyStack[int32](),
		freeShapeIndices:  ds.EmptyStack[int32](),
		freeMeshIndices:   ds.EmptyStack[int32](),

		objects: make([]sceneObject[O], 0),
		shapes:  make([]shape[S], 0),
		meshes:  make([]meshShape[M], 0),

		shapeCandidates: make([]int32, 0),
		meshCandidates:  make([]int32, 0),

		tempGJKSource: gjk3d.Shape{
			Points: make([]dprec.Vec3, 0, 8),
		},
		tempGJKTarget: gjk3d.Shape{
			Points: make([]dprec.Vec3, 0, 8),
		},
	}
}

// CreateObject creates a new object.
func (s *Scene[O, S, M]) CreateObject(info ObjectInfo[O]) ObjectID {
	transform := shape3d.Transform{
		Translation: info.Position.ValueOrDefault(dprec.ZeroVec3()),
		Rotation: shape3d.RotationFromQuat(
			info.Rotation.ValueOrDefault(dprec.IdentityQuat()),
		),
	}

	index := s.allocateObject()
	s.objects[index] = sceneObject[O]{
		transform:       transform,
		firstShapeIndex: nilIndex,
		lastShapeIndex:  nilIndex,
		userData:        info.UserData,
	}
	return ObjectID(index)
}

// DeleteObject deletes an object.
func (s *Scene[O, S, M]) DeleteObject(objID ObjectID) {
	index := int32(objID)
	object := &s.objects[index]
	object.userData = gog.Zero[O]() // in case of pointer
	s.eachObjectShape(object, func(shapeIndex int32, _ *shape[S]) {
		s.detachShape(shapeIndex)
	})
	s.releaseObject(index)
}

// GetObjectUserData returns the user data associated with the given object.
func (s *Scene[O, S, M]) GetObjectUserData(objID ObjectID) O {
	object := &s.objects[objID]
	return object.userData
}

// SetObjectUserData assigns the specified user data to the object.
func (s *Scene[O, S, M]) SetObjectUserData(objID ObjectID, userData O) {
	object := &s.objects[objID]
	object.userData = userData
}

// GetObjectTransform returns the given object's transform.
func (s *Scene[O, S, M]) GetObjectTransform(objID ObjectID) shape3d.Transform {
	object := &s.objects[objID]
	return object.transform
}

// SetObjectTransform relocates the given object.
func (s *Scene[O, S, M]) SetObjectTransform(objID ObjectID, transform shape3d.Transform) {
	object := &s.objects[objID]
	object.transform = transform

	s.eachObjectShape(object, func(_ int32, shape *shape[S]) {
		shape.update(transform)
		bs := shape.boundingSphere()
		s.shapeTree.Update(shape.spatialID, query3d.AreaFromSphere(bs))
	})
}

// GetShapeObject returns the ID of the object that the given shape is
// attached to.
func (s *Scene[O, S, M]) GetShapeObject(shapeID ShapeID) ObjectID {
	index := int32(shapeID)
	shape := &s.shapes[index]
	return ObjectID(shape.objectIndex)
}

// AttachSphere creates a sphere shape and attaches it to the object to be
// used for intersection tests.
func (s *Scene[O, S, M]) AttachSphere(objID ObjectID, info SphereInfo[S]) ShapeID {
	sphere := info.Sphere
	transform := shape3d.Transform{
		Translation: sphere.Center,
		Rotation:    shape3d.IdentityRotation(),
	}

	return s.attachShape(int32(objID), info.ShapeInfo, shapeRepresentation{
		lsBSphere:   sphere,
		wsBSphere:   sphere,
		lsTransform: transform,
		wsTransform: transform,
		kind:        shapeKindSphere,
		points: []dprec.Vec3{ // TODO: Consider reusing from a buffer.
			dprec.ZeroVec3(),
		},
		skinRadius: sphere.Radius,
	})
}

// AttachBox creates a box shape and attaches it to the object to be used for
// intersection tests.
func (s *Scene[O, S, M]) AttachBox(objID ObjectID, info BoxInfo[S]) ShapeID {
	box := info.Box
	transform := shape3d.Transform{
		Translation: info.Box.Center,
		Rotation:    info.Box.Rotation,
	}
	bSphere := box.BoundingSphere()
	halfWidth := box.HalfWidth
	halfHeight := box.HalfHeight
	halfLength := box.HalfLength

	return s.attachShape(int32(objID), info.ShapeInfo, shapeRepresentation{
		lsBSphere:   bSphere,
		wsBSphere:   bSphere,
		lsTransform: transform,
		wsTransform: transform,
		kind:        shapeKindBox,
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
	})
}

// DeleteShape deletes a shape from an object. The object is not
// deleted and continues to exist in the scene.
func (s *Scene[O, S, M]) DeleteShape(shapeID ShapeID) {
	index := int32(shapeID)
	s.detachShape(index)
}

// GetShapeUserData returns the user data associated with the given shape.
func (s *Scene[O, S, M]) GetShapeUserData(shapeID ShapeID) S {
	index := int32(shapeID)
	shape := &s.shapes[index]
	return shape.userData
}

// SetShapeUserData assigns the specified user data to the shape.
func (s *Scene[O, S, M]) SetShapeUserData(shapeID ShapeID, userData S) {
	index := int32(shapeID)
	shape := &s.shapes[index]
	shape.userData = userData
}

// EachSphere iterates over all sphere shapes in the scene that match the
// filter and yields them to the provided callback.
func (s *Scene[O, S, M]) EachSphere(filter Filter, yield func(shape3d.Sphere) bool) {
	for index := range s.shapes {
		shape := &s.shapes[index]
		if shape.spatialID == query3d.InvalidTreeItemID {
			continue
		}
		if shape.kind != shapeKindSphere {
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
func (s *Scene[O, S, M]) SphereIter(filter Filter) iter.Seq[shape3d.Sphere] {
	return func(yield func(shape3d.Sphere) bool) {
		s.EachSphere(filter, yield)
	}
}

// EachBox iterates over all box shapes in the scene that match the
// filter and yields them to the provided callback.
func (s *Scene[O, S, M]) EachBox(filter Filter, yield func(shape3d.Box) bool) {
	for index := range s.shapes {
		shape := &s.shapes[index]
		if shape.spatialID == query3d.InvalidTreeItemID {
			continue
		}
		if shape.kind != shapeKindBox {
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
func (s *Scene[O, S, M]) BoxIter(filter Filter) iter.Seq[shape3d.Box] {
	return func(yield func(shape3d.Box) bool) {
		s.EachBox(filter, yield)
	}
}

// CreateMesh creates a new static mesh in the scene.
//
// Unlike shapes, a mesh is not attached to an object. It is positioned
// directly through the [MeshInfo.Position] and [MeshInfo.Rotation] fields and
// is intended for static geometry that participates in intersection tests as a
// collection of triangles.
func (s *Scene[O, S, M]) CreateMesh(info MeshInfo[M]) MeshID {
	transform := shape3d.Transform{
		Translation: info.Position.ValueOrDefault(dprec.ZeroVec3()),
		Rotation: shape3d.RotationFromQuat(
			info.Rotation.ValueOrDefault(dprec.IdentityQuat()),
		),
	}
	representation := newMeshRepresentation(shape3d.TransformedMesh(info.Mesh, transform))
	area := query3d.AreaFromSphere(representation.boundingSphere())

	index := s.allocateMesh()
	s.meshes[index] = meshShape[M]{
		spatialID: s.meshTree.Insert(area, index),
		filterRepresentation: filterRepresentation{
			rejectGroup: info.RejectGroup,
			sourceMask:  info.SourceMask.ValueOrDefault(0b1),
			targetMask:  info.TargetMask.ValueOrDefault(0b1),
		},
		meshRepresentation: representation,
		userData:           info.UserData,
	}

	return MeshID(index)
}

// DeleteMesh removes the given mesh from the scene.
func (s *Scene[O, S, M]) DeleteMesh(meshID MeshID) {
	index := int32(meshID)
	mesh := &s.meshes[index]
	s.meshTree.Remove(mesh.spatialID)
	mesh.spatialID = query3d.InvalidTreeItemID
	mesh.userData = gog.Zero[M]() // in case of pointer
	s.releaseMesh(index)
}

// GetMeshUserData returns the user data associated with the given mesh.
func (s *Scene[O, S, M]) GetMeshUserData(meshID MeshID) M {
	mesh := &s.meshes[meshID]
	return mesh.userData
}

// SetMeshUserData assigns the specified user data to the mesh.
func (s *Scene[O, S, M]) SetMeshUserData(meshID MeshID, userData M) {
	mesh := &s.meshes[meshID]
	mesh.userData = userData
}

// CollectSegmentIntersections collects all intersections of the segment
// with objects in the scene.
func (s *Scene[O, S, M]) CollectSegmentIntersections(segment shape3d.Segment, filter Filter, yield ContactCallback) {
	querySegment := query3d.NewSegment(segment.A, segment.B)

	if !filter.SkipDynamic {
		s.shapeCandidates = s.shapeCandidates[:0]
		s.shapeTree.QuerySegment(querySegment, func(index int32) bool {
			s.shapeCandidates = append(s.shapeCandidates, index)
			return true
		})
		s.collectSegmentShape(segment, filter, yield)
	}

	if !filter.SkipStatic {
		s.meshCandidates = s.meshCandidates[:0]
		s.meshTree.QuerySegment(querySegment, func(index int32) bool {
			s.meshCandidates = append(s.meshCandidates, index)
			return true
		})
		s.collectSegmentMesh(segment, filter, yield)
	}
}

// CheckSegmentIntersection returns the deepest intersection of the segment
// with the scene.
func (s *Scene[O, S, M]) CheckSegmentIntersection(segment shape3d.Segment, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectSegmentIntersections(segment, filter, collection.AddContact)
	return collection.Contact()
}

// CollectSphereIntersections collects all intersections of the sphere
// with objects in the scene.
func (s *Scene[O, S, M]) CollectSphereIntersections(sphere shape3d.Sphere, filter Filter, yield ContactCallback) {
	queryAABB := query3d.AABBFromSphere(sphere)

	if !filter.SkipDynamic {
		s.shapeCandidates = s.shapeCandidates[:0]
		s.shapeTree.QueryAABB(queryAABB, func(index int32) bool {
			s.shapeCandidates = append(s.shapeCandidates, index)
			return true
		})
		s.collectSphereShape(sphere, filter, yield)
	}

	if !filter.SkipStatic {
		s.meshCandidates = s.meshCandidates[:0]
		s.meshTree.QueryAABB(queryAABB, func(index int32) bool {
			s.meshCandidates = append(s.meshCandidates, index)
			return true
		})
		s.collectSphereMesh(sphere, filter, yield)
	}
}

// CheckSphereIntersection returns the deepest intersection of the sphere
// with the scene.
func (s *Scene[O, S, M]) CheckSphereIntersection(sphere shape3d.Sphere, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectSphereIntersections(sphere, filter, collection.AddContact)
	return collection.Contact()
}

// CollectBoxIntersections collects all intersections of the box
// with objects in the scene.
func (s *Scene[O, S, M]) CollectBoxIntersections(box shape3d.Box, filter Filter, yield ContactCallback) {
	queryAABB := query3d.AABBFromBox(box)

	if !filter.SkipDynamic {
		s.shapeCandidates = s.shapeCandidates[:0]
		s.shapeTree.QueryAABB(queryAABB, func(index int32) bool {
			s.shapeCandidates = append(s.shapeCandidates, index)
			return true
		})
		s.collectBoxShape(box, filter, yield)
	}

	if !filter.SkipStatic {
		s.meshCandidates = s.meshCandidates[:0]
		s.meshTree.QueryAABB(queryAABB, func(index int32) bool {
			s.meshCandidates = append(s.meshCandidates, index)
			return true
		})
		s.collectBoxMesh(box, filter, yield)
	}
}

// CheckBoxIntersection returns the deepest intersection of the box
// with the scene.
func (s *Scene[O, S, M]) CheckBoxIntersection(box shape3d.Box, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectBoxIntersections(box, filter, collection.AddContact)
	return collection.Contact()
}

// CollectIntersections yields intersections found in this scene.
func (s *Scene[O, S, M]) CollectIntersections(yield ContactCallback) {
	for i := range s.shapes {
		srcIndex := int32(i)
		srcShape := &s.shapes[srcIndex]
		if srcShape.spatialID == query3d.InvalidTreeItemID {
			continue
		}

		queryAABB := query3d.AABBFromSphere(srcShape.boundingSphere())

		s.shapeCandidates = s.shapeCandidates[:0]
		s.shapeTree.QueryAABB(queryAABB, func(tgtIndex int32) bool {
			s.shapeCandidates = append(s.shapeCandidates, tgtIndex)
			return true
		})
		s.collectShapeShape(srcIndex, srcShape, yield)

		s.meshCandidates = s.meshCandidates[:0]
		s.meshTree.QueryAABB(queryAABB, func(tgtIndex int32) bool {
			s.meshCandidates = append(s.meshCandidates, tgtIndex)
			return true
		})
		s.collectShapeMesh(srcIndex, srcShape, yield)
	}
}

const nilIndex = -1

func (s *Scene[O, S, M]) allocateObject() int32 {
	if s.freeObjectIndices.IsEmpty() {
		index := len(s.objects)
		s.objects = append(s.objects, sceneObject[O]{})
		return int32(index)
	} else {
		return s.freeObjectIndices.Pop()
	}
}

func (s *Scene[O, S, M]) releaseObject(index int32) {
	s.freeObjectIndices.Push(index)
}

func (s *Scene[O, S, M]) eachObjectShape(object *sceneObject[O], cb func(int32, *shape[S])) {
	index := object.firstShapeIndex
	for index >= 0 {
		shape := &s.shapes[index]
		nextIndex := shape.nextShapeIndex
		cb(index, shape)
		index = nextIndex
	}
}

func (s *Scene[O, S, M]) allocateShape() int32 {
	if s.freeShapeIndices.IsEmpty() {
		index := len(s.shapes)
		s.shapes = append(s.shapes, shape[S]{})
		return int32(index)
	} else {
		return s.freeShapeIndices.Pop()
	}
}

func (s *Scene[O, S, M]) releaseShape(index int32) {
	s.freeShapeIndices.Push(index)
}

func (s *Scene[O, S, M]) attachShape(objectIndex int32, info ShapeInfo[S], representation shapeRepresentation) ShapeID {
	object := &s.objects[objectIndex]

	index := s.allocateShape()

	representation.update(object.transform)
	area := query3d.AreaFromSphere(representation.boundingSphere())

	s.shapes[index] = shape[S]{
		objectIndex:    objectIndex,
		nextShapeIndex: nilIndex,
		prevShapeIndex: object.lastShapeIndex,
		spatialID:      s.shapeTree.Insert(area, index),
		filterRepresentation: filterRepresentation{
			rejectGroup: info.RejectGroup,
			sourceMask:  info.SourceMask.ValueOrDefault(0b1),
			targetMask:  info.TargetMask.ValueOrDefault(0b1),
		},
		shapeRepresentation: representation,
		userData:            info.UserData,
	}
	if object.firstShapeIndex == nilIndex {
		object.firstShapeIndex = index
	} else {
		s.shapes[object.lastShapeIndex].nextShapeIndex = index
	}
	object.lastShapeIndex = index

	return ShapeID(index)
}

func (s *Scene[O, S, M]) detachShape(index int32) {
	shape := &s.shapes[index]

	s.shapeTree.Remove(shape.spatialID)
	shape.spatialID = query3d.InvalidTreeItemID

	object := &s.objects[shape.objectIndex]
	if object.firstShapeIndex == index {
		object.firstShapeIndex = shape.nextShapeIndex
	}
	if object.lastShapeIndex == index {
		object.lastShapeIndex = shape.prevShapeIndex
	}
	if shape.prevShapeIndex != nilIndex {
		prevShape := &s.shapes[shape.prevShapeIndex]
		prevShape.nextShapeIndex = shape.nextShapeIndex
	}
	if shape.nextShapeIndex != nilIndex {
		nextShape := &s.shapes[shape.nextShapeIndex]
		nextShape.prevShapeIndex = shape.prevShapeIndex
	}
	shape.objectIndex = -1
	shape.userData = gog.Zero[S]() // in case of pointer

	s.releaseShape(index)
}

func (s *Scene[O, S, M]) allocateMesh() int32 {
	if s.freeMeshIndices.IsEmpty() {
		index := len(s.meshes)
		s.meshes = append(s.meshes, meshShape[M]{})
		return int32(index)
	} else {
		return s.freeMeshIndices.Pop()
	}
}

func (s *Scene[O, S, M]) releaseMesh(index int32) {
	s.freeMeshIndices.Push(index)
}

func (s *Scene[O, S, M]) collectSegmentShape(segment shape3d.Segment, filter Filter, yield ContactCallback) {
	for _, index := range s.shapeCandidates {
		shape := &s.shapes[index]
		if !shape.matchesFilter(filter) {
			continue
		}
		if !isec3d.CheckSegmentSphereOverlap(segment, shape.wsBSphere) {
			continue
		}
		onContact := func(contact shape3d.Contact) {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: ShapeID(index),
				TargetMeshID:  InvalidMeshID,
				Contact:       contact,
			})
		}
		switch shape.kind {
		case shapeKindSphere:
			sphere := shape.toSphere()
			isec3d.ResolveSegmentSphere(segment, sphere, onContact)
		case shapeKindBox:
			box := shape.toBox()
			isec3d.ResolveSegmentBox(segment, box, onContact)
		}
	}
}

func (s *Scene[O, S, M]) collectSegmentMesh(segment shape3d.Segment, filter Filter, yield ContactCallback) {
	for _, index := range s.meshCandidates {
		mesh := &s.meshes[index]
		if !mesh.matchesFilter(filter) {
			continue
		}
		if !isec3d.CheckSegmentSphereOverlap(segment, mesh.wsBSphere) {
			continue
		}
		var deepestContact shape3d.DeepestContact
		for _, triangle := range mesh.wsTriangles {
			isec3d.ResolveSegmentTriangle(segment, triangle, deepestContact.AddContact)
		}
		if contact, ok := deepestContact.Contact(); ok {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: InvalidShapeID,
				TargetMeshID:  MeshID(index),
				Contact:       contact,
			})
		}
	}
}

func (s *Scene[O, S, M]) collectSphereShape(sphere shape3d.Sphere, filter Filter, yield ContactCallback) {
	initGJKShapeForSphere(sphere, &s.tempGJKSource)

	for _, index := range s.shapeCandidates {
		shape := &s.shapes[index]
		if !shape.matchesFilter(filter) {
			continue
		}
		if !isec3d.CheckSphereSphere(sphere, shape.wsBSphere) {
			continue
		}
		tgtGJKShape := shape.gjkShape()
		if contact, ok := s.solver.Resolve(s.tempGJKSource, tgtGJKShape); ok {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: ShapeID(index),
				TargetMeshID:  InvalidMeshID,
				Contact:       contact,
			})
		}
	}
}

func (s *Scene[O, S, M]) collectSphereMesh(sphere shape3d.Sphere, filter Filter, yield ContactCallback) {
	initGJKShapeForSphere(sphere, &s.tempGJKSource)

	for _, tgtIndex := range s.meshCandidates {
		tgtMesh := &s.meshes[tgtIndex]
		if !tgtMesh.matchesFilter(filter) {
			continue
		}
		if !isec3d.CheckSphereSphere(sphere, tgtMesh.wsBSphere) {
			continue
		}

		points := initGJKShapeForMesh(&s.tempGJKTarget)

		var deepestContact shape3d.DeepestContact
		for _, triangle := range tgtMesh.wsTriangles {
			tgtBSphere := triangle.BoundingSphere()
			if !isec3d.CheckSphereSphere(sphere, tgtBSphere) {
				continue
			}
			points[0] = triangle.A
			points[1] = triangle.B
			points[2] = triangle.C
			if contact, ok := s.solver.Resolve(s.tempGJKSource, s.tempGJKTarget); ok {
				// Prevent contacts that try to push the source shape into the triangle.
				if dprec.Vec3Dot(contact.TargetNormal, triangle.Normal()) > 0 {
					deepestContact.AddContact(contact)
				}
			}
		}
		if contact, ok := deepestContact.Contact(); ok {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: InvalidShapeID,
				TargetMeshID:  MeshID(tgtIndex),
				Contact:       contact,
			})
		}
	}
}

func (s *Scene[O, S, M]) collectBoxShape(box shape3d.Box, filter Filter, yield ContactCallback) {
	initGJKShapeForBox(box, &s.tempGJKSource)

	for _, index := range s.shapeCandidates {
		shape := &s.shapes[index]
		if !shape.matchesFilter(filter) {
			continue
		}
		if !isec3d.CheckSphereSphere(box.BoundingSphere(), shape.wsBSphere) {
			continue
		}
		tgtGJKShape := shape.gjkShape()
		if contact, ok := s.solver.Resolve(s.tempGJKSource, tgtGJKShape); ok {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: ShapeID(index),
				TargetMeshID:  InvalidMeshID,
				Contact:       contact,
			})
		}
	}
}

func (s *Scene[O, S, M]) collectBoxMesh(box shape3d.Box, filter Filter, yield ContactCallback) {
	initGJKShapeForBox(box, &s.tempGJKSource)

	for _, tgtIndex := range s.meshCandidates {
		tgtMesh := &s.meshes[tgtIndex]
		if !tgtMesh.matchesFilter(filter) {
			continue
		}
		if !isec3d.CheckSphereSphere(box.BoundingSphere(), tgtMesh.wsBSphere) {
			continue
		}

		points := initGJKShapeForMesh(&s.tempGJKTarget)

		var deepestContact shape3d.DeepestContact
		for _, triangle := range tgtMesh.wsTriangles {
			tgtBSphere := triangle.BoundingSphere()
			if !isec3d.CheckSphereSphere(box.BoundingSphere(), tgtBSphere) {
				continue
			}
			points[0] = triangle.A
			points[1] = triangle.B
			points[2] = triangle.C
			if contact, ok := s.solver.Resolve(s.tempGJKSource, s.tempGJKTarget); ok {
				// Prevent contacts that try to push the source shape into the triangle.
				if dprec.Vec3Dot(contact.TargetNormal, triangle.Normal()) > 0 {
					deepestContact.AddContact(contact)
				}
			}
		}
		if contact, ok := deepestContact.Contact(); ok {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: InvalidShapeID,
				TargetMeshID:  MeshID(tgtIndex),
				Contact:       contact,
			})
		}
	}
}

func (s *Scene[O, S, M]) collectShapeShape(srcIndex int32, srcShape *shape[S], yield ContactCallback) {
	srcGJKShape := srcShape.gjkShape()
	for _, tgtIndex := range s.shapeCandidates {
		tgtShape := &s.shapes[tgtIndex]
		if !shapesCanIntersect(srcShape, tgtShape) {
			continue
		}
		if !isec3d.CheckSphereSphere(srcShape.wsBSphere, tgtShape.wsBSphere) {
			continue
		}
		tgtGJKShape := tgtShape.gjkShape()
		if contact, ok := s.solver.Resolve(srcGJKShape, tgtGJKShape); ok {
			yield(Contact{
				SourceShapeID: ShapeID(srcIndex),
				TargetShapeID: ShapeID(tgtIndex),
				TargetMeshID:  InvalidMeshID,
				Contact:       contact,
			})
		}
	}
}

func (s *Scene[O, S, M]) collectShapeMesh(srcIndex int32, srcShape *shape[S], yield ContactCallback) {
	srcGJKShape := srcShape.gjkShape()
	for _, tgtIndex := range s.meshCandidates {
		tgtMesh := &s.meshes[tgtIndex]
		if !shapeMeshCanIntersect(srcShape, tgtMesh) {
			continue
		}
		if !isec3d.CheckSphereSphere(srcShape.wsBSphere, tgtMesh.wsBSphere) {
			continue
		}

		points := initGJKShapeForMesh(&s.tempGJKTarget)

		var deepestContact shape3d.DeepestContact
		for _, triangle := range tgtMesh.wsTriangles {
			tgtBSphere := triangle.BoundingSphere()
			if !isec3d.CheckSphereSphere(srcShape.wsBSphere, tgtBSphere) {
				continue
			}
			points[0] = triangle.A
			points[1] = triangle.B
			points[2] = triangle.C
			if contact, ok := s.solver.Resolve(srcGJKShape, s.tempGJKTarget); ok {
				// Prevent contacts that try to push the source shape into the triangle.
				if dprec.Vec3Dot(contact.TargetNormal, triangle.Normal()) > 0 {
					deepestContact.AddContact(contact)
				}
			}
		}
		if contact, ok := deepestContact.Contact(); ok {
			yield(Contact{
				SourceShapeID: ShapeID(srcIndex),
				TargetShapeID: InvalidShapeID,
				TargetMeshID:  MeshID(tgtIndex),
				Contact:       contact,
			})
		}
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
