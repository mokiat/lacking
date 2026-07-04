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
	s.solution = EPASolution{
		VertexA: simplex.Vertices[0],
		VertexB: simplex.Vertices[0],
		Normal:  dprec.BasisXVec2(),
		Lerp:    0.0,
		Depth:   0.0,
	}
	s.skinRadius = shape.SkinRadius
	s.remainingIterations = uint32(shape.MaxIterations())

	if !containsOrigin {
		switch simplex.VertexCount {
		case 1:
			s.terminatePoint(shape, simplex.Vertices[0])
		case 2:
			// Pass a flipped edge as we want all terminal edges to point towards
			// separation direction.
			s.terminateEdge(shape, simplex.Vertices[1], simplex.Vertices[0])
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
		s.terminateEdge(shape, minEdge.VertexA, minEdge.VertexB)
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

// terminateEdge completes the solve with an edge as the closest feature of
// the Minkowski difference to the origin, falling back to one of its
// vertices when the origin does not project onto the edge.
//
// Both call sites guarantee that the origin lies on the left side of the
// directed edge from vertexA to vertexB (the [GJKSolver.Simplex] invariant
// for the non-contained case and the counter-clockwise polytope winding for
// the contained case), so the right-hand normal of the edge points from the
// origin toward the edge.
func (s *EPASolver) terminateEdge(shape *MinkowskiShape, vertexA, vertexB MinkowskiVertex) {
	edge := dprec.Vec2Diff(vertexB.Position, vertexA.Position)
	sqrLength := edge.SqrLength()
	if sqrLength < 1e-12 {
		s.terminatePoint(shape, vertexA) // treat degenerate edges as points
		return
	}

	dot := -dprec.Vec2Dot(edge, vertexA.Position)
	normal := dprec.NormalVec2(edge)
	distance := dprec.Vec2Dot(normal, vertexA.Position)

	s.remainingIterations = 0 // ensure we can't be asked to iterate further
	s.solution = EPASolution{
		VertexA: vertexA,
		VertexB: vertexB,
		Normal:  normal,
		Lerp:    max(0.0, min(dot/sqrLength, 1.0)),
		Depth:   s.skinRadius + distance,
	}
}

// terminatePoint completes the solve with a single vertex as the closest
// feature of the Minkowski difference to the origin.
func (s *EPASolver) terminatePoint(shape *MinkowskiShape, vertex MinkowskiVertex) {
	// Handle degenerate cases.
	var normal dprec.Vec2
	if vertex.Position.SqrLength() < 1e-12 {
		normal = shape.VertexNormal(vertex)
	} else {
		normal = dprec.InverseVec2(dprec.UnitVec2(vertex.Position))
	}

	s.remainingIterations = 0 // ensure we can't be asked to iterate further
	s.solution = EPASolution{
		VertexA: vertex,
		VertexB: vertex,
		Normal:  normal,
		Lerp:    0.0,
		Depth:   s.skinRadius + dprec.Vec2Dot(normal, vertex.Position),
	}
}

func (s *EPASolver) addPolytopeEdge(vertexA, vertexB MinkowskiVertex) {
	pointA := vertexA.Position
	pointB := vertexB.Position

	edge := dprec.Vec2Diff(pointB, pointA)
	sqrLength := edge.SqrLength()
	if sqrLength < 1e-12 {
		return // ignore degenerate edges
	}
	lerp := -dprec.Vec2Dot(edge, pointA) / sqrLength

	var distance float64
	switch {
	case lerp <= 0.0:
		distance = pointA.Length()
	case lerp >= 1.0:
		distance = pointB.Length()
	default:
		projection := dprec.Vec2Lerp(pointA, pointB, lerp)
		distance = projection.Length()
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
	// TODO: godoc
	VertexA MinkowskiVertex
	// TODO: godoc
	VertexB MinkowskiVertex
	// Normal is the unit direction along which the source shape must be
	// moved by Depth to separate the shapes. In world space it points from
	// the target shape toward the source shape. When the origin is outside
	// the Minkowski difference core it points from the closest feature
	// toward the origin; when the origin is contained it is the outward
	// normal of the closest boundary feature.
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
