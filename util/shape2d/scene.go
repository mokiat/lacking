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

	tempShape   sceneShape[S]
	tempSegment Segment

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
	// TODO: Continue implementing from here...

	// 	// Circle vs Circle intersections.
	// 	s.checks = s.checks[:0]
	// 	s.eachDynamicCircle(func(srcIndex uint32, srcCircle *sceneCircleShape[S]) {
	// 		area := NewCompactQueryAABBFromCircle(srcCircle.boundingCircle())
	// 		s.circleTree.QueryAABB(area, func(tgtIndex uint32) bool {
	// 			tgtCircle := &s.circles[tgtIndex]
	// 			if (srcIndex < tgtIndex) && shapesCanIntersect(&srcCircle.sceneShape, &tgtCircle.sceneShape) {
	// 				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
	// 			}
	// 			return true
	// 		})
	// 	})
	// 	s.collectCircleCircleIntersections(s.checks, collection)

	// 	// Circle vs Rectangle intersections.
	// 	s.checks = s.checks[:0]
	// 	s.eachDynamicCircle(func(srcIndex uint32, srcCircle *sceneCircleShape[S]) {
	// 		area := NewCompactQueryAABBFromCircle(srcCircle.boundingCircle())
	// 		s.rectangleTree.QueryAABB(area, func(tgtIndex uint32) bool {
	// 			tgtRectangle := &s.rectangles[tgtIndex]
	// 			if shapesCanIntersect(&srcCircle.sceneShape, &tgtRectangle.sceneShape) {
	// 				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
	// 			}
	// 			return true
	// 		})
	// 	})
	// 	s.eachDynamicRectangle(func(srcIndex uint32, srcRectangle *sceneRectangleShape[S]) {
	// 		area := NewCompactQueryAABBFromCircle(srcRectangle.boundingCircle())
	// 		s.circleTree.QueryAABB(area, func(tgtIndex uint32) bool {
	// 			tgtCircle := &s.circles[tgtIndex]
	// 			if shapesCanIntersect(&tgtCircle.sceneShape, &srcRectangle.sceneShape) {
	// 				s.checks = append(s.checks, newIndexPair(tgtIndex, srcIndex)) // flipped
	// 			}
	// 			return true
	// 		})
	// 	})
	// 	s.collectCircleRectangleIntersections(s.checks, collection)

	// 	// Circle vs Polygon intersections.
	// 	s.checks = s.checks[:0]
	// 	s.eachDynamicCircle(func(srcIndex uint32, srcCircle *sceneCircleShape[S]) {
	// 		area := NewCompactQueryAABBFromCircle(srcCircle.boundingCircle())
	// 		s.polygonTree.QueryAABB(area, func(tgtIndex uint32) bool {
	// 			tgtPolygon := &s.polygons[tgtIndex]
	// 			if shapesCanIntersect(&srcCircle.sceneShape, &tgtPolygon.sceneShape) {
	// 				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
	// 			}
	// 			return true
	// 		})
	// 	})
	// 	s.eachDynamicPolygon(func(srcIndex uint32, srcMesh *scenePolygonShape[S]) {
	// 		area := NewCompactQueryAABBFromCircle(srcMesh.boundingCircle())
	// 		s.circleTree.QueryAABB(area, func(tgtIndex uint32) bool {
	// 			tgtCircle := &s.circles[tgtIndex]
	// 			if shapesCanIntersect(&tgtCircle.sceneShape, &srcMesh.sceneShape) {
	// 				s.checks = append(s.checks, newIndexPair(tgtIndex, srcIndex)) // flipped
	// 			}
	// 			return true
	// 		})
	// 	})
	// 	s.collectCirclePolygonIntersections(s.checks, collection)

	// // Rectangle vs Polygon intersections.
	// s.checks = s.checks[:0]
	//
	//	s.eachDynamicRectangle(func(srcIndex uint32, srcRectangle *sceneRectangleShape[S]) {
	//		area := NewCompactQueryAABBFromCircle(srcRectangle.boundingCircle())
	//		s.polygonTree.QueryAABB(area, func(tgtIndex uint32) bool {
	//			tgtPolygon := &s.polygons[tgtIndex]
	//			if shapesCanIntersect(&srcRectangle.sceneShape, &tgtPolygon.sceneShape) {
	//				s.checks = append(s.checks, newIndexPair(srcIndex, tgtIndex))
	//			}
	//			return true
	//		})
	//	})
	//
	//	s.eachDynamicPolygon(func(srcIndex uint32, srcPolygon *scenePolygonShape[S]) {
	//		area := NewCompactQueryAABBFromCircle(srcPolygon.boundingCircle())
	//		s.rectangleTree.QueryAABB(area, func(tgtIndex uint32) bool {
	//			tgtRectangle := &s.rectangles[tgtIndex]
	//			if shapesCanIntersect(&tgtRectangle.sceneShape, &srcPolygon.sceneShape) {
	//				s.checks = append(s.checks, newIndexPair(tgtIndex, srcIndex)) // flipped
	//			}
	//			return true
	//		})
	//	})
	//
	// s.collectRectanglePolygonIntersections(s.checks, collection)
}

// CollectSegmentIntersections collects all intersections of the segment
// with objects in the scene.
func (s *Scene[O, S]) CollectSegmentIntersections(segment Segment, mask uint32, collection ObjectIntersectionCollection) {
	querySegment := NewCompactQuerySegment(segment.A, segment.B)

	s.tempShape = sceneShape[S]{
		objectIndex: invalidObjectIndex,
		targetMask:  mask,
	}
	s.tempSegment = segment
	srcRef := newTempShapeRef(shapeKindSegment)

	s.checks = s.checks[:0]
	s.dynamicTree.QuerySegment(querySegment, func(tgtRef shapeRef) bool {
		s.checks = append(s.checks, newShapeRefPair(srcRef, tgtRef))
		return true
	})
	s.collectIntersections(collection)
}

// GC cleans up internal data and allows for memory reuse. This should be
// called once per frame.
func (s *Scene[O, S]) GC() {
	s.dynamicTree.GC()
}

func (s *Scene[O, S]) collectIntersections(collection ObjectIntersectionCollection) {
	slices.Sort(s.checks)

	index := 0
	for index < len(s.checks) {
		refPair := s.checks[index]

		srcRef := refPair.source()
		srcIndex := srcRef.index()
		srcKind := srcRef.kind()

		tgtRef := refPair.target()
		tgtIndex := tgtRef.index()
		tgtKind := tgtRef.kind()
		if (srcKind == tgtKind) && (srcIndex >= tgtIndex) {
			index++
			continue // prevent self-intersection and double checks
		}

		srcShape := s.getShape(srcRef)
		tgtShape := s.getShape(tgtRef)
		if !shapesCanIntersect(srcShape, tgtShape) {
			index++
			continue // shapes have mismatching rules
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
		default:
			index++
		}
	}
}

func (s *Scene[O, S]) collectSegmentCircleIntersections(index int, flipped bool, collection ObjectIntersectionCollection) int {
	return s.consumeSameKindRefPairs(index, flipped, func(refPair shapeRefPair) {
		srcRef := refPair.source()
		if !srcRef.isTemporary() {
			panic("expected temporary source shape reference")
		}
		tgtRef := refPair.target()
		tgtCircleIndex := tgtRef.index()
		tgtCircle := s.circles[tgtCircleIndex]

		if intersection, ok := CheckSegmentCircleIntersection(s.tempSegment, tgtCircle.wsCircle); ok {
			intersection := ObjectIntersection{
				SourceObjectID: InvalidObjectID,
				SourceShapeID:  InvalidShapeID,
				TargetObjectID: ObjectID(tgtCircle.objectIndex),
				TargetShapeID:  ShapeID(newShapeRef(shapeKindCircle, tgtCircleIndex)),
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
		if !srcRef.isTemporary() {
			panic("expected temporary source shape reference")
		}
		tgtRef := refPair.target()
		tgtRectangleIndex := tgtRef.index()
		tgtRectangle := s.rectangles[tgtRectangleIndex]

		if intersection, ok := CheckSegmentRectangleIntersection(s.tempSegment, tgtRectangle.wsRectangle); ok {
			intersection := ObjectIntersection{
				SourceObjectID: InvalidObjectID,
				SourceShapeID:  InvalidShapeID,
				TargetObjectID: ObjectID(tgtRectangle.objectIndex),
				TargetShapeID:  ShapeID(newShapeRef(shapeKindRectangle, tgtRectangleIndex)),
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
		if !srcRef.isTemporary() {
			panic("expected temporary source shape reference")
		}
		tgtRef := refPair.target()
		tgtPolygonIndex := tgtRef.index()
		tgtPolygon := s.polygons[tgtPolygonIndex]

		if intersection, ok := CheckSegmentPolygonIntersection(s.tempSegment, tgtPolygon.wsPolygon); ok {
			intersection := ObjectIntersection{
				SourceObjectID: InvalidObjectID,
				SourceShapeID:  InvalidShapeID,
				TargetObjectID: ObjectID(tgtPolygon.objectIndex),
				TargetShapeID:  ShapeID(newShapeRef(shapeKindPolygon, tgtPolygonIndex)),
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
		s.dynamicTree.Remove(circle.spatialID)
		circle.spatialID = InvalidCompactTreeItemID
		circle.userData = gog.Zero[S]() // in case of pointer
		circle.nextShape = invalidShapeRef
		circle.circleSolver = newCircleSolver(Circle{})
		s.freeCircleIndices.Push(index)
	case shapeKindRectangle:
		rectangle := &s.rectangles[index]
		s.dynamicTree.Remove(rectangle.spatialID)
		rectangle.spatialID = InvalidCompactTreeItemID
		rectangle.userData = gog.Zero[S]() // in case of pointer
		rectangle.nextShape = invalidShapeRef
		rectangle.rectangleSolver = newRectangleSolver(Rectangle{})
		s.freeRectangleIndices.Push(index)
	case shapeKindPolygon:
		polygon := &s.polygons[index]
		s.dynamicTree.Remove(polygon.spatialID)
		polygon.spatialID = InvalidCompactTreeItemID
		polygon.userData = gog.Zero[S]() // in case of pointer
		polygon.nextShape = invalidShapeRef
		polygon.polygonSolver = newPolygonSolver(Polygon{})
		s.freePolygonIndices.Push(index)
	default:
		panic("unknown shape reference")
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

// func (s *Scene[O, S]) collectCircleCircleIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
// 	// lastPair := invalidIndexPair
// 	// slices.Sort(pairs)
// 	// for _, pair := range pairs {
// 	// 	if pair != lastPair {
// 	// 		srcSphereIndex := pair.srcIndex()
// 	// 		srcSphere := &s.spheres[srcSphereIndex]
// 	// 		tgtSphereIndex := pair.tgtIndex()
// 	// 		tgtSphere := &s.spheres[tgtSphereIndex]
// 	// 		if intersection, ok := s.checkSphereSphereIntersection(&srcSphere.sphereSolver, &tgtSphere.sphereSolver); ok {
// 	// 			collection.AddIntersection(ObjectIntersection{
// 	// 				SourceObjectID: ObjectID(srcSphere.objectIndex),
// 	// 				SourceShapeID:  ShapeID(newShapeRef(shapeKindSphere, srcSphereIndex)),
// 	// 				TargetObjectID: ObjectID(tgtSphere.objectIndex),
// 	// 				TargetShapeID:  ShapeID(newShapeRef(shapeKindSphere, tgtSphereIndex)),
// 	// 				Intersection:   intersection,
// 	// 			})
// 	// 		}
// 	// 	}
// 	// 	lastPair = pair
// 	// }
// }

// func (s *Scene[O, S]) collectCircleRectangleIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
// 	// lastPair := invalidIndexPair
// 	// slices.Sort(pairs)
// 	// for _, pair := range pairs {
// 	// 	if pair != lastPair {
// 	// 		srcSphereIndex := pair.srcIndex()
// 	// 		srcSphere := &s.spheres[srcSphereIndex]
// 	// 		tgtBoxIndex := pair.tgtIndex()
// 	// 		tgtBox := &s.boxes[tgtBoxIndex]
// 	// 		if intersection, ok := s.checkSphereBoxIntersection(&srcSphere.sphereSolver, &tgtBox.boxSolver); ok {
// 	// 			collection.AddIntersection(ObjectIntersection{
// 	// 				SourceObjectID: ObjectID(srcSphere.objectIndex),
// 	// 				SourceShapeID:  ShapeID(newShapeRef(shapeKindSphere, srcSphereIndex)),
// 	// 				TargetObjectID: ObjectID(tgtBox.objectIndex),
// 	// 				TargetShapeID:  ShapeID(newShapeRef(shapeKindBox, tgtBoxIndex)),
// 	// 				Intersection:   intersection,
// 	// 			})
// 	// 		}
// 	// 	}
// 	// 	lastPair = pair
// 	// }
// }

// func (s *Scene[O, S]) collectCirclePolygonIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
// 	// lastPair := invalidIndexPair
// 	// slices.Sort(pairs)
// 	// for _, pair := range pairs {
// 	// 	if pair != lastPair {
// 	// 		srcSphereIndex := pair.srcIndex()
// 	// 		srcSphere := &s.spheres[srcSphereIndex]
// 	// 		tgtMeshIndex := pair.tgtIndex()
// 	// 		tgtMesh := &s.meshes[tgtMeshIndex]
// 	// 		if intersection, ok := s.checkSphereMeshIntersection(&srcSphere.sphereSolver, &tgtMesh.meshSolver); ok {
// 	// 			collection.AddIntersection(ObjectIntersection{
// 	// 				SourceObjectID: ObjectID(srcSphere.objectIndex),
// 	// 				SourceShapeID:  ShapeID(newShapeRef(shapeKindSphere, srcSphereIndex)),
// 	// 				TargetObjectID: ObjectID(tgtMesh.objectIndex),
// 	// 				TargetShapeID:  ShapeID(newShapeRef(shapeKindMesh, tgtMeshIndex)),
// 	// 				Intersection:   intersection,
// 	// 			})
// 	// 		}
// 	// 	}
// 	// 	lastPair = pair
// 	// }
// }

//	func (s *Scene[O, S]) collectRectanglePolygonIntersections(pairs []indexPair, collection ObjectIntersectionCollection) {
//		// lastPair := invalidIndexPair
//		// slices.Sort(pairs)
//		// for _, pair := range pairs {
//		// 	if pair != lastPair {
//		// 		srcBoxIndex := pair.srcIndex()
//		// 		srcBox := &s.boxes[srcBoxIndex]
//		// 		tgtMeshIndex := pair.tgtIndex()
//		// 		tgtMesh := &s.meshes[tgtMeshIndex]
//		// 		if intersection, ok := s.checkBoxMeshIntersection(&srcBox.boxSolver, &tgtMesh.meshSolver); ok {
//		// 			collection.AddIntersection(ObjectIntersection{
//		// 				SourceObjectID: ObjectID(srcBox.objectIndex),
//		// 				SourceShapeID:  ShapeID(newShapeRef(shapeKindBox, srcBoxIndex)),
//		// 				TargetObjectID: ObjectID(tgtMesh.objectIndex),
//		// 				TargetShapeID:  ShapeID(newShapeRef(shapeKindMesh, tgtMeshIndex)),
//		// 				Intersection:   intersection,
//		// 			})
//		// 		}
//		// 	}
//		// 	lastPair = pair
//		// }
//	}
