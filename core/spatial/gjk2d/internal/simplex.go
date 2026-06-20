package internal

import "github.com/mokiat/gomath/sprec"

type Simplex struct {
	remainingiterations uint32
	sqrSkinRadius       float32

	points          [2]sprec.Vec2
	pointCount      uint32
	searchDirection sprec.Vec2

	canProgress   bool
	touchesOrigin bool
}

func NewSimplex(maxIterations int, skinRadius float32) Simplex {
	return Simplex{
		remainingiterations: uint32(maxIterations),
		sqrSkinRadius:       skinRadius * skinRadius,

		points:          [2]sprec.Vec2{},
		pointCount:      0,
		searchDirection: sprec.BasisXVec2(),

		canProgress:   true,
		touchesOrigin: false,
	}
}

func (s *Simplex) CanProgress() bool {
	return s.canProgress
}

func (s *Simplex) Append(point, dir sprec.Vec2) {
	if s.remainingiterations == 0 {
		s.terminate(false)
		return
	}
	s.remainingiterations--

	switch s.pointCount {
	case 0:
		s.appendToEmpty(point)
	case 1:
		s.appendToPoint(point, dir)
	case 2:
		s.appendToEdge(point, dir)
	}
}

func (s *Simplex) appendToEmpty(point sprec.Vec2) {
	// If the first point is already within skin-radius distance of the origin,
	// then we can immediately conclude that the simplex touches the origin.
	if point.SqrLength() <= s.sqrSkinRadius {
		s.terminate(true)
		return
	}

	// Append the first point and set the search direction towards the origin.
	s.points[0] = point
	s.pointCount++
	s.searchDirection = sprec.InverseVec2(point)
}

func (s *Simplex) appendToPoint(point, lastDir sprec.Vec2) {
	// Check if the new point is at all applicable.
	if !s.crossedSupportPlane(point, lastDir) {
		s.terminate(false)
		return
	}

	// Check if the new point is within skin-radius distance of the origin, which
	// would mean that the simplex touches the origin.
	if point.SqrLength() <= s.sqrSkinRadius {
		s.terminate(true)
		return
	}

	// Append the new point, making an edge.
	s.points[1] = point
	s.pointCount++

	// Ensure that the edge is oriented towards the origin. Follow CCW winding convention.
	norm := transposeVec2(sprec.Vec2Diff(s.points[1], s.points[0]))
	if dot := sprec.Vec2Dot(norm, point); dot > 0 { // edge normal points away from the origin
		s.points[0], s.points[1] = s.points[1], s.points[0]
		norm = sprec.InverseVec2(norm)
	}

	// Check that the origin is not already within skin-radius distance of the edge, which
	// would mean that the simplex touches the origin.
	if originProjectsToEdge(s.points[0], s.points[1]) {
		if dot := sprec.Vec2Dot(norm, point); dot*dot <= norm.SqrLength()*s.sqrSkinRadius {
			s.terminate(true)
			return
		}
	}

	// Set the search direction to be perpendicular to the edge, towards the origin.
	s.searchDirection = norm
}

func (s *Simplex) appendToEdge(point, lastDir sprec.Vec2) {
	// Check if the new point is at all applicable.
	if !s.crossedSupportPlane(point, lastDir) {
		s.terminate(false)
		return
	}

	// Check if the new point is within skin-radius distance of the origin, which
	// would mean that the simplex touches the origin.
	if point.SqrLength() <= s.sqrSkinRadius {
		s.terminate(true)
		return
	}

	// Figure out which edge to keep and the new search direction. Follow CCW winding convention.
	normAC := transposeVec2(sprec.Vec2Diff(point, s.points[0]))
	normCB := transposeVec2(sprec.Vec2Diff(s.points[1], point))

	dotAC := -sprec.Vec2Dot(normAC, point)
	dotCB := -sprec.Vec2Dot(normCB, point)

	switch {
	case dotAC <= 0.0 && dotCB <= 0.0: // origin is within the triangle
		s.terminate(true)
		return

	case dotAC > 0.0 && dotCB <= 0.0: // origin is past edge AC
		// Check if the edge AC is within skin-radius distance of the origin,
		// which would mean that the simplex touches the origin.
		if originProjectsToEdge(s.points[0], point) {
			if dotAC*dotAC <= normAC.SqrLength()*s.sqrSkinRadius {
				s.terminate(true)
				return
			}
		}
		// Keep edge AC.
		s.points[1] = point
		// Set the search direction to be perpendicular to the edge, towards the origin.
		s.searchDirection = normAC

	case dotAC <= 0.0 && dotCB > 0.0: // origin is past edge CB
		// Check if the edge CB is within skin-radius distance of the origin,
		// which would mean that the simplex touches the origin.
		if originProjectsToEdge(point, s.points[1]) {
			if dotCB*dotCB <= normCB.SqrLength()*s.sqrSkinRadius {
				s.terminate(true)
				return
			}
		}
		// Keep edge CB.
		s.points[0] = point
		// Set the search direction to be perpendicular to the edge, towards the origin.
		s.searchDirection = normCB

	default: // origin is past both edge lines (the vertex C wedge)
		// In case of very shallow angles, the origin may be within skin-radius
		// distance of one of the edges or a point past the edges.
		projectsAC := originProjectsToEdge(s.points[0], point)
		if projectsAC && (dotAC*dotAC <= normAC.SqrLength()*s.sqrSkinRadius) {
			s.terminate(true)
			return
		}
		projectsCB := originProjectsToEdge(point, s.points[1])
		if projectsCB && (dotCB*dotCB <= normCB.SqrLength()*s.sqrSkinRadius) {
			s.terminate(true)
			return
		}
		switch {
		case projectsAC: // keep edge AC and keep searching towards the origin
			s.points[1] = point
			s.searchDirection = normAC
		case projectsCB: // keep edge CB and keep searching towards the origin
			s.points[0] = point
			s.searchDirection = normCB
		default: // the closest feature really is vertex C, already rejected above
			s.terminate(false)
			return
		}
	}
}

func (s *Simplex) TouchesOrigin() bool {
	return s.touchesOrigin
}

func (s *Simplex) SearchDirection() sprec.Vec2 {
	return s.searchDirection
}

// crossedSupportPlane checks if the point is past the plane defined by the
// origin and the skin radius along the inverse of the last search direction.
//
// If the furthers point along the last search direction never even reached
// anywhere past the plane skin-radius distance away from the origin, then the
// origin can never be touched by the simplex.
func (s *Simplex) crossedSupportPlane(point, lastDir sprec.Vec2) bool {
	dot := sprec.Vec2Dot(point, lastDir)
	if dot >= 0 {
		return true // The point is past the plane at the origin so we are good.
	}
	return dot*dot <= lastDir.SqrLength()*s.sqrSkinRadius
}

func (s *Simplex) terminate(success bool) {
	s.touchesOrigin = success
	s.canProgress = false
}

func originProjectsToEdge(start, end sprec.Vec2) bool {
	edge := sprec.Vec2Diff(end, start)
	dot := -sprec.Vec2Dot(edge, start)
	return dot >= 0.0 && dot <= edge.SqrLength()
}

func transposeVec2(v sprec.Vec2) sprec.Vec2 {
	return sprec.NewVec2(v.Y, -v.X)
}
