// Package gjk3d implements the Gilbert-Johnson-Keerthi (GJK) algorithm for
// intersection detection between 3D convex shapes with optional skin radius
// support.
//
// A [Shape] is a convex polyhedron that can be inflated by a skin radius,
// which allows spheres, capsules, and rounded boxes to be represented
// exactly. Use a [Solver] to test two shapes for overlap through
// [Solver.Intersect].
package gjk3d
