package gjk2d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d/internal"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// Intersect reports whether shapeA and shapeB overlap, accounting for their
// skin radii. It returns false immediately if either shape has no points.
func Intersect(shapeA, shapeB Shape) bool {
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

	dir := dprec.BasisXVec2()
	support := minkowskiSupport(&polyA, &polyB, offset, dir)
	simplex.Append(support, dir)

	for simplex.CanProgress() {
		dir = simplex.SearchDirection()
		support = minkowskiSupport(&polyA, &polyB, offset, dir)
		simplex.Append(support, dir)
	}

	return simplex.OverlapsOrigin()
}

// Resolve reports whether shapeA and shapeB overlap and, if so, returns a
// Contact describing how to separate them. The Contact is expressed relative to
// shapeB as the target shape (see shape2d.Contact). The boolean result is false
// when the shapes do not overlap, in which case the Contact is meaningless.
func Resolve(shapeA, shapeB Shape) (shape2d.Contact, bool) {
	// TODO: Add proper implementation using EPA algorithm.
	return shape2d.Contact{
		Depth:        5.0,
		TargetPoint:  dprec.NewVec2(200.0, 200.0),
		TargetNormal: dprec.BasisXVec2(),
	}, Intersect(shapeA, shapeB)
}

func minkowskiSupport(polyA, polyB *internal.Polygon, offset, dir dprec.Vec2) dprec.Vec2 {
	supportA := polyA.Support(dprec.InverseVec2(dir))
	supportB := polyB.Support(dir)
	return dprec.Vec2Sum(offset, dprec.Vec2Diff(supportB, supportA))
}
