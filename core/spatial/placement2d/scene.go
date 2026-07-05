package placement2d

import (
	"iter"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/query2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// SceneSettings contains information needed to create an optimal scene.
type SceneSettings struct {

	// Size specifies the dimension (from side to side) of the scene.
	// Inserting an item outside these bounds has undefined behavior.
	Size opt.T[float64]

	// MaxDepth controls the maximum depth that the underlying quadtree can reach.
	MaxDepth opt.T[uint32]

	// InitialNodeCapacity is a hint as to the number of nodes that will be
	// needed to place all items. Usually one would find that number empirically.
	// This allows the quadtree to preallocate memory and avoid dynamic allocations.
	InitialNodeCapacity opt.T[uint32]

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the quadtree. This allows the quadtree to preallocate
	// memory and avoid dynamic allocations during insertion.
	InitialItemCapacity opt.T[uint32]
}

// Scene represents a 2D scene into which dynamic objects (built from convex
// shapes) and static meshes can be placed and tested for intersection.
//
// The type parameters specify the user data attached to each kind of entity:
// O for objects, S for shapes, and M for meshes.
type Scene[O, S, M any] struct {
	shapeTree *query2d.Quadtree[int32]
	meshTree  *query2d.Quadtree[int32]

	solver *gjk2d.Solver

	freeObjectIndices *ds.Stack[int32]
	freeShapeIndices  *ds.Stack[int32]
	freeMeshIndices   *ds.Stack[int32]

	objects []sceneObject[O]
	shapes  []shape[S]
	meshes  []meshShape[M]

	shapeCandidates []int32
	meshCandidates  []int32

	tempGJKSource gjk2d.Shape
	tempGJKTarget gjk2d.Shape
}

// NewScene creates a new scene.
func NewScene[O, S, M any](settings SceneSettings) *Scene[O, S, M] {
	treeSettings := query2d.QuadtreeSettings(settings)

	return &Scene[O, S, M]{
		shapeTree: query2d.NewQuadtree[int32](treeSettings),
		meshTree:  query2d.NewQuadtree[int32](treeSettings),

		solver: gjk2d.NewSolver(),

		freeObjectIndices: ds.EmptyStack[int32](),
		freeShapeIndices:  ds.EmptyStack[int32](),
		freeMeshIndices:   ds.EmptyStack[int32](),

		objects: make([]sceneObject[O], 0),
		shapes:  make([]shape[S], 0),
		meshes:  make([]meshShape[M], 0),

		shapeCandidates: make([]int32, 0),
		meshCandidates:  make([]int32, 0),

		tempGJKSource: gjk2d.Shape{
			Points: make([]dprec.Vec2, 0, 4),
		},
		tempGJKTarget: gjk2d.Shape{
			Points: make([]dprec.Vec2, 0, 4),
		},
	}
}

// CreateObject creates a new object.
func (s *Scene[O, S, M]) CreateObject(info ObjectInfo[O]) ObjectID {
	transform := shape2d.Transform{
		Translation: info.Position.ValueOrDefault(dprec.ZeroVec2()),
		Rotation: shape2d.RotationFromAngle(
			info.Rotation.ValueOrDefault(dprec.Radians(0.0)),
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
func (s *Scene[O, S, M]) GetObjectTransform(objID ObjectID) shape2d.Transform {
	object := &s.objects[objID]
	return object.transform
}

// SetObjectTransform relocates the given object.
func (s *Scene[O, S, M]) SetObjectTransform(objID ObjectID, transform shape2d.Transform) {
	object := &s.objects[objID]
	object.transform = transform

	s.eachObjectShape(object, func(_ int32, shape *shape[S]) {
		shape.update(transform)
		bc := shape.boundingCircle()
		s.shapeTree.Update(shape.spatialID, query2d.AreaFromCircle(bc))
	})
}

// GetShapeObject returns the ID of the object that the given shape is
// attached to.
func (s *Scene[O, S, M]) GetShapeObject(shapeID ShapeID) ObjectID {
	index := int32(shapeID)
	shape := &s.shapes[index]
	return ObjectID(shape.objectIndex)
}

// AttachCircle creates a circle shape and attaches it to the object to be
// used for intersection tests.
func (s *Scene[O, S, M]) AttachCircle(objID ObjectID, info CircleInfo[S]) ShapeID {
	circle := info.Circle
	transform := shape2d.Transform{
		Translation: circle.Center,
		Rotation:    shape2d.IdentityRotation(),
	}

	return s.attachShape(int32(objID), info.Filtering, shapeRepresentation{
		lsBCircle:   circle,
		wsBCircle:   circle,
		lsTransform: transform,
		wsTransform: transform,
		kind:        shapeKindCircle,
		points: []dprec.Vec2{ // TODO: Consider reusing from a buffer.
			dprec.ZeroVec2(),
		},
		skinRadius: circle.Radius,
	}, info.UserData)
}

// AttachRectangle creates a rectangle shape and attaches it to the object to be
// used for intersection tests.
func (s *Scene[O, S, M]) AttachRectangle(objID ObjectID, info RectangleInfo[S]) ShapeID {
	rectangle := info.Rectangle
	transform := shape2d.Transform{
		Translation: rectangle.Center,
		Rotation:    rectangle.Rotation,
	}
	bCircle := rectangle.BoundingCircle()
	halfWidth := rectangle.HalfWidth
	halfHeight := rectangle.HalfHeight

	return s.attachShape(int32(objID), info.Filtering, shapeRepresentation{
		lsBCircle:   bCircle,
		wsBCircle:   bCircle,
		lsTransform: transform,
		wsTransform: transform,
		kind:        shapeKindRectangle,
		points: []dprec.Vec2{ // TODO: Consider reusing from a buffer.
			dprec.NewVec2(-halfWidth, -halfHeight),
			dprec.NewVec2(halfWidth, -halfHeight),
			dprec.NewVec2(halfWidth, halfHeight),
			dprec.NewVec2(-halfWidth, halfHeight),
		},
		skinRadius: 0.0,
	}, info.UserData)
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

// EachCircle iterates over all circle shapes in the scene that match the
// filter and yields them to the provided callback.
func (s *Scene[O, S, M]) EachCircle(filter Filter, yield func(shape2d.Circle) bool) {
	if filter.SkipDynamic {
		return
	}
	for index := range s.shapes {
		shape := &s.shapes[index]
		if shape.spatialID == query2d.InvalidTreeItemID {
			continue
		}
		if shape.kind != shapeKindCircle {
			continue
		}
		if !shape.matchesFilter(filter) {
			continue
		}
		if !yield(shape.toCircle()) {
			return
		}
	}
}

// CircleIter returns an iterator over all circle shapes in the scene that match
// the filter.
func (s *Scene[O, S, M]) CircleIter(filter Filter) iter.Seq[shape2d.Circle] {
	return func(yield func(shape2d.Circle) bool) {
		s.EachCircle(filter, yield)
	}
}

// EachRectangle iterates over all rectangle shapes in the scene that match the
// filter and yields them to the provided callback.
func (s *Scene[O, S, M]) EachRectangle(filter Filter, yield func(shape2d.Rectangle) bool) {
	if filter.SkipDynamic {
		return
	}
	for index := range s.shapes {
		shape := &s.shapes[index]
		if shape.spatialID == query2d.InvalidTreeItemID {
			continue
		}
		if shape.kind != shapeKindRectangle {
			continue
		}
		if !shape.matchesFilter(filter) {
			continue
		}
		if !yield(shape.toRectangle()) {
			return
		}
	}
}

// RectangleIter returns an iterator over all rectangle shapes in the scene that
// match the filter.
func (s *Scene[O, S, M]) RectangleIter(filter Filter) iter.Seq[shape2d.Rectangle] {
	return func(yield func(shape2d.Rectangle) bool) {
		s.EachRectangle(filter, yield)
	}
}

// CreateMesh creates a new static mesh in the scene.
//
// Unlike shapes, a mesh is not attached to an object. It is positioned
// directly through the [MeshInfo.Position] and [MeshInfo.Rotation] fields and
// is intended for static geometry that participates in intersection tests as a
// collection of edges.
func (s *Scene[O, S, M]) CreateMesh(info MeshInfo[M]) MeshID {
	transform := shape2d.Transform{
		Translation: info.Position.ValueOrDefault(dprec.ZeroVec2()),
		Rotation: shape2d.RotationFromAngle(
			info.Rotation.ValueOrDefault(dprec.Radians(0.0)),
		),
	}
	representation := newMeshRepresentation(shape2d.TransformedMesh(info.Mesh, transform))
	area := query2d.AreaFromCircle(representation.boundingCircle())

	index := s.allocateMesh()
	s.meshes[index] = meshShape[M]{
		spatialID:            s.meshTree.Insert(area, index),
		filterRepresentation: newFilterRepresentation(info.Filtering),
		meshRepresentation:   representation,
		userData:             info.UserData,
	}

	return MeshID(index)
}

// DeleteMesh removes the given mesh from the scene.
func (s *Scene[O, S, M]) DeleteMesh(meshID MeshID) {
	index := int32(meshID)
	mesh := &s.meshes[index]
	s.meshTree.Remove(mesh.spatialID)
	mesh.spatialID = query2d.InvalidTreeItemID
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
func (s *Scene[O, S, M]) CollectSegmentIntersections(segment shape2d.Segment, filter Filter, yield ContactCallback) {
	querySegment := query2d.NewSegment(segment.A, segment.B)

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
func (s *Scene[O, S, M]) CheckSegmentIntersection(segment shape2d.Segment, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectSegmentIntersections(segment, filter, collection.AddContact)
	return collection.Contact()
}

// CollectCircleIntersections collects all intersections of the circle
// with objects in the scene.
func (s *Scene[O, S, M]) CollectCircleIntersections(circle shape2d.Circle, filter Filter, yield ContactCallback) {
	queryAABB := query2d.AABBFromCircle(circle)

	if !filter.SkipDynamic {
		s.shapeCandidates = s.shapeCandidates[:0]
		s.shapeTree.QueryAABB(queryAABB, func(index int32) bool {
			s.shapeCandidates = append(s.shapeCandidates, index)
			return true
		})
		s.collectCircleShape(circle, filter, yield)
	}

	if !filter.SkipStatic {
		s.meshCandidates = s.meshCandidates[:0]
		s.meshTree.QueryAABB(queryAABB, func(index int32) bool {
			s.meshCandidates = append(s.meshCandidates, index)
			return true
		})
		s.collectCircleMesh(circle, filter, yield)
	}
}

// CheckCircleIntersection returns the deepest intersection of the circle
// with the scene.
func (s *Scene[O, S, M]) CheckCircleIntersection(circle shape2d.Circle, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectCircleIntersections(circle, filter, collection.AddContact)
	return collection.Contact()
}

// CollectRectangleIntersections collects all intersections of the rectangle
// with objects in the scene.
func (s *Scene[O, S, M]) CollectRectangleIntersections(rectangle shape2d.Rectangle, filter Filter, yield ContactCallback) {
	queryAABB := query2d.AABBFromRectangle(rectangle)

	if !filter.SkipDynamic {
		s.shapeCandidates = s.shapeCandidates[:0]
		s.shapeTree.QueryAABB(queryAABB, func(index int32) bool {
			s.shapeCandidates = append(s.shapeCandidates, index)
			return true
		})
		s.collectRectangleShape(rectangle, filter, yield)
	}

	if !filter.SkipStatic {
		s.meshCandidates = s.meshCandidates[:0]
		s.meshTree.QueryAABB(queryAABB, func(index int32) bool {
			s.meshCandidates = append(s.meshCandidates, index)
			return true
		})
		s.collectRectangleMesh(rectangle, filter, yield)
	}
}

// CheckRectangleIntersection returns the deepest intersection of the rectangle
// with the scene.
func (s *Scene[O, S, M]) CheckRectangleIntersection(rectangle shape2d.Rectangle, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectRectangleIntersections(rectangle, filter, collection.AddContact)
	return collection.Contact()
}

// CollectIntersections yields intersections found in this scene.
func (s *Scene[O, S, M]) CollectIntersections(yield ContactCallback) {
	for i := range s.shapes {
		srcIndex := int32(i)
		srcShape := &s.shapes[srcIndex]
		if srcShape.spatialID == query2d.InvalidTreeItemID {
			continue
		}

		queryAABB := query2d.AABBFromCircle(srcShape.boundingCircle())

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

func (s *Scene[O, S, M]) attachShape(
	objectIndex int32,
	filterInfo FilterInfo,
	representation shapeRepresentation,
	userData S,
) ShapeID {

	object := &s.objects[objectIndex]
	index := s.allocateShape()

	representation.update(object.transform)
	area := query2d.AreaFromCircle(representation.boundingCircle())

	s.shapes[index] = shape[S]{
		objectIndex:          objectIndex,
		nextShapeIndex:       nilIndex,
		prevShapeIndex:       object.lastShapeIndex,
		spatialID:            s.shapeTree.Insert(area, index),
		filterRepresentation: newFilterRepresentation(filterInfo),
		shapeRepresentation:  representation,
		userData:             userData,
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
	shape.spatialID = query2d.InvalidTreeItemID

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

func (s *Scene[O, S, M]) collectSegmentShape(segment shape2d.Segment, filter Filter, yield ContactCallback) {
	for index, shape := range s.iterCandidateShape(filter) {
		if !isec2d.CheckSegmentCircleOverlap(segment, shape.wsBCircle) {
			continue
		}
		onContact := func(contact shape2d.Contact) {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: ShapeID(index),
				TargetMeshID:  InvalidMeshID,
				Contact:       contact,
			})
		}
		switch shape.kind {
		case shapeKindCircle:
			circle := shape.toCircle()
			isec2d.ResolveSegmentCircle(segment, circle, onContact)
		case shapeKindRectangle:
			rectangle := shape.toRectangle()
			isec2d.ResolveSegmentRectangle(segment, rectangle, onContact)
		}
	}
}

func (s *Scene[O, S, M]) collectSegmentMesh(segment shape2d.Segment, filter Filter, yield ContactCallback) {
	for index, mesh := range s.iterCandidateMesh(filter) {
		if !isec2d.CheckSegmentCircleOverlap(segment, mesh.wsBCircle) {
			continue
		}
		var deepestContact shape2d.DeepestContact
		for _, edge := range mesh.wsEdges {
			isec2d.ResolveSegmentEdge(segment, edge, deepestContact.AddContact)
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

func (s *Scene[O, S, M]) collectCircleShape(circle shape2d.Circle, filter Filter, yield ContactCallback) {
	initGJKShapeForCircle(circle, &s.tempGJKSource)
	for index, shape := range s.iterCandidateShape(filter) {
		if !isec2d.CheckCircleCircle(circle, shape.wsBCircle) {
			continue
		}
		if contact, ok := s.solver.Resolve(s.tempGJKSource, shape.gjkShape()); ok {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: ShapeID(index),
				TargetMeshID:  InvalidMeshID,
				Contact:       contact,
			})
		}
	}
}

func (s *Scene[O, S, M]) collectCircleMesh(circle shape2d.Circle, filter Filter, yield ContactCallback) {
	initGJKShapeForCircle(circle, &s.tempGJKSource)
	for tgtIndex, tgtMesh := range s.iterCandidateMesh(filter) {
		s.resolveGJKMesh(s.tempGJKSource, circle, tgtMesh, func(contact shape2d.Contact) {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: InvalidShapeID,
				TargetMeshID:  MeshID(tgtIndex),
				Contact:       contact,
			})
		})
	}
}

func (s *Scene[O, S, M]) collectRectangleShape(rectangle shape2d.Rectangle, filter Filter, yield ContactCallback) {
	initGJKShapeForRectangle(rectangle, &s.tempGJKSource)
	for index, shape := range s.iterCandidateShape(filter) {
		if !isec2d.CheckCircleCircle(rectangle.BoundingCircle(), shape.wsBCircle) {
			continue
		}
		if contact, ok := s.solver.Resolve(s.tempGJKSource, shape.gjkShape()); ok {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: ShapeID(index),
				TargetMeshID:  InvalidMeshID,
				Contact:       contact,
			})
		}
	}
}

func (s *Scene[O, S, M]) collectRectangleMesh(rectangle shape2d.Rectangle, filter Filter, yield ContactCallback) {
	initGJKShapeForRectangle(rectangle, &s.tempGJKSource)
	for tgtIndex, tgtMesh := range s.iterCandidateMesh(filter) {
		s.resolveGJKMesh(s.tempGJKSource, rectangle.BoundingCircle(), tgtMesh, func(contact shape2d.Contact) {
			yield(Contact{
				SourceShapeID: InvalidShapeID,
				TargetShapeID: InvalidShapeID,
				TargetMeshID:  MeshID(tgtIndex),
				Contact:       contact,
			})
		})
	}
}

func (s *Scene[O, S, M]) collectShapeShape(srcIndex int32, srcShape *shape[S], yield ContactCallback) {
	srcGJKShape := srcShape.gjkShape()
	for _, tgtIndex := range s.shapeCandidates {
		tgtShape := &s.shapes[tgtIndex]
		if !shapesCanIntersect(srcShape, tgtShape) {
			continue
		}
		if !isec2d.CheckCircleCircle(srcShape.wsBCircle, tgtShape.wsBCircle) {
			continue
		}
		if contact, ok := s.solver.Resolve(srcGJKShape, tgtShape.gjkShape()); ok {
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
		s.resolveGJKMesh(srcGJKShape, srcShape.wsBCircle, tgtMesh, func(contact shape2d.Contact) {
			yield(Contact{
				SourceShapeID: ShapeID(srcIndex),
				TargetShapeID: InvalidShapeID,
				TargetMeshID:  MeshID(tgtIndex),
				Contact:       contact,
			})
		})
	}
}

func (s *Scene[O, S, M]) resolveGJKMesh(srcGJK gjk2d.Shape, srcBC shape2d.Circle, tgtMesh *meshShape[M], yield shape2d.ContactCallback) {
	if !isec2d.CheckCircleCircle(srcBC, tgtMesh.wsBCircle) {
		return
	}
	points := initGJKShapeForMesh(&s.tempGJKTarget)
	var deepestContact shape2d.DeepestContact
	for _, edge := range tgtMesh.wsEdges {
		tgtBCircle := edge.BoundingCircle()
		if !isec2d.CheckCircleCircle(srcBC, tgtBCircle) {
			continue
		}
		points[0] = edge.A
		points[1] = edge.B
		if contact, ok := s.solver.Resolve(srcGJK, s.tempGJKTarget); ok {
			// Prevent contacts that try to push the source shape into the edge.
			if dprec.Vec2Dot(contact.TargetNormal, edge.Normal()) > 0 {
				deepestContact.AddContact(contact)
			}
		}
	}
	if contact, ok := deepestContact.Contact(); ok {
		yield(contact)
	}
}

func (s *Scene[O, S, M]) eachCandidateShape(filter Filter, cb func(int32, *shape[S]) bool) {
	for _, index := range s.shapeCandidates {
		shape := &s.shapes[index]
		if !shape.matchesFilter(filter) {
			continue
		}
		if !cb(index, shape) {
			return
		}
	}
}

func (s *Scene[O, S, M]) iterCandidateShape(filter Filter) iter.Seq2[int32, *shape[S]] {
	return func(yield func(int32, *shape[S]) bool) {
		s.eachCandidateShape(filter, yield)
	}
}

func (s *Scene[O, S, M]) eachCandidateMesh(filter Filter, cb func(int32, *meshShape[M]) bool) {
	for _, index := range s.meshCandidates {
		mesh := &s.meshes[index]
		if !mesh.matchesFilter(filter) {
			continue
		}
		if !cb(index, mesh) {
			return
		}
	}
}

func (s *Scene[O, S, M]) iterCandidateMesh(filter Filter) iter.Seq2[int32, *meshShape[M]] {
	return func(yield func(int32, *meshShape[M]) bool) {
		s.eachCandidateMesh(filter, yield)
	}
}

func initGJKShapeForCircle(circle shape2d.Circle, out *gjk2d.Shape) {
	out.Position = circle.Center
	out.Rotation = shape2d.IdentityRotation()
	out.Points = out.Points[:1]
	out.Points[0] = dprec.ZeroVec2()
	out.SkinRadius = circle.Radius
}

func initGJKShapeForRectangle(rectangle shape2d.Rectangle, out *gjk2d.Shape) {
	out.Position = rectangle.Center
	out.Rotation = rectangle.Rotation
	out.Points = out.Points[:4]
	halfWidth := rectangle.HalfWidth
	halfHeight := rectangle.HalfHeight
	out.Points[0] = dprec.NewVec2(-halfWidth, -halfHeight)
	out.Points[1] = dprec.NewVec2(halfWidth, -halfHeight)
	out.Points[2] = dprec.NewVec2(halfWidth, halfHeight)
	out.Points[3] = dprec.NewVec2(-halfWidth, halfHeight)
	out.SkinRadius = 0.0
}

func initGJKShapeForMesh(out *gjk2d.Shape) []dprec.Vec2 {
	out.Position = dprec.ZeroVec2()
	out.Rotation = shape2d.IdentityRotation()
	out.Points = out.Points[:2]
	out.SkinRadius = 0.0
	return out.Points
}
