// Package gjk2d implements the Gilbert-Johnson-Keerthi (GJK) algorithm for
// intersection detection and contact resolution between 2D convex shapes
// with optional skin radius support.
//
// A [Shape] is a convex polygon that can be inflated by a skin radius,
// which allows circles, capsules, and rounded polygons to be represented
// exactly. Use a [Solver] to test two shapes for overlap through
// [Solver.Intersect] or to compute the contact needed to separate them
// through [Solver.Resolve].
package gjk2d
