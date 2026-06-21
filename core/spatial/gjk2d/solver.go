package gjk2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d/internal"
)

type Solver struct{}

func NewSolver() *Solver {
	return &Solver{}
}

func (s *Solver) Intersect(shapeA, shapeB Shape) bool {
	if len(shapeA.Points) == 0 || len(shapeB.Points) == 0 {
		return false
	}

	simplex := internal.NewSimplex(
		len(shapeA.Points)+len(shapeB.Points),
		shapeA.SkinRadius+shapeB.SkinRadius,
	)

	offset := dprec.Vec2Diff(shapeB.Position, shapeA.Position)
	polyA := internal.Polygon{
		Rotation:    shapeA.Rotation,
		InvRotation: shapeA.Rotation.Inverse(),
		Points:      shapeA.Points,
	}
	polyB := internal.Polygon{
		Rotation:    shapeB.Rotation,
		InvRotation: shapeB.Rotation.Inverse(),
		Points:      shapeB.Points,
	}

	dir := s.pickInitialDirection(&polyA, &polyB, offset)
	support := s.minkowskiSupport(&polyA, &polyB, offset, dir)
	simplex.Append(support, dir)

	for simplex.CanProgress() {
		dir = simplex.SearchDirection()
		support = s.minkowskiSupport(&polyA, &polyB, offset, dir)
		simplex.Append(support, dir)
	}

	return simplex.TouchesOrigin()
}

func (s *Solver) pickInitialDirection(polyA, polyB *internal.Polygon, offset dprec.Vec2) dprec.Vec2 {
	pointA := polyA.InitialPoint()
	pointB := polyB.InitialPoint()
	result := dprec.Vec2Sum(offset, dprec.Vec2Diff(pointB, pointA))
	if result.SqrLength() < 0.001 {
		return dprec.BasisXVec2()
	}
	return result
}

func (s *Solver) minkowskiSupport(polyA, polyB *internal.Polygon, offset, dir dprec.Vec2) dprec.Vec2 {
	supportA := polyA.Support(dprec.InverseVec2(dir))
	supportB := polyB.Support(dir)
	return dprec.Vec2Sum(offset, dprec.Vec2Diff(supportB, supportA))
}
