package shape2d

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

// NewScene creates a new scene.
func NewScene[O, S any](info SceneSettings) *Scene[O, S] {
	cubeOctreeSettings := CompactTreeSettings(info)

	return &Scene[O, S]{
		freeObjectIndices:    ds.NewStack[uint32](256), // ~ 1 KiB
		freeCircleIndices:    ds.NewStack[uint32](256), // ~ 1 KiB
		freeRectangleIndices: ds.NewStack[uint32](256), // ~ 1 KiB
		freePolygonIndices:   ds.NewStack[uint32](256), // ~ 1 KiB

		objects:    make([]sceneObject[O], 0, 128),
		circles:    make([]sceneCircleShape[S], 0, 128),
		rectangles: make([]sceneRectangleShape[S], 0, 128),
		polygons:   make([]scenePolygonShape[S], 0, 128),

		staticTree:  NewCompactTree[shapeRef](cubeOctreeSettings),
		dynamicTree: NewCompactTree[shapeRef](cubeOctreeSettings),

		checks: make([]shapeRefPair, 0, 1024),
	}
}

// Scene represents a 2D scene where objects made of shapes can be added.
type Scene[T, S any] struct {
	freeObjectIndices    *ds.Stack[uint32]
	freeCircleIndices    *ds.Stack[uint32]
	freeRectangleIndices *ds.Stack[uint32]
	freePolygonIndices   *ds.Stack[uint32]

	objects    []sceneObject[T]
	circles    []sceneCircleShape[S]
	rectangles []sceneRectangleShape[S]
	polygons   []scenePolygonShape[S]

	staticTree  *CompactTree[shapeRef]
	dynamicTree *CompactTree[shapeRef]

	tempShape     sceneShape[S]
	tempSegment   Segment
	tempCircle    circleSolver
	tempRectangle rectangleSolver
	tempPolygon   polygonSolver

	checks []shapeRefPair
}

// CreateObject creates a new object.
func (s *Scene[O, S]) CreateObject(info ObjectInfo[O]) ObjectID {
	transform := TRTransform(
		info.Position.ValueOrDefault(dprec.ZeroVec2()),
		info.Rotation.ValueOrDefault(dprec.Radians(0.0)),
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
func (s *Scene[O, S]) GetObjectTransform(objID ObjectID) Transform {
	object := &s.objects[objID]
	return object.transform
}

// SetObjectTransform relocates the given object.
func (s *Scene[O, S]) SetObjectTransform(objID ObjectID, transform Transform) {
	object := &s.objects[objID]
	object.transform = transform
	s.eachObjectShape(object, shapeKindCircle, func(index uint32) {
		circle := &s.circles[index]
		circle.update(transform)
		bc := NewCompactQuadFromCircle(circle.boundingCircle())
		if circle.static {
			s.staticTree.Update(circle.spatialID, bc)
		} else {
			s.dynamicTree.Update(circle.spatialID, bc)
		}
	})
	s.eachObjectShape(object, shapeKindRectangle, func(index uint32) {
		rectangle := &s.rectangles[index]
		rectangle.update(transform)
		bc := NewCompactQuadFromCircle(rectangle.boundingCircle())
		if rectangle.static {
			s.staticTree.Update(rectangle.spatialID, bc)
		} else {
			s.dynamicTree.Update(rectangle.spatialID, bc)
		}
	})
	s.eachObjectShape(object, shapeKindPolygon, func(index uint32) {
		polygon := &s.polygons[index]
		polygon.update(transform)
		bc := NewCompactQuadFromCircle(polygon.boundingCircle())
		if polygon.static {
			s.staticTree.Update(polygon.spatialID, bc)
		} else {
			s.dynamicTree.Update(polygon.spatialID, bc)
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
	var spatialID CompactTreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(NewCompactQuadFromCircle(bc), ref)
	} else {
		spatialID = s.dynamicTree.Insert(NewCompactQuadFromCircle(bc), ref)
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
	var spatialID CompactTreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(NewCompactQuadFromCircle(bc), ref)
	} else {
		spatialID = s.dynamicTree.Insert(NewCompactQuadFromCircle(bc), ref)
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

// AttachPolygon creates a polygon shape and attaches it to the object to be
// used for intersection tests.
func (s *Scene[O, S]) AttachPolygon(objID ObjectID, info PolygonInfo[S]) ShapeID {
	var index uint32
	if s.freePolygonIndices.IsEmpty() {
		index = uint32(len(s.polygons))
		s.polygons = append(s.polygons, scenePolygonShape[S]{})
	} else {
		index = s.freePolygonIndices.Pop()
	}
	ref := newShapeRef(shapeKindPolygon, index)

	object := &s.objects[objID]

	solver := newPolygonSolver(info.Polygon)
	solver.update(object.transform)

	bc := solver.boundingCircle()
	var spatialID CompactTreeItemID
	if object.isStatic() {
		spatialID = s.staticTree.Insert(NewCompactQuadFromCircle(bc), ref)
	} else {
		spatialID = s.dynamicTree.Insert(NewCompactQuadFromCircle(bc), ref)
	}

	s.polygons[index] = scenePolygonShape[S]{
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
		polygonSolver: solver,
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

	s.eachDynamicCircle(func(srcIndex uint32, srcCircle *sceneCircleShape[S]) {
		srcRef := newShapeRef(shapeKindCircle, srcIndex)
		queryAABB := NewCompactQueryAABBFromCircle(srcCircle.boundingCircle())
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
		queryAABB := NewCompactQueryAABBFromCircle(srcRectangle.boundingCircle())
		s.dynamicTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
		s.staticTree.QueryAABB(queryAABB, func(tgtRef shapeRef) bool {
			s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
			return true
		})
	})

	s.eachDynamicPolygon(func(srcIndex uint32, srcPolygon *scenePolygonShape[S]) {
		srcRef := newShapeRef(shapeKindPolygon, srcIndex)
		queryAABB := NewCompactQueryAABBFromCircle(srcPolygon.boundingCircle())
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

// CheckCircleIntersection returns the first intersection of the circle
// with the scene.
func (s *Scene[O, S]) CheckCircleIntersection(circle Circle, mask uint32) (ObjectIntersection, bool) {
	var collection LargestObjectIntersection
	s.CollectCircleIntersections(circle, mask, &collection)
	return collection.Intersection()
}

// CollectCircleIntersections collects all intersections of the circle
// with objects in the scene.
func (s *Scene[O, S]) CollectCircleIntersections(circle Circle, mask uint32, collection ObjectIntersectionCollection) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempCircle = newCircleSolver(circle)
	srcRef := newTempShapeRef(shapeKindCircle)

	s.checks = s.checks[:0]
	queryAABB := NewCompactQueryAABBFromCircle(circle)
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

// CheckRectangleIntersection returns the first intersection of the rectangle
// with the scene.
func (s *Scene[O, S]) CheckRectangleIntersection(rectangle Rectangle, mask uint32) (ObjectIntersection, bool) {
	var collection LargestObjectIntersection
	s.CollectRectangleIntersections(rectangle, mask, &collection)
	return collection.Intersection()
}

// CollectRectangleIntersections collects all intersections of the rectangle
// with objects in the scene.
func (s *Scene[O, S]) CollectRectangleIntersections(rectangle Rectangle, mask uint32, collection ObjectIntersectionCollection) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempRectangle = newRectangleSolver(rectangle)
	srcRef := newTempShapeRef(shapeKindRectangle)

	s.checks = s.checks[:0]
	queryAABB := NewCompactQueryAABBFromCircle(rectangle.BoundingCircle())
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

// CheckPolygonIntersection returns the first intersection of the polygon
// with the scene.
func (s *Scene[O, S]) CheckPolygonIntersection(polygon Polygon, mask uint32) (ObjectIntersection, bool) {
	var collection LargestObjectIntersection
	s.CollectPolygonIntersections(polygon, mask, &collection)
	return collection.Intersection()
}

// CollectPolygonIntersections collects all intersections of the polygon
// with objects in the scene.
func (s *Scene[O, S]) CollectPolygonIntersections(polygon Polygon, mask uint32, collection ObjectIntersectionCollection) {
	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
		static:      true, // important, otherwise double-check prevention will kick in
	}
	s.tempPolygon = newPolygonSolver(polygon)
	srcRef := newTempShapeRef(shapeKindPolygon)

	s.checks = s.checks[:0]
	queryAABB := NewCompactQueryAABBFromCircle(polygon.BoundingCircle())
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
	case shapeKindCircle:
		circle := &s.circles[ref.index()]
		return &circle.sceneShape
	case shapeKindRectangle:
		rectangle := &s.rectangles[ref.index()]
		return &rectangle.sceneShape
	case shapeKindPolygon:
		polygon := &s.polygons[ref.index()]
		return &polygon.sceneShape
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
		circle.spatialID = InvalidCompactTreeItemID
		circle.userData = gog.Zero[S]() // in case of pointer
		circle.nextShape = invalidShapeRef
		circle.circleSolver = newCircleSolver(Circle{})
		s.freeCircleIndices.Push(index)
	case shapeKindRectangle:
		rectangle := &s.rectangles[index]
		if rectangle.static {
			s.staticTree.Remove(rectangle.spatialID)
		} else {
			s.dynamicTree.Remove(rectangle.spatialID)
		}
		rectangle.spatialID = InvalidCompactTreeItemID
		rectangle.userData = gog.Zero[S]() // in case of pointer
		rectangle.nextShape = invalidShapeRef
		rectangle.rectangleSolver = newRectangleSolver(Rectangle{})
		s.freeRectangleIndices.Push(index)
	case shapeKindPolygon:
		polygon := &s.polygons[index]
		if polygon.static {
			s.staticTree.Remove(polygon.spatialID)
		} else {
			s.dynamicTree.Remove(polygon.spatialID)
		}
		polygon.spatialID = InvalidCompactTreeItemID
		polygon.userData = gog.Zero[S]() // in case of pointer
		polygon.nextShape = invalidShapeRef
		polygon.polygonSolver = newPolygonSolver(Polygon{})
		s.freePolygonIndices.Push(index)
	default:
		panic("unknown shape reference")
	}
}

func (s *Scene[O, S]) eachDynamicCircle(cb func(uint32, *sceneCircleShape[S])) {
	for index := range uint32(len(s.circles)) {
		shape := &s.circles[index]
		if shape.static || (shape.spatialID == InvalidCompactTreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[O, S]) eachDynamicRectangle(cb func(uint32, *sceneRectangleShape[S])) {
	for index := range uint32(len(s.rectangles)) {
		shape := &s.rectangles[index]
		if shape.static || (shape.spatialID == InvalidCompactTreeItemID) {
			continue
		}
		cb(index, shape)
	}
}

func (s *Scene[O, S]) eachDynamicPolygon(cb func(uint32, *scenePolygonShape[S])) {
	for index := range uint32(len(s.polygons)) {
		shape := &s.polygons[index]
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
		case srcKind == shapeKindSegment && tgtKind == shapeKindCircle:
			index = s.collectSegmentCircleIntersections(index, false, collection)
		case srcKind == shapeKindCircle && tgtKind == shapeKindSegment:
			index = s.collectSegmentCircleIntersections(index, true, collection)
		case srcKind == shapeKindSegment && tgtKind == shapeKindRectangle:
			index = s.collectSegmentRectangleIntersections(index, false, collection)
		case srcKind == shapeKindRectangle && tgtKind == shapeKindSegment:
			index = s.collectSegmentRectangleIntersections(index, true, collection)
		case srcKind == shapeKindSegment && tgtKind == shapeKindPolygon:
			index = s.collectSegmentPolygonIntersections(index, false, collection)
		case srcKind == shapeKindPolygon && tgtKind == shapeKindSegment:
			index = s.collectSegmentPolygonIntersections(index, true, collection)
		case srcKind == shapeKindCircle && tgtKind == shapeKindCircle:
			index = s.collectCircleCircleIntersections(index, collection)
		case srcKind == shapeKindCircle && tgtKind == shapeKindRectangle:
			index = s.collectCircleRectangleIntersections(index, false, collection)
		case srcKind == shapeKindRectangle && tgtKind == shapeKindCircle:
			index = s.collectCircleRectangleIntersections(index, true, collection)
		case srcKind == shapeKindCircle && tgtKind == shapeKindPolygon:
			index = s.collectCirclePolygonIntersections(index, false, collection)
		case srcKind == shapeKindPolygon && tgtKind == shapeKindCircle:
			index = s.collectCirclePolygonIntersections(index, true, collection)
		case srcKind == shapeKindRectangle && tgtKind == shapeKindPolygon:
			index = s.collectRectanglePolygonIntersections(index, false, collection)
		case srcKind == shapeKindPolygon && tgtKind == shapeKindRectangle:
			index = s.collectRectanglePolygonIntersections(index, true, collection)
		default:
			index++
		}
	}
}

func (s *Scene[O, S]) collectSegmentCircleIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getCircleSolver(tgtRef)

		if intersection, ok := s.checkSegmentCircleIntersection(srcSegment, tgtSolver); ok {
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

func (s *Scene[O, S]) collectSegmentRectangleIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getRectangleSolver(tgtRef)

		if intersection, ok := s.checkSegmentRectangleIntersection(srcSegment, tgtSolver); ok {
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

func (s *Scene[O, S]) collectSegmentPolygonIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSegment := s.getSegmentSolver(srcRef)
		tgtShape, tgtSolver := s.getPolygonSolver(tgtRef)

		if intersection, ok := s.checkSegmentPolygonIntersection(srcSegment, tgtSolver); ok {
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

func (s *Scene[O, S]) collectCircleCircleIntersections(index int, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, false, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getCircleSolver(srcRef)
		tgtShape, tgtSolver := s.getCircleSolver(tgtRef)

		if intersection, ok := s.checkCircleCircleIntersection(srcSolver, tgtSolver); ok {
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

func (s *Scene[O, S]) collectCircleRectangleIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getCircleSolver(srcRef)
		tgtShape, tgtSolver := s.getRectangleSolver(tgtRef)

		if intersection, ok := s.checkCircleRectangleIntersection(srcSolver, tgtSolver); ok {
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

func (s *Scene[O, S]) collectCirclePolygonIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getCircleSolver(srcRef)
		tgtShape, tgtSolver := s.getPolygonSolver(tgtRef)

		if intersection, ok := s.checkCirclePolygonIntersection(srcSolver, tgtSolver); ok {
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

func (s *Scene[O, S]) collectRectanglePolygonIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		tgtRef := refPair.target()

		srcShape, srcSolver := s.getRectangleSolver(srcRef)
		tgtShape, tgtSolver := s.getPolygonSolver(tgtRef)

		if intersection, ok := s.checkRectanglePolygonIntersection(srcSolver, tgtSolver); ok {
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

func (s *Scene[O, S]) getPolygonSolver(ref shapeRef) (*sceneShape[S], *polygonSolver) {
	if ref.isTemporary() {
		return &s.tempShape, &s.tempPolygon
	}
	polygon := &s.polygons[ref.index()]
	return &polygon.sceneShape, &polygon.polygonSolver
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

func (s *Scene[O, S]) checkSegmentCircleIntersection(source Segment, target *circleSolver) (Intersection, bool) {
	return CheckSegmentCircleIntersection(source, target.wsCircle)
}

func (s *Scene[O, S]) checkSegmentRectangleIntersection(source Segment, target *rectangleSolver) (Intersection, bool) {
	if !IsSegmentCircleOverlap(source, target.wsBoundingCircle) {
		return Intersection{}, false
	}
	return CheckSegmentRectangleIntersection(source, target.wsRectangle)
}

func (s *Scene[O, S]) checkSegmentPolygonIntersection(source Segment, target *polygonSolver) (Intersection, bool) {
	if !IsSegmentCircleOverlap(source, target.wsBoundingCircle) {
		return Intersection{}, false
	}
	return CheckSegmentPolygonIntersection(source, target.wsPolygon)
}

func (s *Scene[O, S]) checkCircleCircleIntersection(source, target *circleSolver) (Intersection, bool) {
	return CheckCircleCircleIntersection(source.wsCircle, target.wsCircle)
}

func (s *Scene[O, S]) checkCircleRectangleIntersection(source *circleSolver, target *rectangleSolver) (Intersection, bool) {
	if !IsCircleCircleIntersection(source.wsCircle, target.wsBoundingCircle) {
		return Intersection{}, false
	}
	return CheckCircleRectangleIntersection(source.wsCircle, target.wsRectangle)
}

func (s *Scene[O, S]) checkCirclePolygonIntersection(source *circleSolver, target *polygonSolver) (Intersection, bool) {
	if !IsCircleCircleIntersection(source.wsCircle, target.wsBoundingCircle) {
		return Intersection{}, false
	}
	return CheckCirclePolygonIntersection(source.wsCircle, target.wsPolygon)
}

func (s *Scene[O, S]) checkRectanglePolygonIntersection(source *rectangleSolver, target *polygonSolver) (Intersection, bool) {
	if !IsCircleCircleIntersection(source.wsBoundingCircle, target.wsBoundingCircle) {
		return Intersection{}, false
	}
	return CheckRectanglePolygonIntersection(source.wsRectangle, target.wsPolygon)
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
