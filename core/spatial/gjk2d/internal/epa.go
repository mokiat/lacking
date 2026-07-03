package internal

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

// EPASolver computes contact information (separation normal, penetration
// depth and closest feature) for two overlapping shapes, based on the final
// simplex produced by a [GJKSolver] run over the same [MinkowskiShape].
//
// Usage: call [EPASolver.Reset] with the shape and the final GJK simplex,
// then call [EPASolver.Next] repeatedly until it returns false, then obtain
// the result through [EPASolver.Solution].
type EPASolver struct {
	polytope            map[PolytopeEdgeID]PolytopeEdge
	solution            EPASolution
	skinRadius          float64
	remainingIterations uint32
}

// NewEPASolver creates a new [EPASolver] instance.
func NewEPASolver() *EPASolver {
	return &EPASolver{
		polytope: make(map[PolytopeEdgeID]PolytopeEdge),
	}
}

// Reset prepares the solver for a new query against the given shape. The
// simplex must be the final simplex of a [GJKSolver] run over the same
// shape and containsOrigin must be the corresponding
// [GJKSolver.ContainsOrigin] result.
//
// When the origin is not contained, the simplex is the point or edge
// feature of the Minkowski difference that is closest to the origin, and
// the solution is computed immediately, requiring no iteration.
func (s *EPASolver) Reset(shape *MinkowskiShape, simplex Simplex, containsOrigin bool) {
	clear(s.polytope)
	s.skinRadius = shape.SkinRadius
	s.remainingIterations = uint32(shape.MaxIterations())

	if !containsOrigin {
		switch simplex.VertexCount {
		case 1:
			s.terminatePoint(simplex.Vertices[0])
		case 2:
			s.terminateEdge(simplex.Vertices[0], simplex.Vertices[1])
		default:
			panic("unexpected simplex vertex count")
		}
	} else {
		s.addPolytopeEdge(simplex.Vertices[0], simplex.Vertices[1])
		s.addPolytopeEdge(simplex.Vertices[1], simplex.Vertices[2])
		s.addPolytopeEdge(simplex.Vertices[2], simplex.Vertices[0])
	}
}

// Next runs a single EPA iteration. It returns false once the algorithm
// has converged (or the iteration budget is exhausted) and further calls
// would make no progress.
func (s *EPASolver) Next(shape *MinkowskiShape) bool {
	if s.remainingIterations == 0 {
		return false
	}
	s.remainingIterations--

	var (
		minEdgeID   PolytopeEdgeID
		minDistance = math.MaxFloat64
	)
	for edgeID, edge := range s.polytope {
		if edge.Distance < minDistance {
			minEdgeID = edgeID
			minDistance = edge.Distance
		}
	}
	minEdge := s.polytope[minEdgeID]

	pointA := minEdge.VertexA.Position
	pointB := minEdge.VertexB.Position
	edgeVec := dprec.Vec2Diff(pointB, pointA)
	edgeNorm := transposeVec2(edgeVec)
	support := shape.Support(edgeNorm)

	if s.polytopeContainsVertex(support.Refs) {
		s.terminateEdge(minEdge.VertexA, minEdge.VertexB)
		return false
	}

	delete(s.polytope, minEdgeID)
	s.addPolytopeEdge(minEdge.VertexA, support)
	s.addPolytopeEdge(support, minEdge.VertexB)

	return true
}

// Solution returns the computed [EPASolution]. The result is only reliable
// once [EPASolver.Next] has been iterated until it returned false.
func (s *EPASolver) Solution() EPASolution {
	return s.solution
}

// terminatePoint completes the solve with a single vertex as the closest
// feature of the Minkowski difference to the origin.
func (s *EPASolver) terminatePoint(vertex MinkowskiVertex) {
	s.remainingIterations = 0 // ensure we can't be asked to iterate further
	s.solution = EPASolution{
		VertexA: vertex,
		VertexB: vertex,
		Normal:  s.pointNormal(vertex.Position),
		Lerp:    0.0,
		Depth:   s.pointDepth(vertex.Position),
	}
}

// terminateEdge completes the solve with an edge as the closest feature of
// the Minkowski difference to the origin, falling back to one of its
// vertices when the origin does not project onto the edge.
func (s *EPASolver) terminateEdge(vertexA, vertexB MinkowskiVertex) {
	s.remainingIterations = 0 // ensure we can't be asked to iterate further

	edge := dprec.Vec2Diff(vertexB.Position, vertexA.Position)
	dot := -dprec.Vec2Dot(edge, vertexA.Position)
	sqrLength := edge.SqrLength()

	switch {
	case dot <= 0.0:
		s.terminatePoint(vertexA)
	case dot >= sqrLength:
		s.terminatePoint(vertexB)
	default:
		// Note: The previous cases will handle a zero length edge.
		lerp := dot / sqrLength
		edgePoint := dprec.Vec2Lerp(vertexA.Position, vertexB.Position, lerp)
		s.solution = EPASolution{
			VertexA: vertexA,
			VertexB: vertexB,
			Normal:  s.pointNormal(edgePoint),
			Lerp:    lerp,
			Depth:   s.pointDepth(edgePoint),
		}
	}
}

// pointNormal returns the unit direction from the point toward the origin.
// An arbitrary fixed direction is returned when the point coincides with
// the origin.
func (s *EPASolver) pointNormal(point dprec.Vec2) dprec.Vec2 {
	length := point.Length()
	if length == 0.0 {
		return dprec.BasisXVec2()
	}
	return dprec.Vec2Quot(point, -length)
}

// pointDepth returns the penetration depth for the given closest point,
// which is the amount by which the combined skin radius exceeds the
// point's distance to the origin.
func (s *EPASolver) pointDepth(point dprec.Vec2) float64 {
	return s.skinRadius - point.Length()
}

func (s *EPASolver) addPolytopeEdge(vertexA, vertexB MinkowskiVertex) {
	pointA := vertexA.Position
	pointB := vertexB.Position

	edge := dprec.Vec2Diff(pointB, pointA)
	sqrLength := edge.SqrLength()
	var distance float64
	if sqrLength == 0.0 {
		distance = pointA.Length()
	} else {
		t := -dprec.Vec2Dot(edge, pointA) / sqrLength
		if t <= 0.0 {
			distance = pointA.Length()
		} else if t >= 1.0 {
			distance = pointB.Length()
		} else {
			projection := dprec.Vec2Lerp(pointA, pointB, t)
			distance = projection.Length()
		}
	}

	id := PolytopeEdgeID{vertexA.Refs, vertexB.Refs}
	s.polytope[id] = PolytopeEdge{
		VertexA:  vertexA,
		VertexB:  vertexB,
		Distance: distance,
	}
}

func (s *EPASolver) polytopeContainsVertex(refs RefPair) bool {
	for _, edge := range s.polytope {
		if edge.VertexA.Refs == refs || edge.VertexB.Refs == refs {
			return true
		}
	}
	return false
}

// EPASolution describes how two overlapping shapes can be separated.
type EPASolution struct {
	// VertexA and VertexB are the Minkowski difference vertices that bound
	// the closest feature. They are equal when the feature is a single
	// vertex.
	VertexA MinkowskiVertex
	VertexB MinkowskiVertex
	// Normal is the unit direction from the closest feature of the
	// Minkowski difference toward the origin. In world space it points from
	// the target shape toward the source shape.
	Normal dprec.Vec2
	// Lerp is the interpolation factor between VertexA and VertexB at which
	// the closest point to the origin lies.
	Lerp float64
	// Depth is the amount by which the shapes, inflated by their skin
	// radii, overlap along Normal.
	Depth float64
}

type PolytopeEdgeID [2]RefPair

type PolytopeEdge struct {
	VertexA  MinkowskiVertex
	VertexB  MinkowskiVertex
	Distance float64
}
