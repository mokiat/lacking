package placement2d

import (
	"iter"
	"slices"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/query2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// SceneSettings contains information needed to create an optimal scene.
type SceneSettings struct {

	// Size specifies the dimension (from side to side) of the scene.
	// Inserting an object outside these bounds has undefined behavior.
	//
	// If not specified, a default size of 4096 units is used.
	Size opt.T[float64]

	// MaxDepth specifies the maximum depth of the internal spatial
	// partitioning structure.
	//
	// If not specified, a default max depth of 8 is used.
	MaxDepth opt.T[uint32]

	// InitialNodeCapacity is a hint as to the number of nodes that will be
	// needed to place all items. Usually one would find that number empirically.
	// This allows the quadtree to preallocate memory and avoid dynamic allocations.
	//
	// By default the initial capacity is 4096.
	InitialNodeCapacity opt.T[uint32]

	// InitialItemCapacity is a hint as to the likely upper bound of items that
	// will be inserted into the quadtree. This allows the quadtree to preallocate
	// memory and avoid dynamic allocations during insertion.
	//
	// By default the initial capacity is 1024.
	InitialItemCapacity opt.T[uint32]
}

// Scene represents a 2D scene where objects made of shapes can be added.
type Scene[T, S any] struct {
	freeObjectIndices    *ds.Stack[uint32]
	freeCircleIndices    *ds.Stack[uint32]
	freeRectangleIndices *ds.Stack[uint32]
	freeMeshIndices      *ds.Stack[uint32]

	objects    []sceneObject[T]
	circles    []sceneCircleShape[S]
	rectangles []sceneRectangleShape[S]
	meshes     []sceneMeshShape[S]

	staticTree  *query2d.Quadtree[shapeRef]
	dynamicTree *query2d.Quadtree[shapeRef]

	tempShape     sceneShape[S]
	tempSegment   shape2d.Segment
	tempCircle    circleSolver
	tempRectangle rectangleSolver
	tempMesh      meshSolver

	checks []shapeRefPair
}

// NewScene creates a new scene.
func NewScene[O, S any](settings SceneSettings) *Scene[O, S] {
	treeSettings := query2d.QuadtreeSettings(settings)

	return &Scene[O, S]{
		freeObjectIndices:    ds.EmptyStack[uint32](),
		freeCircleIndices:    ds.EmptyStack[uint32](),
		freeRectangleIndices: ds.EmptyStack[uint32](),
		freeMeshIndices:      ds.EmptyStack[uint32](),

		objects:    make([]sceneObject[O], 0),
		circles:    make([]sceneCircleShape[S], 0),
		rectangles: make([]sceneRectangleShape[S], 0),
		meshes:     make([]sceneMeshShape[S], 0),

		staticTree:  query2d.NewQuadtree[shapeRef](treeSettings),
		dynamicTree: query2d.NewQuadtree[shapeRef](treeSettings),

		checks: make([]shapeRefPair, 0, 1024),
	}
}

// CreateObject creates a new object.
func (s *Scene[O, S]) CreateObject(info ObjectInfo[O]) ObjectID {
	transform := shape2d.TRTransform(
		info.Position.ValueOrDefault(dprec.ZeroVec2()),
		shape2d.RotationFromAngle(
			info.Rotation.ValueOrDefault(dprec.Radians(0.0)),
		),
	)

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
func (s *Scene[O, S]) GetObjectTransform(objID ObjectID) shape2d.Transform {
	object := &s.objects[objID]
	return object.transform
}

// SetObjectTransform relocates the given object.
func (s *Scene[O, S]) SetObjectTransform(objID ObjectID, transform shape2d.Transform) {
	object := &s.objects[objID]
	object.transform = transform
	s.eachObjectShape(object, shapeKindCircle, func(index uint32) {
		circle := &s.circles[index]
		circle.update(transform)
		bc := query2d.AreaFromCircle(circle.boundingCircle())
		if circle.static {
			s.staticTree.Update(circle.spatialID, bc)
		} else {
			s.dynamicTree.Update(circle.spatialID, bc)
		}
	})
	s.eachObjectShape(object, shapeKindRectangle, func(index uint32) {
		rectangle := &s.rectangles[index]
		rectangle.update(transform)
		bc := query2d.AreaFromCircle(rectangle.boundingCircle())
		if rectangle.static {
			s.staticTree.Update(rectangle.spatialID, bc)
		} else {
			s.dynamicTree.Update(rectangle.spatialID, bc)
		}
	})
	s.eachObjectShape(object, shapeKindMesh, func(index uint32) {
		mesh := &s.meshes[index]
		mesh.update(transform)
		bc := query2d.AreaFromCircle(mesh.boundingCircle())
		if mesh.static {
			s.staticTree.Update(mesh.spatialID, bc)
		} else {
			s.dynamicTree.Update(mesh.spatialID, bc)
		}
	})
}

// AttachCircle creates a circle shape and attaches it to the object to be
// used for intersection tests.
func (s *Scene[O, S]) AttachCircle(objID ObjectID, info CircleInfo[S]) ShapeID {
	var index uint32
	if s.freeCircleIndices.IsEmpty() {
		index = uint32(len(s.circles))
		s.circles = append(s.circles, sceneCircleShape[S]{})
	} else {
		index = s.freeCircleIndices.Pop()
	}
	ref := newShapeRef(shapeKindCircle, index)

	object := &s.objects[objID]

	solver := newCircleSolver(info.Circle)
	solver.update(object.transform)

	bc := solver.boundingCircle()
	var spatialID query2d.TreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(query2d.AreaFromCircle(bc), ref)
	} else {
		spatialID = s.dynamicTree.Insert(query2d.AreaFromCircle(bc), ref)
	}

	s.circles[index] = sceneCircleShape[S]{
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
		circleSolver: solver,
	}
	object.firstShape = ref

	return ShapeID(ref)
}

// AttachRectangle creates a rectangle shape and attaches it to the object to be
// used for intersection tests.
func (s *Scene[O, S]) AttachRectangle(objID ObjectID, info RectangleInfo[S]) ShapeID {
	var index uint32
	if s.freeRectangleIndices.IsEmpty() {
		index = uint32(len(s.rectangles))
		s.rectangles = append(s.rectangles, sceneRectangleShape[S]{})
	} else {
		index = s.freeRectangleIndices.Pop()
	}
	ref := newShapeRef(shapeKindRectangle, index)

	object := &s.objects[objID]

	solver := newRectangleSolver(info.Rectangle)
	solver.update(object.transform)

	bc := solver.boundingCircle()
	var spatialID query2d.TreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(query2d.AreaFromCircle(bc), ref)
	} else {
		spatialID = s.dynamicTree.Insert(query2d.AreaFromCircle(bc), ref)
	}

	s.rectangles[index] = sceneRectangleShape[S]{
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
		rectangleSolver: solver,
	}
	object.firstShape = ref

	return ShapeID(ref)
}

// AttachMesh creates a mesh shape and attaches it to the object to be
// used for intersection tests.
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

	bc := solver.boundingCircle()
	var spatialID query2d.TreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(query2d.AreaFromCircle(bc), ref)
	} else {
		spatialID = s.dynamicTree.Insert(query2d.AreaFromCircle(bc), ref)
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

// EachCircle iterates over all circle shapes in the scene that match the given
// filter and yields them to the provided callback.
func (s *Scene[O, S]) EachCircle(filter Filter, yield func(shape2d.Circle) bool) {
	for index := range uint32(len(s.circles)) {
		shape := &s.circles[index]
		if shape.spatialID == query2d.InvalidTreeItemID {
			continue
		}
		if !shape.matchesFilter(filter) {
			continue
		}
		if !yield(shape.circleSolver.wsCircle) {
			return
		}
	}
}

// CircleIter returns an iterator over all circle shapes in the scene that match
// the given filter.
func (s *Scene[O, S]) CircleIter(filter Filter) iter.Seq[shape2d.Circle] {
	return func(yield func(shape2d.Circle) bool) {
		s.EachCircle(filter, yield)
	}
}

// EachRectangle iterates over all rectangle shapes in the scene that match the
// given filter and yields them to the provided callback.
func (s *Scene[O, S]) EachRectangle(filter Filter, yield func(shape2d.Rectangle) bool) {
	for index := range uint32(len(s.rectangles)) {
		shape := &s.rectangles[index]
		if shape.spatialID == query2d.InvalidTreeItemID {
			continue
		}
		if !shape.matchesFilter(filter) {
			continue
		}
		if !yield(shape.rectangleSolver.wsRectangle) {
			return
		}
	}
}

// RectangleIter returns an iterator over all rectangle shapes in the scene that
// match the given filter.
func (s *Scene[O, S]) RectangleIter(filter Filter) iter.Seq[shape2d.Rectangle] {
	return func(yield func(shape2d.Rectangle) bool) {
		s.EachRectangle(filter, yield)
	}
}

// EachMesh iterates over all mesh shapes in the scene that match the
// given filter and yields them to the provided callback.
func (s *Scene[O, S]) EachMesh(filter Filter, yield func(shape2d.Mesh) bool) {
	for index := range uint32(len(s.meshes)) {
		shape := &s.meshes[index]
		if shape.spatialID == query2d.InvalidTreeItemID {
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

// CollectIntersections yields intersections found in this scene.
func (s *Scene[O, S]) CollectIntersections(yield ContactCallback) {
	s.checks = s.checks[:0]

	s.eachDynamicCircle(func(srcIndex uint32, srcCircle *sceneCircleShape[S]) {
		srcRef := newShapeRef(shapeKindCircle, srcIndex)
		queryAABB := query2d.AABBFromCircle(srcCircle.boundingCircle())
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	})

	s.eachDynamicRectangle(func(srcIndex uint32, srcRectangle *sceneRectangleShape[S]) {
		srcRef := newShapeRef(shapeKindRectangle, srcIndex)
		queryAABB := query2d.AABBFromCircle(srcRectangle.boundingCircle())
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
		queryAABB := query2d.AABBFromCircle(srcMesh.boundingCircle())
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

// CheckSegmentIntersection returns the largest intersection (likely first to
// have occurred) of the segment with the scene.
func (s *Scene[O, S]) CheckSegmentIntersection(segment shape2d.Segment, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectSegmentIntersections(segment, filter, collection.AddContact)
	return collection.Contact()
}

// CollectSegmentIntersections collects all intersections of the segment
// with objects in the scene.
func (s *Scene[O, S]) CollectSegmentIntersections(segment shape2d.Segment, filter Filter, yield ContactCallback) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempSegment = segment
	srcRef := newTempShapeRef(shapeKindSegment)

	s.checks = s.checks[:0]
	querySegment := query2d.NewSegment(segment.A, segment.B)
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

// CheckCircleIntersection returns the largest intersection (likely first to
// have occurred) of the circle with the scene.
func (s *Scene[O, S]) CheckCircleIntersection(circle shape2d.Circle, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectCircleIntersections(circle, filter, collection.AddContact)
	return collection.Contact()
}

// CollectCircleIntersections collects all intersections of the circle
// with objects in the scene.
func (s *Scene[O, S]) CollectCircleIntersections(circle shape2d.Circle, filter Filter, yield ContactCallback) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempCircle = newCircleSolver(circle)
	srcRef := newTempShapeRef(shapeKindCircle)

	s.checks = s.checks[:0]
	queryAABB := query2d.AABBFromCircle(circle)
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

// CheckRectangleIntersection returns the largest intersection (likely first to
// have occurred) of the rectangle with the scene.
func (s *Scene[O, S]) CheckRectangleIntersection(rectangle shape2d.Rectangle, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectRectangleIntersections(rectangle, filter, collection.AddContact)
	return collection.Contact()
}

// CollectRectangleIntersections collects all intersections of the rectangle
// with objects in the scene.
func (s *Scene[O, S]) CollectRectangleIntersections(rectangle shape2d.Rectangle, filter Filter, yield ContactCallback) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempRectangle = newRectangleSolver(rectangle)
	srcRef := newTempShapeRef(shapeKindRectangle)

	s.checks = s.checks[:0]
	queryAABB := query2d.AABBFromCircle(rectangle.BoundingCircle())
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

// CheckMeshIntersection returns the largest intersection (likely first to
// have occurred) of the mesh with the scene.
func (s *Scene[O, S]) CheckMeshIntersection(mesh shape2d.Mesh, filter Filter) (Contact, bool) {
	var collection DeepestContact
	s.CollectMeshIntersections(mesh, filter, collection.AddContact)
	return collection.Contact()
}

// CollectMeshIntersections collects all intersections of the mesh
// with objects in the scene.
func (s *Scene[O, S]) CollectMeshIntersections(mesh shape2d.Mesh, filter Filter, yield ContactCallback) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  filter.Mask.ValueOrDefault(0xFFFFFFFF),
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempMesh = newMeshSolver(mesh)
	srcRef := newTempShapeRef(shapeKindMesh)

	s.checks = s.checks[:0]
	queryAABB := query2d.AABBFromCircle(mesh.BoundingCircle())
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
	case shapeKindCircle:
		circle := &s.circles[ref.index()]
		return &circle.sceneShape
	case shapeKindRectangle:
		rectangle := &s.rectangles[ref.index()]
		return &rectangle.sceneShape
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
	case shapeKindCircle:
		circle := &s.circles[index]
		if circle.static {
			s.staticTree.Remove(circle.spatialID)
		} else {
			s.dynamicTree.Remove(circle.spatialID)
		}
		circle.spatialID = query2d.InvalidTreeItemID
		circle.userData = gog.Zero[S]() // in case of pointer
		circle.nextShape = invalidShapeRef
		circle.circleSolver = newCircleSolver(shape2d.Circle{})
		s.freeCircleIndices.Push(index)
	case shapeKindRectangle:
		rectangle := &s.rectangles[index]
		if rectangle.static {
			s.staticTree.Remove(rectangle.spatialID)
		} else {
			s.dynamicTree.Remove(rectangle.spatialID)
		}
		rectangle.spatialID = query2d.InvalidTreeItemID
		rectangle.userData = gog.Zero[S]() // in case of pointer
		rectangle.nextShape = invalidShapeRef
		rectangle.rectangleSolver = newRectangleSolver(shape2d.Rectangle{})
		s.freeRectangleIndices.Push(index)
	case shapeKindMesh:
		mesh := &s.meshes[index]
		if mesh.static {
			s.staticTree.Remove(mesh.spatialID)
		} else {
			s.dynamicTree.Remove(mesh.spatialID)
		}
		mesh.spatialID = query2d.InvalidTreeItemID
		mesh.userData = gog.Zero[S]() // in case of pointer
		mesh.nextShape = invalidShapeRef
		mesh.meshSolver = newMeshSolver(shape2d.Mesh{})
		s.freeMeshIndices.Push(index)
	default:
		panic("unknown shape reference")
	}
}

func (s *Scene[O, S]) eachDynamicCircle(cb func(uint32, *sceneCircleShape[S])) {
	for index := range uint32(len(s.circles)) {
		shape := &s.circles[index]
		if shape.static || (shape.spatialID == query2d.InvalidTreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[O, S]) eachDynamicRectangle(cb func(uint32, *sceneRectangleShape[S])) {
	for index := range uint32(len(s.rectangles)) {
		shape := &s.rectangles[index]
		if shape.static || (shape.spatialID == query2d.InvalidTreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[O, S]) eachDynamicMesh(cb func(uint32, *sceneMeshShape[S])) {
	for index := range uint32(len(s.meshes)) {
		shape := &s.meshes[index]
		if shape.static || (shape.spatialID == query2d.InvalidTreeItemID) {
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
		case srcKind == shapeKindSegment && tgtKind == shapeKindCircle:
			index = s.collectSegmentCircleIntersections(index, false, yield)
		case srcKind == shapeKindCircle && tgtKind == shapeKindSegment:
			index = s.collectSegmentCircleIntersections(index, true, yield)
		case srcKind == shapeKindSegment && tgtKind == shapeKindRectangle:
			index = s.collectSegmentRectangleIntersections(index, false, yield)
		case srcKind == shapeKindRectangle && tgtKind == shapeKindSegment:
			index = s.collectSegmentRectangleIntersections(index, true, yield)
		case srcKind == shapeKindSegment && tgtKind == shapeKindMesh:
			index = s.collectSegmentMeshIntersections(index, false, yield)
		case srcKind == shapeKindMesh && tgtKind == shapeKindSegment:
			index = s.collectSegmentMeshIntersections(index, true, yield)
		case srcKind == shapeKindCircle && tgtKind == shapeKindCircle:
			index = s.collectCircleCircleIntersections(index, yield)
		case srcKind == shapeKindCircle && tgtKind == shapeKindRectangle:
			index = s.collectCircleRectangleIntersections(index, false, yield)
		case srcKind == shapeKindRectangle && tgtKind == shapeKindCircle:
			index = s.collectCircleRectangleIntersections(index, true, yield)
		case srcKind == shapeKindCircle && tgtKind == shapeKindMesh:
			index = s.collectCircleMeshIntersections(index, false, yield)
		case srcKind == shapeKindMesh && tgtKind == shapeKindCircle:
			index = s.collectCircleMeshIntersections(index, true, yield)
		case srcKind == shapeKindRectangle && tgtKind == shapeKindMesh:
			index = s.collectRectangleMeshIntersections(index, false, yield)
		case srcKind == shapeKindMesh && tgtKind == shapeKindRectangle:
			index = s.collectRectangleMeshIntersections(index, true, yield)
		default:
			index++
		}
	}
}

func (s *Scene[O, S]) collectSegmentCircleIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getCircleSolver(tgtRef)

		s.checkSegmentCircleIntersection(srcSegment, tgtSolver, func(contact shape2d.Contact) {
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

func (s *Scene[O, S]) collectSegmentRectangleIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getRectangleSolver(tgtRef)

		s.checkSegmentRectangleIntersection(srcSegment, tgtSolver, func(contact shape2d.Contact) {
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

		s.checkSegmentMeshIntersection(srcSegment, tgtSolver, func(contact shape2d.Contact) {
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

func (s *Scene[O, S]) collectCircleCircleIntersections(index int, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, false, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getCircleSolver(srcRef)
		tgtShape, tgtSolver := s.getCircleSolver(tgtRef)

		s.checkCircleCircleIntersection(srcSolver, tgtSolver, func(contact shape2d.Contact) {
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

func (s *Scene[O, S]) collectCircleRectangleIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getCircleSolver(srcRef)
		tgtShape, tgtSolver := s.getRectangleSolver(tgtRef)

		s.checkCircleRectangleIntersection(srcSolver, tgtSolver, func(contact shape2d.Contact) {
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

func (s *Scene[O, S]) collectCircleMeshIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getCircleSolver(srcRef)
		tgtShape, tgtSolver := s.getMeshSolver(tgtRef)

		s.checkCircleMeshIntersection(srcSolver, tgtSolver, func(contact shape2d.Contact) {
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

func (s *Scene[O, S]) collectRectangleMeshIntersections(index int, flipped bool, yield ContactCallback) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getRectangleSolver(srcRef)
		tgtShape, tgtSolver := s.getMeshSolver(tgtRef)

		s.checkRectangleMeshIntersection(srcSolver, tgtSolver, func(contact shape2d.Contact) {
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

func (s *Scene[O, S]) getSegmentSolver(ref shapeRef) (*sceneShape[S], shape2d.Segment) {
	if !ref.isTemporary() {
		panic("expected temporary shape reference")
	}
	return &s.tempShape, s.tempSegment
}

func (s *Scene[O, S]) getCircleSolver(ref shapeRef) (*sceneShape[S], *circleSolver) {
	if ref.isTemporary() {
		return &s.tempShape, &s.tempCircle
	}
	circle := &s.circles[ref.index()]
	return &circle.sceneShape, &circle.circleSolver
}

func (s *Scene[O, S]) getRectangleSolver(ref shapeRef) (*sceneShape[S], *rectangleSolver) {
	if ref.isTemporary() {
		return &s.tempShape, &s.tempRectangle
	}
	rectangle := &s.rectangles[ref.index()]
	return &rectangle.sceneShape, &rectangle.rectangleSolver
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

func (s *Scene[O, S]) checkSegmentCircleIntersection(source shape2d.Segment, target *circleSolver, yield shape2d.ContactCallback) {
	isec2d.ResolveSegmentCircle(source, target.wsCircle, yield)
}

func (s *Scene[O, S]) checkSegmentRectangleIntersection(source shape2d.Segment, target *rectangleSolver, yield shape2d.ContactCallback) {
	if !isec2d.CheckSegmentCircleOverlap(source, target.wsBoundingCircle) {
		return
	}
	isec2d.ResolveSegmentRectangle(source, target.wsRectangle, yield)
}

func (s *Scene[O, S]) checkSegmentMeshIntersection(source shape2d.Segment, target *meshSolver, yield shape2d.ContactCallback) {
	if !isec2d.CheckSegmentCircleOverlap(source, target.wsBoundingCircle) {
		return
	}
	isec2d.ResolveSegmentMesh(source, target.wsMesh, yield)
}

func (s *Scene[O, S]) checkCircleCircleIntersection(source, target *circleSolver, yield shape2d.ContactCallback) {
	isec2d.ResolveCircleCircle(source.wsCircle, target.wsCircle, yield)
}

func (s *Scene[O, S]) checkCircleRectangleIntersection(source *circleSolver, target *rectangleSolver, yield shape2d.ContactCallback) {
	if !isec2d.CheckCircleCircle(source.wsCircle, target.wsBoundingCircle) {
		return
	}
	isec2d.ResolveCircleRectangle(source.wsCircle, target.wsRectangle, yield)
}

func (s *Scene[O, S]) checkCircleMeshIntersection(source *circleSolver, target *meshSolver, yield shape2d.ContactCallback) {
	if !isec2d.CheckCircleCircle(source.wsCircle, target.wsBoundingCircle) {
		return
	}
	isec2d.ResolveCircleMesh(source.wsCircle, target.wsMesh, yield)
}

func (s *Scene[O, S]) checkRectangleMeshIntersection(source *rectangleSolver, target *meshSolver, yield shape2d.ContactCallback) {
	if !isec2d.CheckCircleCircle(source.wsBoundingCircle, target.wsBoundingCircle) {
		return
	}
	isec2d.ResolveRectangleMesh(source.wsRectangle, target.wsMesh, yield)
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
